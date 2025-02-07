// controller/receiptController.go
package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"receipt-processor-challenge/config"
	"receipt-processor-challenge/model"
	"regexp"
	"strings"
)

// GET MethodS
func GetReceipt(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/receipts/"), "/")
	//path := strings.TrimPrefix(r.URL.Path, "/receipts/")
	if path == "" {
		// Return all receipts
		receipts := model.GetAllReceipts()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(receipts)
		return
	}
	// Check if the path ends with "/points" to route to the GetReceiptPoints handler
	if strings.HasSuffix(path, "/points") {
		id := strings.TrimSuffix(path, "/points")
		GetReceiptPoints(w, r, id)
		return
	}
	// Return specific receipt
	receipt, exists := model.GetReceiptById(path)
	if !exists {
		http.Error(w, "Receipt not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(receipt)
}

// POST Method
func ProcessReceipt(w http.ResponseWriter, r *http.Request) {
	var receipt model.Receipt
	receipt.Items = []model.Item{} // Initialize Items to an empty slice

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields() // Reject unknown 
	err := decoder.Decode(&receipt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate receipt before any processing
    if err := receipt.ValidateReceipt(); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

	// run trimming operation for itemShortDescriptions...
	cleanItemShortDescriptions(&receipt)

    // Format date using consolidated utility
    formattedDate, err := config.ValidateAndFormatDate(receipt.PurchaseDate)
    if err != nil {
        http.Error(w, fmt.Sprintf("Date formatting error: %v", err), http.StatusBadRequest)
        return
    }
    receipt.PurchaseDate = formattedDate

    // Format time using consolidated utility
    formattedTime, err := config.ValidateAndFormatTime(receipt.PurchaseTime)
    if err != nil {
        http.Error(w, fmt.Sprintf("Time formatting error: %v", err), http.StatusBadRequest)
        return
    }
    receipt.PurchaseTime = formattedTime

	receipt.GenerateUniqueID()
	receipt.CalculatePoints()

	model.AddReceipt(receipt)
	

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {
		ID string `json:"id"`
	}{
		ID: receipt.ID,
	})
}

// Updated handler to get points for a specific receipt
func GetReceiptPoints(w http.ResponseWriter, r *http.Request, id string) {
	receipt, exists := model.GetReceiptById(id)
	if !exists {
		http.Error(w, "Receipt not found, as such, 0 Points", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {
		Points uint `json:"points"`
	}{
		Points: receipt.Points,
	})
}

// NotFoundHandler handles requests to non-existent endpoints
func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
    // Set the status code to 404
    w.WriteHeader(http.StatusNotFound)
    
    // Write a standard error message
    response := map[string]string{
        "error": "Endpoint not found",
        "message": fmt.Sprintf("The requested URL %s was not found on this server.", r.URL.Path),
    }
    
    // Write the response in JSON format
    if err := json.NewEncoder(w).Encode(response); err != nil {
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
    }
}
/*
	Helper Functions
*/
// Clean item descriptions by trimming and reducing multiple spaces
func cleanItemShortDescriptions(receipt *model.Receipt) {
	for i, item := range receipt.Items {
		// Trim leading and trailing spaces
		description := strings.TrimSpace(item.ShortDescription)
		// Replace multiple spaces with a single space
		re := regexp.MustCompile(`\s+`)
		description = re.ReplaceAllString(description, " ")
		receipt.Items[i].ShortDescription = description
	}
}