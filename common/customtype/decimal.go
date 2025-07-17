package customtype

import (
	"database/sql/driver"
	"fmt"

	"github.com/primadi/lokstra/common/json"

	"github.com/shopspring/decimal"
)

var DefaultScale int32 = 2

type Decimal struct {
	decimal.Decimal
	Scale int32
}

func normalize(d decimal.Decimal, scale int32) decimal.Decimal {
	return d.Round(scale)
}

func NewDecimalFromString(str string) (Decimal, error) {
	d, err := decimal.NewFromString(str)
	if err != nil {
		return Decimal{}, err
	}
	return Decimal{Decimal: normalize(d, DefaultScale), Scale: DefaultScale}, nil
}

func NewDecimalFromFloat(f float64) Decimal {
	d := decimal.NewFromFloat(f)
	return Decimal{Decimal: normalize(d, DefaultScale), Scale: DefaultScale}
}

func (d Decimal) MarshalJSON() ([]byte, error) {
	return []byte(d.Decimal.StringFixed(d.Scale)), nil
}

func (d *Decimal) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		d.Decimal = decimal.Zero
		d.Scale = DefaultScale
		return nil
	}

	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		if str == "" {
			d.Decimal = decimal.Zero
			d.Scale = DefaultScale
			return nil
		}

		parsed, err := decimal.NewFromString(str)
		if err != nil {
			return err
		}
		d.Decimal = normalize(parsed, DefaultScale)
		d.Scale = DefaultScale
		return nil
	}

	var num float64
	if err := json.Unmarshal(data, &num); err != nil {
		return err
	}

	d.Decimal = normalize(decimal.NewFromFloat(num), DefaultScale)
	d.Scale = DefaultScale
	return nil
}

func (d Decimal) Value() (driver.Value, error) {
	return d.StringFixed(d.Scale), nil
}

func (d *Decimal) Scan(value any) error {
	var str string
	switch v := value.(type) {
	case []byte:
		str = string(v)
	case string:
		str = v
	default:
		return fmt.Errorf("cannot scan type %T into Decimal", value)
	}
	parsed, err := decimal.NewFromString(str)
	if err != nil {
		return err
	}
	d.Scale = DefaultScale
	d.Decimal = normalize(parsed, d.Scale)
	return nil
}
