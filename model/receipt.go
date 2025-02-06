// model/receipt.go
package model
import (
	"sync"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"
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

	// Validate purchase date
	if err := validateDate(receipt.PurchaseDate); err != nil {
		return err
	}
	// Validate purchase time
	if err := validateTime(receipt.PurchaseTime); err != nil {
		return err
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
// Helper Functions:

/* 
	Validators:
	- Date
	- Time
*/
func validateDate(dateStr string) error {
	// No need to match if Valid, just convert
	/*
		if matched, _ := regexp.MatchString(`^\d{4}-\d{2}-\d{2}$`, dateStr); !matched {
			return errors.New("error: date format must be YYYY-MM-DD")
		} */
	// Check for empty date
    if dateStr == "" {
        return errors.New("date cannot be empty")
    }

	// Handle different date formats
    parts := []string{}
    if strings.Contains(dateStr, "/") {
        parts = strings.Split(dateStr, "/")
    } else if strings.Contains(dateStr, "-") {
        parts = strings.Split(dateStr, "-")
    }

	if len(parts) == 3 {
        var month, day int
        var err error
        
        // Try to parse the first two components as numbers
        if len(parts[0]) <= 2 { // MM/DD format
            month, err = strconv.Atoi(parts[0])
            if err != nil {
                return fmt.Errorf("invalid month format: %s", parts[0])
            }
            day, err = strconv.Atoi(parts[1])
            if err != nil {
                return fmt.Errorf("invalid day format: %s", parts[1])
            }
        } else { // YYYY-MM format
            month, err = strconv.Atoi(parts[1])
            if err != nil {
                return fmt.Errorf("invalid month format: %s", parts[1])
            }
            day, err = strconv.Atoi(parts[2])
            if err != nil {
                return fmt.Errorf("invalid day format: %s", parts[2])
            }
        }

        // Validate month and day ranges
        if month < 1 || month > 12 {
            return fmt.Errorf("invalid month: %d (must be between 1 and 12)", month)
        }
        if day < 1 || day > 31 {
            return fmt.Errorf("invalid day: %d (must be between 1 and 31)", day)
        }

        // Additional validation for months with less than 31 days
        if day == 31 && (month == 4 || month == 6 || month == 9 || month == 11) {
            return fmt.Errorf("invalid day: month %d has only 30 days", month)
        }
        // Special case for February
        if month == 2 {
            if day > 29 {
                return fmt.Errorf("invalid day: February cannot have more than 29 days")
            }
        }
    } else {
        return errors.New("invalid date format: must be YYYY-MM-DD, MM/DD/YYYY, or DD/MM/YYYY")
    }

    return nil
}

func validateTime(timeStr string) error {
    if timeStr == "" {
        return errors.New("time cannot be empty")
    }

    // Handle different time formats
    var hours, minutes int
    var err error

    // Check if time contains AM/PM
    isAMPM := strings.Contains(strings.ToUpper(timeStr), "AM") || strings.Contains(strings.ToUpper(timeStr), "PM")

    if isAMPM {
        // Parse 12-hour format
        timeStr = strings.ToUpper(timeStr)
        timeStr = strings.TrimSpace(timeStr)
        
        // Remove any seconds if present
        if strings.Count(timeStr, ":") == 2 {
            parts := strings.Split(timeStr, ":")
            timeStr = parts[0] + ":" + parts[1] + strings.Split(parts[2], " ")[1]
        }

        t, err := time.Parse("3:04 PM", timeStr)
        if err != nil {
            return fmt.Errorf("invalid 12-hour time format: %s", timeStr)
        }
        hours = t.Hour()
        minutes = t.Minute()
    } else {
        // Parse 24-hour format
        parts := strings.Split(timeStr, ":")
        if len(parts) < 2 {
            return errors.New("time must be in HH:MM format")
        }

        hours, err = strconv.Atoi(parts[0])
        if err != nil {
            return fmt.Errorf("invalid hours format: %s", parts[0])
        }

        minutes, err = strconv.Atoi(parts[1])
        if err != nil {
            return fmt.Errorf("invalid minutes format: %s", parts[1])
        }
    }

    // Validate hours and minutes
    if hours < 0 || hours > 23 {
        return fmt.Errorf("invalid hours: %d (must be between 0 and 23)", hours)
    }
    if minutes < 0 || minutes > 59 {
        return fmt.Errorf("invalid minutes: %d (must be between 0 and 59)", minutes)
    }

    return nil
}

// Time Check
func isTimeInRange(timeStr string) (bool, error) {
	// Define the layout for time parsing
	layout := "15:04" // 24-hour time format
	// Parse the input time string
	t, err := time.Parse(layout, timeStr)
	if err != nil {
		return false, fmt.Errorf("error parsing time %s: %v", timeStr, err)
	}
	// Define the start and end times of the range
	startTime, _ := time.Parse(layout, "14:00")
	endTime, _ := time.Parse(layout, "16:00")
	
	// return True if time -> after Start AND before End
	inRange := t.After(startTime) && t.Before(endTime)
	return inRange, nil
}


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
func calculatePointsFromPurchaseDate(purchaseDate string) uint {
	points := uint(0)
	layout := "2006-01-02"
	if t, err := time.Parse(layout, purchaseDate); err == nil {
		// if odd
		if t.Day()%2 != 0 {
			points += 6
		}
	} else {
		fmt.Printf("Error parsing date %s: %v\n", purchaseDate, err)
	}
	return points
}
func calculatePointsFromPurchaseTime(purchaseTime string) uint {
	points := uint(0)
	if inRange, err := isTimeInRange(purchaseTime); err == nil {
		if inRange {
			points += 10
		}
	} else {
		fmt.Println(err)
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