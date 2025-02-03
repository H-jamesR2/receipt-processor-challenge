// model/receipt_test.go
package model
import (
	"testing"
)
func TestValidateReceipt(t *testing.T) {
	tests := []struct {
		receipt   Receipt
		isValid bool
	}{
		{Receipt{Retailer: "Test", PurchaseDate: "2022-01-01", PurchaseTime: "12:00", Items: []Item{{ShortDescription: "Test Item", Price: "10"}}, Total: "100"}, true},
		{Receipt{Retailer: "", PurchaseDate: "2022-01-01", PurchaseTime: "12:00", Items: []Item{{ShortDescription: "Test Item", Price: "10"}}, Total: "100"}, false},
		{Receipt{Retailer: "Test", PurchaseDate: "2022-01-01", PurchaseTime: "12:00", Items: []Item{}}, false},
		{Receipt{Retailer: "Test", PurchaseDate: "invalid-date", PurchaseTime: "12:00", Items: []Item{{ShortDescription: "Test Item", Price: "10"}}, Total: "100"}, false},
		{Receipt{Retailer: "Test", PurchaseDate: "2022-01-01", PurchaseTime: "invalid-time", Items: []Item{{ShortDescription: "Test Item", Price: "10"}}, Total: "100"}, false},
	}
	for _, test := range tests {
		err := ValidateReceipt(test.receipt)
		if (err == nil) != test.isValid {
			t.Errorf("ValidateOrder(%v) returned %v, expected %v", test.receipt, err == nil, test.isValid)
		}
	}
}
func TestCalculatePoints(t *testing.T) {
	receipt := Receipt{
		Retailer:     "M&M Corner Market",
		PurchaseDate: "2022-03-20",
		PurchaseTime: "14:33",	
		Items: []Item{
			{ShortDescription: "Gatorade", Price: "2.25"},
			{ShortDescription: "Gatorade", Price: "2.25"},
			{ShortDescription: "Gatorade", Price: "2.25"},
			{ShortDescription: "Gatorade", Price: "2.25"},
		},
		Total:        "9.00",
	}
	CalculatePoints(&receipt)
	if receipt.Points != 109 {
		t.Errorf("CalculatePoints(%v) returned %d points, expected 109", receipt, receipt.Points)
	}
}