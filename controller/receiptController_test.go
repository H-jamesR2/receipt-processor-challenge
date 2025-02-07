package controller

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "receipt-processor-challenge/model"
	"io/ioutil" // Add this for reading response body
    "fmt" // Add this for debugging
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
}

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