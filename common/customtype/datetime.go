package customtype

import (
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/primadi/lokstra/common/json"
)

type DateTime struct {
	time.Time
}

var StandardDateTimeFormat = "2006-01-02 15:04:05" // YYYY-MM-DD HH:mm:SS format

// UnmarshalJSON implements the json.Unmarshaler interface
func (d *DateTime) UnmarshalJSON(data []byte) error {
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
func (d DateTime) MarshalJSON() ([]byte, error) {
	if d.Time.IsZero() {
		return []byte("null"), nil
	}
	return fmt.Appendf(nil, "\"%s\"", d.Time.Format(StandardDateTimeFormat)), nil
}

// String returns the date as a string in YYYY-MM-DD HH:mm:SS format
func (d DateTime) String() string {
	return d.Time.Format(StandardDateTimeFormat)
}

func (d DateTime) FormatYMD(YMDFormat string) string {
	replacer := strings.NewReplacer(
		"YYYY", "2006",
		"YY", "06",
		"MMMM", "January",
		"Mmmm", "January",
		"MMM", "Jan",
		"Mmm", "Jan",
		"MM", "01",
		"DD", "02",
		"HH", "15",
		"hh", "03",
		"mm", "04",
		"ss", "05",
		"AMPM", "PM",
		"A", "PM",
		"ampm", "PM",
		"a", "PM")
	layout := replacer.Replace(YMDFormat)
	return d.Format(layout)
}

func (d DateTime) parse(input string) (time.Time, error) {
	var year, month, day, hour, minute, second int
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
				case 3:
					hour = number
				case 4:
					minute = number
				case 5:
					second = number
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
		case 3:
			hour = number
		case 4:
			minute = number
		case 5:
			second = number
		default:
			return time.Time{}, fmt.Errorf("too many fields in input")
		}
		field++
	}

	if field < 3 {
		return time.Time{}, fmt.Errorf("not enough fields in input")
	}

	return time.Date(year, time.Month(month), day, hour, minute, second, 0, time.UTC), nil
}
