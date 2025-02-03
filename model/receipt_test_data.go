// model/receipt_test_data.go
package model

var ValidateReceiptTestCases = []struct {
    Name     string
    JsonData string
    IsValid  bool
}{
    {
        Name: "Valid Receipt",
        JsonData: `{
            "retailer": "Target",
            "purchaseDate": "2022-01-01",
            "purchaseTime": "13:01",
            "total": "6.49",
            "items": [
                {
                    "shortDescription": "Mountain Dew 12PK",
                    "price": "6.49"
                }
            ]
        }`,
        IsValid: true,
    },
    {   
        Name: "Empty Retailer",
        JsonData: `{
            "retailer": "",
            "purchaseDate": "2022-01-02",
            "purchaseTime": "13:13",
            "total": "14.25",
            "items": [
                {
                    "shortDescription": "Pepsi - 12-oz",
                    "price": "14.25"
                }
            ]
        }`,
        IsValid: false,
    },
    {
        Name: "Invalid Date",
        JsonData: `{
            "retailer": "Walmart",
            "purchaseDate": "13/13/2023",
            "purchaseTime": "13:13",
            "total": "14.25",
            "items": [
                {
                    "shortDescription": "Pepsi - 12-oz",
                    "price": "14.25"
                }
            ]
        }`,
        IsValid: false,
    },
    {
        Name: "Invalid Time",
        JsonData: `{
            "retailer": "Walmart",
            "purchaseDate": "2022-01-02",
            "purchaseTime": "24:01",
            "total": "14.25",
            "items": [
                {
                    "shortDescription": "Pepsi - 12-oz",
                    "price": "14.25"
                }
            ]
        }`,
        IsValid: false,
    },
    {
        Name: "Mismatched Total vs. ItemsTotal",
        JsonData: `{
            "retailer": "Walmart",
            "purchaseDate": "2022-01-02",
            "purchaseTime": "05:00",
            "total": "11.00",
            "items": [
                {
                    "shortDescription": "Pepsi - 12-oz",
                    "price": "6.25"
                }
            ]
        }`,
        IsValid: false,
    },
    {
        Name: "Invalid Total",
        JsonData: `{
            "retailer": "Walmart",
            "purchaseDate": "2022-01-02",
            "purchaseTime": "05:00",
            "total": "InvalidTotal",
            "items": [
                {
                    "shortDescription": "Pepsi - 12-oz",
                    "price": "6.25"
                }
            ]
        }`,
        IsValid: false,
    },
    {
        Name: "Missing Item Price",
        JsonData: `{
            "retailer": "Walmart",
            "purchaseDate": "2022-01-02",
            "purchaseTime": "05:00",
            "total": "10.00",
            "items": [
                {
                    "shortDescription": "Pepsi - 12-oz"
                }
            ]
        }`,
        IsValid: false,
    },
    {
        Name: "Missing Item Description",
        JsonData: `{
            "retailer": "Walmart",
            "purchaseDate": "2022-01-02",
            "purchaseTime": "05:00",
            "total": "10.00",
            "items": [
                {
                    "price": "10.00"
                }
            ]
        }`,
        IsValid: false,
    },
    {
        Name: "Negative Item Price",
        JsonData: `{
            "retailer": "Walmart",
            "purchaseDate": "2022-01-02",
            "purchaseTime": "05:00",
            "total": "6.25",
            "items": [
                {
                    "shortDescription": "Pepsi - 12-oz",
                    "price": "-6.25"
                }
            ]
        }`,
        IsValid: false,
    },
    {
        Name: "Missing Items List",
        JsonData: `{
            "retailer": "Walmart",
            "purchaseDate": "2022-01-02",
            "purchaseTime": "05:00",
            "total": "6.25",
            "items": []
        }`,
        IsValid: false,
    },
}

var CalculatePointsTestCases = []struct {
    Name           string
    JsonData       string
    ExpectedPoints uint
}{
    {
        Name: "Target Receipt",
        JsonData: `{
            "retailer": "Target",
            "purchaseDate": "2022-01-01",
            "purchaseTime": "13:01",
            "items": [
                {
                    "shortDescription": "Mountain Dew 12PK",
                    "price": "6.49"
                },
                {
                    "shortDescription": "Emils Cheese Pizza",
                    "price": "12.25"
                },
                {
                    "shortDescription": "Knorr Creamy Chicken",
                    "price": "1.26"
                },
                {
                    "shortDescription": "Doritos Nacho Cheese",
                    "price": "3.35"
                },
                {
                    "shortDescription": "   Klarbrunn 12-PK 12 FL OZ  ",
                    "price": "12.00"
                }
            ],
            "total": "35.35"
        }`,
        ExpectedPoints: 28,
    },
}