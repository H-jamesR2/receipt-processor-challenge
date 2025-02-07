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
/*
func TestProcessReceipt_ValidInput(t *testing.T) {
    // Create test receipt
    receipt := createTestReceipt()
    
    // Convert receipt to JSON
    body, err := json.Marshal(receipt)
    if err != nil {
        t.Fatalf("Failed to marshal receipt: %v", err)
    }
    
	// Debug: Print the request body
    fmt.Printf("Request body: %s\n", string(body))

    // Create request
    req := httptest.NewRequest("POST", "/receipts/process", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")
    
    // Create response recorder
    rr := httptest.NewRecorder()
    
    // Call the handler
    ProcessReceipt(rr, req)
    
    // Debug: Print the response
    respBody, _ := ioutil.ReadAll(rr.Body)
    fmt.Printf("Response status: %d\n", rr.Code)
    fmt.Printf("Response body: %s\n", string(respBody))

    // Check status code
    if status := rr.Code; status != http.StatusOK {
        t.Errorf("handler returned wrong status code: got %v want %v\nResponse body: %s", 
            status, http.StatusOK, string(respBody))
    }

	// Reset the response body for further reading
    rr.Body = bytes.NewBuffer(respBody)
    
    // Check response structure
    var response struct {
        ID string `json:"id"`
    }
    err = json.Unmarshal(respBody, &response)
    if err != nil {
        t.Errorf("Failed to parse response: %v\nResponse body: %s", err, string(respBody))
    }
    if response.ID == "" {
        t.Error("Expected non-empty ID in response")
    }
}

func TestProcessReceipt_InvalidInput(t *testing.T) {
    // Test cases
    testCases := []struct {
        name     string
        receipt  model.Receipt
        expected int
    }{
        {
            name: "Missing Retailer",
            receipt: model.Receipt{
                PurchaseDate: "2024-02-07",
                PurchaseTime: "13:45",
                Total: "35.35",
            },
            expected: http.StatusBadRequest,
        },
        {
            name: "Invalid Date Format",
            receipt: model.Receipt{
                Retailer: "Target",
                PurchaseDate: "2024/02/07", // Wrong format
                PurchaseTime: "13:45",
                Total: "35.35",
            },
            expected: http.StatusBadRequest,
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            body, _ := json.Marshal(tc.receipt)
            req := httptest.NewRequest("POST", "/receipts/process", bytes.NewBuffer(body))
            req.Header.Set("Content-Type", "application/json")
            rr := httptest.NewRecorder()
            
            ProcessReceipt(rr, req)
            
            if status := rr.Code; status != tc.expected {
                t.Errorf("%s: handler returned wrong status code: got %v want %v", 
                    tc.name, status, tc.expected)
            }
        })
    }
} */
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

/*
func TestGetReceipt(t *testing.T) {
    // First create a receipt
    receipt := createTestReceipt()
    receipt.GenerateUniqueID()
    model.AddReceipt(receipt)
    
    // Test getting the receipt
    req := httptest.NewRequest("GET", "/receipts/"+receipt.ID, nil)
    rr := httptest.NewRecorder()
    
    GetReceipt(rr, req)
    
    // Check status code
    if status := rr.Code; status != http.StatusOK {
        t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
    }
    
    // Verify response content
    var response model.Receipt
    err := json.Unmarshal(rr.Body.Bytes(), &response)
    if err != nil {
        t.Errorf("Failed to parse response: %v", err)
    }
    if response.ID != receipt.ID {
        t.Errorf("Expected receipt ID %v, got %v", receipt.ID, response.ID)
    }
} */
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

/*
func TestParseAndFormatDate(t *testing.T) {
    testData := GetDateTestData() // Only get date test data

    for _, tc := range testData.Valid {
        t.Run(tc.Name, func(t *testing.T) {
            result, err := parseAndFormatDate(tc.Input)
            if err != nil {
                t.Errorf("Expected no error for valid date %s, got error: %v", tc.Input, err)
            }
            if result != tc.Expected {
                t.Errorf("Expected formatted date %s, got %s", tc.Expected, result)
            }
        })
    }

    for _, tc := range testData.Invalid {
        t.Run(tc.Name, func(t *testing.T) {
            _, err := parseAndFormatDate(tc.Input)
            if err == nil {
                t.Errorf("Expected error for invalid date %s, got none", tc.Input)
            }
        })
    }
}
func TestParseAndFormatTime(t *testing.T) {
    testData := GetTimeTestData() // Only get time test data

    for _, tc := range testData.Valid {
        t.Run(tc.Name, func(t *testing.T) {
            result, err := parseAndFormatTime(tc.Input)
            if err != nil {
                t.Errorf("Expected no error for valid time %s, got error: %v", tc.Input, err)
            }
            if result != tc.Expected {
                t.Errorf("Expected formatted time %s, got %s", tc.Expected, result)
            }
        })
    }

    for _, tc := range testData.Invalid {
        t.Run(tc.Name, func(t *testing.T) {
            _, err := parseAndFormatTime(tc.Input)
            if err == nil {
                t.Errorf("Expected error for invalid time %s, got none", tc.Input)
            }
        })
    }
}
*/

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