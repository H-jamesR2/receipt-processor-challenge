// model/receipt.go
package model
import (
	"sync"
	"errors"
	"fmt"
	"math"
	"strconv"
	"unicode"
	"github.com/google/uuid"
	"receipt-processor-challenge/config"

	"strings"
	"regexp"
)
type Item struct {
	ShortDescription string `json:"shortDescription"`
	Price            string `json:"price"`
}
type Receipt struct {
	Retailer     string `json:"retailer"`
	PurchaseDate string `json:"purchaseDate"`
	PurchaseTime string `json:"purchaseTime"`
	Items        []Item `json:"items"`
	Total        string `json:"total"`
	ID           string `json:"id"`
	Points       uint   `json:"points"`
}
var (
	receipts    = make(map[string]Receipt)
	receiptsMux sync.Mutex
)
func (r *Receipt) GenerateUniqueID() {
	r.ID = uuid.New().String()
}

func (receipt *Receipt) ValidateReceipt() error {
	standardErrorPrefix := "error processing receipt:\n   "
	if receipt.Retailer == "" {
		return errors.New(standardErrorPrefix + "retailer cannot be empty")
	}

    // Use consolidated date validation
    if _, err := config.ValidateAndFormatDate(receipt.PurchaseDate); err != nil {
        return fmt.Errorf("%s%v", standardErrorPrefix, err)
    }

    // Use consolidated time validation
    if _, err := config.ValidateAndFormatTime(receipt.PurchaseTime); err != nil {
        return fmt.Errorf("%s%v", standardErrorPrefix, err)
    }

	if len(receipt.Items) == 0 {
		return errors.New(standardErrorPrefix + "items cannot be empty")
	}

	testItemsTotal := float64(0)
	for _, item := range receipt.Items {
		if item.ShortDescription == "" {
			return errors.New(standardErrorPrefix + "item description cannot be empty")
		}
		if item.Price == "" {
			return errors.New(standardErrorPrefix + "item price cannot be empty")
		}
		
		// price less than or equal to 0 or an error...
		price, priceErr := strconv.ParseFloat(item.Price, 64)
		if price <= 0 {
			return errors.New(standardErrorPrefix + "item price must be greater than zero")
		} else if priceErr != nil  {
			return priceErr
		} else {
			// add to testTotal to verify and check
			testItemsTotal += price
		}
	}

	// run check on items before cleaning...

	// convert to float + round...
	receiptTotal, receiptErr := strconv.ParseFloat(receipt.Total, 64)
	receiptTotal, testItemsTotal = config.RoundToNearestCent(receiptTotal), config.RoundToNearestCent(testItemsTotal)
	
	if receiptErr != nil {
		return errors.New(standardErrorPrefix + "error on total price")
	} else if receiptTotal != testItemsTotal {
		return errors.New(standardErrorPrefix + "item calculatedTotal does not match Total price")
	}

	return nil
}


func (receipt *Receipt) CalculatePoints() {
	// Points Calculation
	points := uint(0)

	// add 1 pt for every alphaNumeric char in retailer name..
	points += calculatePointsFromRetailerAlphaNumChar(receipt.Retailer)

	// If the total is a multiple of 0.25, add 25 pts.
	points += calculatePointsFromTotal(receipt.Total)

	// add 5 points for every TWO items in the receipt.
	// 3/2 -> 1 (discards .5)
	points += calculatePointsForEveryTwoItems(receipt.Items)
	//(uint(((len(receipt.Items) / 2) * 5)))

	// go through items w/ pre-trimmed descriptions.
	points += calculatePointsFromItemPriceAndDesc(receipt.Items)

	/*
		Processing date + time.
	*/
	// Parse the purchaseDate and check if the day is odd or even.
	points += calculatePointsFromPurchaseDate(receipt.PurchaseDate)

	// Parse the purchaseTime and check if between 
	// after startTime && before endTime.
	points += calculatePointsFromPurchaseTime(receipt.PurchaseTime)

	receipt.Points = points
}

func AddReceipt(receipt Receipt) error {
	receiptsMux.Lock()
	receipts[receipt.ID] = receipt
	receiptsMux.Unlock()

	return nil
}
func GetReceiptById(id string) (Receipt, bool) {
	receiptsMux.Lock()
	receipt, exists := receipts[id]
	receiptsMux.Unlock()
	return receipt, exists
}
func GetAllReceipts() []Receipt {
	receiptsMux.Lock()
	defer receiptsMux.Unlock()
	receiptsList := make([]Receipt, 0, len(receipts))
	for _, receipt := range receipts {
		receiptsList = append(receiptsList, receipt)
	}
	return receiptsList
}

func ClearReceipts() {
    receipts = make(map[string]Receipt)
}

// Helper Functions:
/* 
	Calculation Functions:
*/
func calculatePointsFromRetailerAlphaNumChar(retailer string) uint {
	points := uint(0)
	for _, c := range retailer {
		if unicode.IsLetter(c) || unicode.IsDigit(c) {
			points++
		}
	}
	return points
}
func calculatePointsFromTotal(total string) uint {
	points := uint(0)
	if totalFloat, err := strconv.ParseFloat(total, 64); err == nil && math.Mod(totalFloat, 0.25) == 0 {
		points += 25
		// Get the decimal part of the float value
		// totalDecimal := totalFloat - float64(int(totalFloat))
		// If decimal part of total == .00, add 50 pts.
		if math.Mod(totalFloat*100, 100) == 0 {
			points += 50
		}
	} else if err != nil {
		fmt.Printf("Error parsing total: %v\n", err)
	}
	return points
}

func calculatePointsForEveryTwoItems(items []Item) uint {
	return uint(((len(items) / 2) * 5))
}

func calculatePointsFromItemPriceAndDesc(items []Item) uint {
	points := uint(0)
	for _, item := range items {
		if len(item.ShortDescription)%3 == 0 {
			itemPrice, err := strconv.ParseFloat(item.Price, 64)
			if err != nil {
				fmt.Printf("Error parsing itemPrice: %v\n", err)
				continue
			} else {
				/*	> multiply itemPrice by 0.2
					> round to nearest integer
					> convert to unsigned int
					> add to points (uint)
				*/
				points += (uint(math.Ceil((itemPrice * 0.2))))
			}
		}
	}
	return points
}
// modified
func calculatePointsFromPurchaseDate(purchaseDate string) uint {
    points := uint(0)
    // Date is already validated and in ISO format
    parts := strings.Split(purchaseDate, "-")
    day, _ := strconv.Atoi(parts[2])
    if day%2 != 0 {
        points += 6
    }
    return points
}

func calculatePointsFromPurchaseTime(purchaseTime string) uint {
    points := uint(0)
    inRange, err := config.IsTimeInRange(purchaseTime, "14:00", "16:00")
    if err == nil && inRange {
        points += 10
    }
    return points
}

func (r *Receipt) CleanItemShortDescriptions() {
    for i, item := range r.Items {
        // Trim leading and trailing spaces
        description := strings.TrimSpace(item.ShortDescription)

        // Replace multiple spaces with a single space
        re := regexp.MustCompile(`\s+`)
        description = re.ReplaceAllString(description, " ")

        r.Items[i].ShortDescription = description
    }
}