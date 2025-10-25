package dto

import (
	"encoding/json"
	"strings"
	"time"
)

// FlexibleTime is a custom time type that can parse multiple time formats
type FlexibleTime struct {
	time.Time
}

// UnmarshalJSON implements json.Unmarshaler interface
func (ft *FlexibleTime) UnmarshalJSON(data []byte) error {
	// Remove quotes from the JSON string
	timeStr := strings.Trim(string(data), `"`)

	if timeStr == "" || timeStr == "null" {
		return nil
	}

	// Try different time formats
	formats := []string{
		time.RFC3339,           // 2006-01-02T15:04:05Z07:00
		time.RFC3339Nano,       // 2006-01-02T15:04:05.999999999Z07:00
		"2006-01-02T15:04:05Z", // 2006-01-02T15:04:05Z
		"2006-01-02T15:04:05",  // 2006-01-02T15:04:05
		"2006-01-02T15:04",     // 2006-01-02T15:04 (missing seconds)
		"2006-01-02 15:04:05",  // 2006-01-02 15:04:05
		"2006-01-02 15:04",     // 2006-01-02 15:04
		"2006-01-02",           // 2006-01-02
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			ft.Time = t
			return nil
		}
	}

	// If no format matches, try parsing with time.Parse which is more flexible
	if t, err := time.Parse(time.RFC3339, timeStr); err == nil {
		ft.Time = t
		return nil
	}

	// If still no match, try adding default seconds and timezone
	if !strings.Contains(timeStr, ":") {
		// If no time part, assume 00:00:00
		timeStr += "T00:00:00Z"
	} else if !strings.Contains(timeStr, "Z") && !strings.Contains(timeStr, "+") && !strings.Contains(timeStr, "-") {
		// If no timezone, add Z
		if !strings.Contains(timeStr, ":") {
			timeStr += ":00:00Z"
		} else if strings.Count(timeStr, ":") == 1 {
			timeStr += ":00Z"
		} else {
			timeStr += "Z"
		}
	}

	if t, err := time.Parse(time.RFC3339, timeStr); err == nil {
		ft.Time = t
		return nil
	}

	return &time.ParseError{
		Layout:     "multiple formats",
		Value:      timeStr,
		LayoutElem: "time",
		ValueElem:  timeStr,
		Message:    "unable to parse time",
	}
}

// MarshalJSON implements json.Marshaler interface
func (ft FlexibleTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(ft.Time.Format(time.RFC3339))
}

// IsZero returns true if the time is zero value
func (ft FlexibleTime) IsZero() bool {
	return ft.Time.IsZero()
}

// String returns the string representation of the time
func (ft FlexibleTime) String() string {
	return ft.Time.String()
}
