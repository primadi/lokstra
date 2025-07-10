package customtype

import (
	"fmt"
	"lokstra/common/json"
	"strings"
	"time"
	"unicode"
)

type Date struct {
	time.Time
}

const StandardDateFormat = "2006-01-02" // YYYY-MM-DD format

// UnmarshalJSON implements the json.Unmarshaler interface
func (d *Date) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		d.Time = time.Time{}
		return nil
	}

	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	if str == "" {
		d.Time = time.Time{}
		return nil
	}
	t, err := d.parse(str)
	if err != nil {
		return err
	}
	d.Time = t
	return nil
}

// MarshalJSON implements the json.Marshaler interface
func (d Date) MarshalJSON() ([]byte, error) {
	if d.Time.IsZero() {
		return []byte("null"), nil
	}
	return fmt.Appendf(nil, "\"%s\"", d.Time.Format(StandardDateFormat)), nil
}

// String returns the date as a string in YYYY-MM-DD format
func (d Date) String() string {
	return d.Time.Format(StandardDateFormat)
}

func (d Date) parse(input string) (time.Time, error) {
	var year, month, day int
	field := 0
	number := 0
	readingNumber := false

	for _, ch := range input {
		if unicode.IsDigit(ch) {
			number = number*10 + int(ch-'0')
			readingNumber = true
		} else {
			if readingNumber {
				switch field {
				case 0:
					year = number
				case 1:
					month = number
				case 2:
					day = number
				default:
					return time.Time{}, fmt.Errorf("too many fields in input")
				}
				field++
				number = 0
				readingNumber = false
			}
			// ignore separator '-', ' ', ':'
		}
	}

	// process the last number if we were reading one
	if readingNumber {
		switch field {
		case 0:
			year = number
		case 1:
			month = number
		case 2:
			day = number
		default:
			return time.Time{}, fmt.Errorf("too many fields in input")
		}
		field++
	}

	if field < 3 {
		return time.Time{}, fmt.Errorf("not enough fields in input")
	}

	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC), nil
}

func (d Date) FormatYMD(YMDFormat string) string {
	replacer := strings.NewReplacer(
		"YYYY", "2006",
		"YY", "06",
		"MMMM", "January",
		"Mmmm", "January",
		"MMM", "Jan",
		"Mmm", "Jan",
		"MM", "01",
		"DD", "02")
	layout := replacer.Replace(YMDFormat)
	return d.Format(layout)
}
