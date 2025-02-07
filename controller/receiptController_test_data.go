// controller/receiptController_test_data.go
package controller

// Struct definitions remain the same
type DateTestData struct {
    Valid   []DateTestCase
    Invalid []DateTestCase
}

type TimeTestData struct {
    Valid   []TimeTestCase
    Invalid []TimeTestCase
}

type ReceiptTestData struct {
    Valid   []ReceiptTestCase
    Invalid []ReceiptTestCase
}

type DateTestCase struct {
    Name        string
    Input       string
    Expected    string
    ShouldError bool
}

type TimeTestCase struct {
    Name        string
    Input       string
    Expected    string
    ShouldError bool
}

type ReceiptTestCase struct {
    Name        string
    Input       string
    StatusCode  int
    Response    string
}

// Separate functions for different test data sets
func GetDateTestData() DateTestData {
    return DateTestData{
        Valid: []DateTestCase{
            {
                Name:        "ISO Format",
                Input:       "2024-02-07",
                Expected:    "2024-02-07",
                ShouldError: false,
            },
            {
                Name:        "MM/DD/YYYY",
                Input:       "02/07/2024",
                Expected:    "2024-02-07",
                ShouldError: false,
            },
            {
                Name:        "DD/MM/YYYY",
                Input:       "07/02/2024",
                Expected:    "2024-02-07",
                ShouldError: false,
            },
            {
                Name:        "YYYY/MM/DD",
                Input:       "2024/02/07",
                Expected:    "2024-02-07",
                ShouldError: false,
            },
            {
                Name:        "Month D, YYYY",
                Input:       "Feb 7, 2024",
                Expected:    "2024-02-07",
                ShouldError: false,
            },
        },
        Invalid: []DateTestCase{
            {
                Name:        "Invalid Month",
                Input:       "2024-13-01",
                ShouldError: true,
            },
            {
                Name:        "Invalid Day",
                Input:       "2024-02-30",
                ShouldError: true,
            },
            {
                Name:        "Invalid Format",
                Input:       "invalid-date",
                ShouldError: true,
            },
            {
                Name:        "Invalid Day for Month",
                Input:       "2024-04-31",
                ShouldError: true,
            },
        },
    }
}

func GetTimeTestData() TimeTestData {
    return TimeTestData{
        Valid: []TimeTestCase{
            {
                Name:        "24 Hour Format",
                Input:       "13:45",
                Expected:    "13:45",
                ShouldError: false,
            },
            {
                Name:        "12 Hour Format PM",
                Input:       "1:45 PM",
                Expected:    "13:45",
                ShouldError: false,
            },
            {
                Name:        "12 Hour Format AM",
                Input:       "9:45 AM",
                Expected:    "09:45",
                ShouldError: false,
            },
            {
                Name:        "With Seconds",
                Input:       "13:45:00",
                Expected:    "13:45",
                ShouldError: false,
            },
        },
        Invalid: []TimeTestCase{
            {
                Name:        "Invalid Hour",
                Input:       "25:00",
                ShouldError: true,
            },
            {
                Name:        "Invalid Minute",
                Input:       "12:60",
                ShouldError: true,
            },
            {
                Name:        "Invalid Format",
                Input:       "invalid-time",
                ShouldError: true,
            },
        },
    }
}

func GetReceiptTestData() ReceiptTestData {
    return ReceiptTestData{
        Valid: []ReceiptTestCase{
            {
                Name: "Valid Receipt",
                Input: `{
                    "retailer": "Target",
                    "purchaseDate": "2024-02-07",
                    "purchaseTime": "13:45",
                    "items": [
                        {
                            "shortDescription": "Mountain Dew",
                            "price": "1.99"
                        }
                    ],
                    "total": "1.99"
                }`,
                StatusCode: 200,
                Response:   `{"id":""}`,
            },
        },
        Invalid: []ReceiptTestCase{
            {
                Name: "Missing Required Fields",
                Input: `{
                    "retailer": "Target",
                    "purchaseDate": "2024-02-07"
                }`,
                StatusCode: 400,
                Response:   "Missing required fields",
            },
            {
                Name: "Invalid JSON",
                Input: `{invalid json}`,
                StatusCode: 400,
                Response:   "Invalid JSON format",
            },
        },
    }
}