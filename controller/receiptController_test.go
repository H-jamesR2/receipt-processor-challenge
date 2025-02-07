package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"receipt-processor-challenge/model"
	"strings"
	"testing"
)

// Test Helpers
// Enhanced test helper with more complete data
func createTestReceipt() model.Receipt {
    return model.Receipt{
        Retailer: "Target",
        PurchaseDate: "2024-02-07",
        PurchaseTime: "13:45",
        Items: []model.Item{
            {
                ShortDescription: "Mountain Dew",
                Price: "1.99",
            },
        },
        Total: "1.99", // Make sure total matches items
    }
}

// Unit Tests
func TestProcessReceipt(t *testing.T) {
    testData := GetReceiptTestData() // Only get receipt test data

    for _, tc := range testData.Valid {
        t.Run(tc.Name, func(t *testing.T) {
            req := httptest.NewRequest("POST", "/receipts/process", bytes.NewBufferString(tc.Input))
            req.Header.Set("Content-Type", "application/json")
            rr := httptest.NewRecorder()

            ProcessReceipt(rr, req)

            if rr.Code != tc.StatusCode {
                t.Errorf("Expected status code %d, got %d", tc.StatusCode, rr.Code)
            }
        })
    }

    for _, tc := range testData.Invalid {
        t.Run(tc.Name, func(t *testing.T) {
            req := httptest.NewRequest("POST", "/receipts/process", bytes.NewBufferString(tc.Input))
            req.Header.Set("Content-Type", "application/json")
            rr := httptest.NewRecorder()

            ProcessReceipt(rr, req)

            if rr.Code != tc.StatusCode {
                t.Errorf("Expected status code %d, got %d", tc.StatusCode, rr.Code)
            }
        })
    }
}

 func TestGetReceipt_Comprehensive(t *testing.T) {
    testData := GetGetterReceiptTestData()

    // Test valid cases
    for _, tc := range testData.Valid {
        t.Run(tc.Name, func(t *testing.T) {
            // Setup
            model.ClearReceipts()
            var receiptID string
            if tc.SetupReceipt {
                receipt := createTestReceipt()
                receipt.GenerateUniqueID()
                receipt.CalculatePoints()
                model.AddReceipt(receipt)
                receiptID = receipt.ID
            }

            // Format path if needed
            path := tc.Path
            if strings.Contains(path, "%s") {
                path = fmt.Sprintf(tc.Path, receiptID)
            }

            // Create and execute request
            req := httptest.NewRequest("GET", path, nil)
            rr := httptest.NewRecorder()

            GetReceipt(rr, req)

            // Check status code
            if status := rr.Code; status != tc.ExpectedStatus {
                t.Errorf("handler returned wrong status code: got %v want %v",
                    status, tc.ExpectedStatus)
            }

            // Validate response based on the type of request
            switch {
            case strings.HasSuffix(path, "/points"):
                var response struct {
                    Points uint `json:"points"`
                }
                if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
                    t.Fatalf("Failed to decode points response: %v", err)
                }
            case strings.HasSuffix(path, "/"):
                var receipts []model.Receipt
                if err := json.NewDecoder(rr.Body).Decode(&receipts); err != nil {
                    t.Fatalf("Failed to decode receipts list: %v", err)
                }
                if tc.SetupReceipt && len(receipts) != 1 {
                    t.Errorf("Expected 1 receipt, got %d", len(receipts))
                }
            default:
                var receipt model.Receipt
                if err := json.NewDecoder(rr.Body).Decode(&receipt); err != nil {
                    t.Fatalf("Failed to decode receipt: %v", err)
                }
                if receipt.ID != receiptID {
                    t.Errorf("Expected receipt ID %s, got %s", receiptID, receipt.ID)
                }
            }
        })
    }

    // Test invalid cases
    for _, tc := range testData.Invalid {
        t.Run(tc.Name, func(t *testing.T) {
            // Setup
            model.ClearReceipts()
            
            // Create and execute request
            req := httptest.NewRequest("GET", tc.Path, nil)
            rr := httptest.NewRecorder()

            GetReceipt(rr, req)

            // Check status code
            if status := rr.Code; status != tc.ExpectedStatus {
                t.Errorf("handler returned wrong status code: got %v want %v",
                    status, tc.ExpectedStatus)
            }

            // Check error message
            if !strings.Contains(rr.Body.String(), tc.ExpectedBody) {
                t.Errorf("Expected response to contain '%s', got '%s'", 
                    tc.ExpectedBody, rr.Body.String())
            }
        })
    }
}

func TestGetReceiptPoints(t *testing.T) {
    // Create and store a receipt
    receipt := createTestReceipt()
    receipt.GenerateUniqueID()
    receipt.CalculatePoints()
    model.AddReceipt(receipt)
    
    // Test getting points
    req := httptest.NewRequest("GET", "/receipts/"+receipt.ID+"/points", nil)
    rr := httptest.NewRecorder()
    
    GetReceiptPoints(rr, req, receipt.ID)
    
    // Check status code
    if status := rr.Code; status != http.StatusOK {
        t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
    }
    
    // Verify response
    var response struct {
        Points uint `json:"points"`
    }
    err := json.Unmarshal(rr.Body.Bytes(), &response)
    if err != nil {
        t.Errorf("Failed to parse response: %v", err)
    }
}

// Add this test to verify receipt validation
func TestReceiptValidation(t *testing.T) {
    receipt := createTestReceipt()
    err := receipt.ValidateReceipt()
    if err != nil {
        t.Errorf("Receipt validation failed: %v", err)
    }
}


func TestNotFoundHandler(t *testing.T) {
    testData := GetNotFoundTestData()

    for _, tc := range testData.Cases {
        t.Run(tc.Name, func(t *testing.T) {
            req := httptest.NewRequest("GET", tc.Path, nil)
            rr := httptest.NewRecorder()

            NotFoundHandler(rr, req)

            if status := rr.Code; status != tc.ExpectedStatus {
                t.Errorf("handler returned wrong status code: got %v want %v",
                    status, tc.ExpectedStatus)
            }

            var response map[string]string
            err := json.NewDecoder(rr.Body).Decode(&response)
            if err != nil {
                t.Fatalf("Failed to decode response body: %v", err)
            }

            if response["error"] != tc.ExpectedError {
                t.Errorf("handler returned unexpected error: got %v want %v",
                    response["error"], tc.ExpectedError)
            }

            expectedMessage := fmt.Sprintf("The requested URL %s was not found on this server.", tc.Path)
            if response["message"] != expectedMessage {
                t.Errorf("handler returned unexpected message: got %v want %v",
                    response["message"], expectedMessage)
            }
        })
    }
}