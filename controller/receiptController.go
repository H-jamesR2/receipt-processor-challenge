// controller/receiptController.go
package controller
import (
	"encoding/json"
	"fmt"
	"net/http"
	"receipt-processor-challenge/model"
	"regexp"
	"strings"
	"time"
	"strconv"
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

	// run trimming operation for itemShortDescriptions...
	cleanItemShortDescriptions(&receipt)

	    // Validate receipt before any processing
    if err := receipt.ValidateReceipt(); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Format date only if validation passed
    formattedDate, err := parseAndFormatDate(receipt.PurchaseDate)
    if err != nil {
        http.Error(w, fmt.Sprintf("Date formatting error: %v", err), http.StatusBadRequest)
        return
    }
    receipt.PurchaseDate = formattedDate
	
	// reformat Time if needed.
	if formattedTime, err := parseAndFormatTime(receipt.PurchaseTime); err == nil {
		receipt.PurchaseTime = formattedTime
	} else {
		fmt.Println(err)
	}

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
// Date Functions
func isISODateFormat(dateStr string) bool {
	// Regular expression to check if the date string is in YYYY-MM-DD format
	re := regexp.MustCompile(`^([0-9]{4})-(0[1-9]|1[0-2])-(0[1-9]|[12][0-9]|3[01])$`)
	return re.MatchString(dateStr)
}

func parseAndFormatDate(dateStr string) (string, error) {
    if isISODateFormat(dateStr) {
        // Even if it's ISO format, validate the components
        parts := strings.Split(dateStr, "-")
        month, _ := strconv.Atoi(parts[1])
        day, _ := strconv.Atoi(parts[2])
        if month < 1 || month > 12 || day < 1 || day > 31 {
            return "", fmt.Errorf("invalid date components in ISO format: month=%d, day=%d", month, day)
        }
        return dateStr, nil
    }

    formats := []string{
        "01/02/2006",   // MM/DD/YYYY
        "02/01/2006",   // DD/MM/YYYY
        "2006/01/02",   // YYYY/MM/DD
        "Jan 2, 2006",  // Jan 2, 2006
        "02-Jan-2006",  // 02-Jan-2006
    }

    var parsedDate time.Time
    var err error
    dateLayout := "2006-01-02"

    for _, format := range formats {
        if parsedDate, err = time.Parse(format, dateStr); err == nil {
            // Verify the parsed date components are valid
            month := int(parsedDate.Month())
            day := parsedDate.Day()
            
            if month < 1 || month > 12 || day < 1 || day > 31 {
                continue // Try next format
            }
            
            // Additional validation for specific months
            if day == 31 && (month == 4 || month == 6 || month == 9 || month == 11) {
                continue // This format produced an invalid date
            }
            if month == 2 && day > 29 {
                continue // Invalid February date
            }
            
            return parsedDate.Format(dateLayout), nil
        }
    }

    return "", fmt.Errorf("unable to parse date: %s", dateStr)
}

// Time Functions
func is24HourFormat(timeStr string) bool {
	// Regular expression to check if the time string is in HH:MM format
	re := regexp.MustCompile(`^([01][0-9]|2[0-3]):[0-5][0-9]$`)
	return re.MatchString(timeStr)
}
func parseAndFormatTime(timeStr string) (string, error) {
	// If the time string is already in 24-hour format, return it directly
	if is24HourFormat(timeStr) {
		return timeStr, nil
	}
	// Define possible time formats
	formats := []string{
		"15:04",       // 24-hour clock with minutes
		"15:04:05",    // 24-hour clock with seconds
		"03:04 PM",    // 12-hour clock with AM/PM
		"03:04:05 PM", // 12-hour clock with seconds and AM/PM
	}
	var parsedTime time.Time
	var err error
	// Try parsing with each format
	for _, format := range formats {
		if parsedTime, err = time.Parse(format, timeStr); err == nil {
			break
		}
	}
	// parsed timeStr not a valid timeString.
	if err != nil {
		return "", fmt.Errorf("error parsing time %s: %v", timeStr, err)
	}
	// Format time to 24-hour clock format
	return parsedTime.Format("15:04"), nil
}

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