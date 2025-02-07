// config/datetime.go
package config

import (
    "fmt"
    "regexp"
    "strconv"
    "strings"
    "time"
)

// Date-related utilities
type DateFormat struct {
    Layout string
    Pattern *regexp.Regexp
    Description string
}

var DateFormats = []DateFormat{
    {
        Layout: "2006-01-02", 
        Pattern: regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`),
        Description: "ISO format (YYYY-MM-DD)",
    },
    {
        Layout: "01/02/2006", 
        Pattern: regexp.MustCompile(`^(0[1-9]|1[0-2])/(0[1-9]|[12]\d|3[01])/\d{4}$`),
        Description: "US format (MM/DD/YYYY)",
    },
    {
        Layout: "02/01/2006", 
        Pattern: regexp.MustCompile(`^(0[1-9]|[12]\d|3[01])/(0[1-9]|1[0-2])/\d{4}$`),
        Description: "UK format (DD/MM/YYYY)",
    },
    {
        Layout: "2006/01/02",
        Pattern: regexp.MustCompile(`^\d{4}/(0[1-9]|1[0-2])/(0[1-9]|[12]\d|3[01])$`),
        Description: "ISO with slashes (YYYY/MM/DD)",
    },
    {
        Layout: "Jan 2, 2006",
        Pattern: regexp.MustCompile(`^(Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)\s+\d{1,2},\s+\d{4}$`),
        Description: "Month DD, YYYY",
    },
    {
        Layout: "2 Jan 2006",
        Pattern: regexp.MustCompile(`^\d{1,2}\s+(Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)\s+\d{4}$`),
        Description: "DD Month YYYY",
    },
}

func ValidateAndFormatDate(dateStr string) (string, error) {
    if dateStr == "" {
        return "", fmt.Errorf("date cannot be empty")
    }

    // Clean input
    dateStr = strings.TrimSpace(dateStr)

    // Already in ISO format
    if DateFormats[0].Pattern.MatchString(dateStr) {
        if isValidDate(dateStr) {
            return dateStr, nil
        }
        return "", fmt.Errorf("invalid date components: %s", dateStr)
    }

    // Try parsing with other formats
    for _, format := range DateFormats {
        if format.Pattern.MatchString(dateStr) {
            parsedDate, err := time.Parse(format.Layout, dateStr)
            if err == nil {
                isoDate := parsedDate.Format("2006-01-02")
                if isValidDate(isoDate) {
                    return isoDate, nil
                }
            }
        }
    }

    return "", fmt.Errorf("unable to parse date: %s", dateStr)
}

func isValidDate(dateStr string) bool {
    parts := strings.Split(dateStr, "-")
    if len(parts) != 3 {
        return false
    }

    year, err1 := strconv.Atoi(parts[0])
    month, err2 := strconv.Atoi(parts[1])
    day, err3 := strconv.Atoi(parts[2])

    if err1 != nil || err2 != nil || err3 != nil {
        return false
    }

    // Validate year (reasonable range check)
    if year < 1900 || year > 2100 {
        return false
    }

    // Basic range checks
    if month < 1 || month > 12 || day < 1 || day > 31 {
        return false
    }

    // Check days in month with leap year consideration
    daysInMonth := getDaysInMonth(year, month)
    return day <= daysInMonth
}

func getDaysInMonth(year, month int) int {
    // Array of days in each month (non-leap year)
    daysInMonth := []int{0, 31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
    
    // Special handling for February in leap years
    if month == 2 && isLeapYear(year) {
        return 29
    }
    
    return daysInMonth[month]
}

func isLeapYear(year int) bool {
    // Leap year calculation according to the Gregorian calendar:
    // 1. Year must be divisible by 4
    // 2. If year is divisible by 100, it must also be divisible by 400
    return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

// Time-related utilities
type TimeFormat struct {
    Layout string
    Pattern *regexp.Regexp
    Description string
}

var TimeFormats = []TimeFormat{
    {
        Layout: "15:04",
        Pattern: regexp.MustCompile(`^([01][0-9]|2[0-3]):([0-5][0-9])$`),
        Description: "24-hour format (HH:MM)",
    },
    {
        Layout: "3:04 PM",
        Pattern: regexp.MustCompile(`^(1[0-2]|[1-9]):[0-5][0-9]\s*(AM|PM)$`),
        Description: "12-hour format without leading zero",
    },
    {
        Layout: "03:04 PM",
        Pattern: regexp.MustCompile(`^(0[1-9]|1[0-2]):[0-5][0-9]\s*(AM|PM)$`),
        Description: "12-hour format with leading zero",
    },
    {
        Layout: "15:04:05",
        Pattern: regexp.MustCompile(`^([01][0-9]|2[0-3]):([0-5][0-9]):([0-5][0-9])$`),
        Description: "24-hour format with seconds",
    },
    {
        Layout: "3:04:05 PM",
        Pattern: regexp.MustCompile(`^(1[0-2]|[1-9]):[0-5][0-9]:[0-5][0-9]\s*(AM|PM)$`),
        Description: "12-hour format with seconds",
    },
}

func ValidateAndFormatTime(timeStr string) (string, error) {
    if timeStr == "" {
        return "", fmt.Errorf("time cannot be empty")
    }

    // Clean input
    timeStr = strings.TrimSpace(timeStr)
    timeStr = strings.ToUpper(timeStr)

    // Remove seconds if present
    if strings.Count(timeStr, ":") == 2 {
        parts := strings.Split(timeStr, ":")
        timeStr = parts[0] + ":" + parts[1]
        if strings.Contains(timeStr, "AM") || strings.Contains(timeStr, "PM") {
            timeStr += strings.Split(parts[2], " ")[1]
        }
    }

    // Already in 24-hour format
    if TimeFormats[0].Pattern.MatchString(timeStr) {
        if isValidTime(timeStr) {
            return timeStr, nil
        }
        return "", fmt.Errorf("invalid time components: %s", timeStr)
    }

    // Try parsing with other formats
    for _, format := range TimeFormats {
        if format.Pattern != nil && format.Pattern.MatchString(timeStr) {
            parsedTime, err := time.Parse(format.Layout, timeStr)
            if err == nil {
                return parsedTime.Format("15:04"), nil
            }
        }
    }

    return "", fmt.Errorf("unable to parse time: %s", timeStr)
}

func isValidTime(timeStr string) bool {
    parts := strings.Split(timeStr, ":")
    if len(parts) != 2 {
        return false
    }

    hours, err1 := strconv.Atoi(parts[0])
    minutes, err2 := strconv.Atoi(parts[1])

    if err1 != nil || err2 != nil {
        return false
    }

    return hours >= 0 && hours <= 23 && minutes >= 0 && minutes <= 59
}

func IsTimeInRange(timeStr, startTime, endTime string) (bool, error) {
    t, err := time.Parse("15:04", timeStr)
    if err != nil {
        return false, fmt.Errorf("error parsing time %s: %v", timeStr, err)
    }

    start, _ := time.Parse("15:04", startTime)
    end, _ := time.Parse("15:04", endTime)

    // Handle cases where the range spans midnight
    if end.Before(start) {
        return !t.Before(start) || !t.After(end), nil
    }

    return t.After(start) && t.Before(end), nil
}