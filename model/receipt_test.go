package model

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/google/uuid"
)

func TestValidateReceipt(t *testing.T) {
	for _, testCase := range ValidateReceiptTestCases {
		t.Run(testCase.Name, func(t *testing.T) {
			var receipt Receipt
			err := json.Unmarshal([]byte(testCase.JsonData), &receipt)
			if err != nil {
				t.Fatalf("Failed to unmarshal JSON: %v", err)
			}

			err = receipt.ValidateReceipt()
			isValid := err == nil

			if isValid != testCase.IsValid {
				t.Errorf("Test case %s failed. Expected isValid to be %v, got %v", testCase.Name, testCase.IsValid, isValid)
			}
		})
	}
}

func TestGenerateID(t *testing.T) {
	receipt := Receipt{}
	receipt.GenerateUniqueID()

	zeroUUID := uuid.New().String()
	if receipt.ID == zeroUUID {
		t.Error("expected a non-zero UUID, got zero UUID")
	}
}

func TestCalculatePoints(t *testing.T) {
	for _, testCase := range CalculatePointsTestCases {
		t.Run(testCase.Name, func(t *testing.T) {
			var receipt Receipt
			err := json.Unmarshal([]byte(testCase.JsonData), &receipt)
			if err != nil {
				t.Fatalf("Failed to unmarshal JSON: %v", err)
			}

			receipt.CleanItemShortDescriptions()
			receipt.CalculatePoints()

			if receipt.Points != testCase.ExpectedPoints {
				t.Errorf("Test case %s failed. Expected %d points, got %d", testCase.Name, testCase.ExpectedPoints, receipt.Points)
			}
		})
	}
}

func TestCalculatePointsEdgeCases(t *testing.T) {
    t.Run("Invalid total format", func(t *testing.T) {
        points := calculatePointsFromTotal("invalid")
        if points != 0 {
            t.Errorf("Expected 0 points for invalid total, got %d", points)
        }
    })
    
    t.Run("Invalid item price", func(t *testing.T) {
        items := []Item{
            {ShortDescription: "abc", Price: "invalid"},
        }
        points := calculatePointsFromItemPriceAndDesc(items)
        if points != 0 {
            t.Errorf("Expected 0 points for invalid price, got %d", points)
        }
    })
}

func TestCalculatePointsFromPurchaseTime(t *testing.T) {
    tests := []struct {
        name     string
        time     string
        expected uint
    }{
        {"Time within range", "15:00", 10},
        {"Time outside range", "13:00", 0},
        {"Invalid time", "25:00", 0},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            points := calculatePointsFromPurchaseTime(tt.time)
            if points != tt.expected {
                t.Errorf("Expected %d points, got %d", tt.expected, points)
            }
        })
    }
}

func TestCleanItemShortDescription(t *testing.T) {
	testCases := []struct {
		name     string
		input    []Item
		expected []Item
	}{
		{
			name: "Clean descriptions",
			input: []Item{
				{ShortDescription: "   Klarbrunn 12-PK 12 FL OZ  "},
				{ShortDescription: "Mountain Dew 12PK"},
			},
			expected: []Item{
				{ShortDescription: "Klarbrunn 12-PK 12 FL OZ"},
				{ShortDescription: "Mountain Dew 12PK"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			receipt := Receipt{Items: tc.input}
			receipt.CleanItemShortDescriptions()

			if !reflect.DeepEqual(receipt.Items, tc.expected) {
				t.Errorf("Expected %v, got %v", tc.expected, receipt.Items)
			}
		})
	}
}

func TestCalculatePointsFromRetailerAlphaNumChar(t *testing.T) {
	testCases := []struct {
		retailer string
		expected uint
	}{
		{"Target", 6},
		{"123", 3},
		{"Walmart!", 7},
		{"M&M Corner Market", 14},
		{"", 0},
	}

	for _, tc := range testCases {
		result := calculatePointsFromRetailerAlphaNumChar(tc.retailer)
		if result != tc.expected {
			t.Errorf("For retailer %s, expected %d, got %d", tc.retailer, tc.expected, result)
		}
	}
}

func TestCalculatePointsFromTotal(t *testing.T) {
	testCases := []struct {
		total    string
		expected uint
	}{
		{"10.00", 75},  // Round dollar amount
		{"10.25", 25},  // Quarter
		{"10.50", 25},  // Quarter
		{"10.75", 25},  // Quarter
		{"10.99", 0},   // Not a round dollar or quarter
	}

	for _, tc := range testCases {
		result := calculatePointsFromTotal(tc.total)
		if result != tc.expected {
			t.Errorf("For total %s, expected %d, got %d", tc.total, tc.expected, result)
		}
	}
}

/*
	Test Endpoint Methods: 
	GET, 
	POST
*/


func TestReceiptManagement(t *testing.T) {
    // Clear the receipts map before testing
    receipts = make(map[string]Receipt)

    // Create test receipts
    receipt1 := Receipt{
        ID:       "test-id-1",
        Retailer: "Store 1",
        Items:    []Item{{ShortDescription: "Item 1", Price: "10.00"}},
    }
    receipt2 := Receipt{
        ID:       "test-id-2",
        Retailer: "Store 2",
        Items:    []Item{{ShortDescription: "Item 2", Price: "20.00"}},
    }

    // Test AddReceipt
    t.Run("AddReceipt", func(t *testing.T) {
        err := AddReceipt(receipt1)
        if err != nil {
            t.Errorf("AddReceipt() error = %v", err)
        }
        err = AddReceipt(receipt2)
        if err != nil {
            t.Errorf("AddReceipt() error = %v", err)
        }
    })

    // Test GetReceiptById
    t.Run("GetReceiptById", func(t *testing.T) {
        got, exists := GetReceiptById("test-id-1")
        if !exists {
            t.Error("GetReceiptById() exists = false, want true")
        }
        if !reflect.DeepEqual(got, receipt1) {
            t.Errorf("GetReceiptById() got = %v, want %v", got, receipt1)
        }

        // Test non-existent receipt
        _, exists = GetReceiptById("non-existent")
        if exists {
            t.Error("GetReceiptById() exists = true, want false")
        }
    })

    // Test GetAllReceipts
    t.Run("GetAllReceipts", func(t *testing.T) {
        got := GetAllReceipts()
        if len(got) != 2 {
            t.Errorf("GetAllReceipts() returned %d receipts, want 2", len(got))
        }
        // Check if both receipts are present
        found1, found2 := false, false
        for _, r := range got {
            if r.ID == receipt1.ID {
                found1 = true
            }
            if r.ID == receipt2.ID {
                found2 = true
            }
        }
        if !found1 || !found2 {
            t.Error("GetAllReceipts() missing expected receipts")
        }
    })
}