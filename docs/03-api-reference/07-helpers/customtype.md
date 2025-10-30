# Custom Type Package

The `customtype` package provides custom types for common data formats including dates, datetimes, and high-precision decimals with proper JSON marshaling/unmarshaling.

## Table of Contents

- [Overview](#overview)
- [DateTime Type](#datetime-type)
- [Date Type](#date-type)
- [Decimal Type](#decimal-type)
- [JSON Marshaling](#json-marshaling)
- [Database Integration](#database-integration)
- [Best Practices](#best-practices)
- [Examples](#examples)

## Overview

**Import Path:** `github.com/primadi/lokstra/common/customtype`

**Key Features:**

```
✓ DateTime Type          - ISO-like datetime with flexible parsing
✓ Date Type              - Date-only type (YYYY-MM-DD)
✓ Decimal Type           - High-precision decimal numbers
✓ JSON Integration       - Automatic marshaling/unmarshaling
✓ Database Support       - SQL driver integration (Decimal)
✓ Flexible Parsing       - Multiple input formats accepted
✓ Custom Formatting      - FormatYMD for custom date formats
```

## DateTime Type

### Basic Usage

```go
import "github.com/primadi/lokstra/common/customtype"

// Create DateTime
dt := customtype.DateTime{Time: time.Now()}

// Standard format: "YYYY-MM-DD HH:mm:SS"
fmt.Println(dt.String())  // "2024-03-15 14:30:45"

// Access underlying time.Time
year := dt.Year()
month := dt.Month()
day := dt.Day()
hour := dt.Hour()
```

### JSON Marshaling

```go
type Event struct {
    Name      string                `json:"name"`
    StartTime customtype.DateTime   `json:"start_time"`
    EndTime   customtype.DateTime   `json:"end_time"`
}

// Marshal to JSON
event := Event{
    Name:      "Conference",
    StartTime: customtype.DateTime{Time: time.Date(2024, 3, 15, 9, 0, 0, 0, time.UTC)},
    EndTime:   customtype.DateTime{Time: time.Date(2024, 3, 15, 17, 0, 0, 0, time.UTC)},
}

jsonData, _ := json.Marshal(event)
// {
//   "name": "Conference",
//   "start_time": "2024-03-15 09:00:00",
//   "end_time": "2024-03-15 17:00:00"
// }
```

### JSON Unmarshaling

Accepts flexible input formats:

```go
// Format 1: "YYYY-MM-DD HH:mm:SS"
jsonData := []byte(`{"created_at": "2024-03-15 14:30:45"}`)

// Format 2: "YYYY-MM-DD HH:mm"
jsonData := []byte(`{"created_at": "2024-03-15 14:30"}`)

// Format 3: "YYYY-MM-DD"
jsonData := []byte(`{"created_at": "2024-03-15"}`)

// Format 4: Compact "YYYYMMDDHHmmSS"
jsonData := []byte(`{"created_at": "20240315143045"}`)

type Record struct {
    CreatedAt customtype.DateTime `json:"created_at"`
}

var record Record
json.Unmarshal(jsonData, &record)
// All formats parse successfully
```

**Null Handling:**

```go
// Null value
jsonData := []byte(`{"created_at": null}`)
var record Record
json.Unmarshal(jsonData, &record)
// record.CreatedAt.IsZero() == true

// Empty string
jsonData := []byte(`{"created_at": ""}`)
json.Unmarshal(jsonData, &record)
// record.CreatedAt.IsZero() == true
```

### Custom Formatting

```go
dt := customtype.DateTime{Time: time.Date(2024, 3, 15, 14, 30, 45, 0, time.UTC)}

// FormatYMD with custom format strings
fmt.Println(dt.FormatYMD("YYYY-MM-DD"))           // "2024-03-15"
fmt.Println(dt.FormatYMD("DD/MM/YYYY"))           // "15/03/2024"
fmt.Println(dt.FormatYMD("MM/DD/YYYY HH:mm"))     // "03/15/2024 14:30"
fmt.Println(dt.FormatYMD("YYYY-MM-DD HH:mm:ss"))  // "2024-03-15 14:30:45"

// Month names
fmt.Println(dt.FormatYMD("DD MMMM YYYY"))         // "15 March 2024"
fmt.Println(dt.FormatYMD("MMM DD, YYYY"))         // "Mar 15, 2024"

// 12-hour format with AM/PM
fmt.Println(dt.FormatYMD("hh:mm AMPM"))           // "02:30 PM"
fmt.Println(dt.FormatYMD("hh:mm a"))              // "02:30 PM"
```

**Format Tokens:**

| Token | Meaning | Example |
|-------|---------|---------|
| `YYYY` | 4-digit year | 2024 |
| `YY` | 2-digit year | 24 |
| `MMMM` | Full month name | January |
| `MMM` | Short month name | Jan |
| `MM` | 2-digit month | 03 |
| `DD` | 2-digit day | 15 |
| `HH` | 24-hour format | 14 |
| `hh` | 12-hour format | 02 |
| `mm` | Minutes | 30 |
| `ss` | Seconds | 45 |
| `AMPM` | AM/PM indicator | PM |

## Date Type

### Basic Usage

```go
// Create Date
date := customtype.Date{Time: time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)}

// Standard format: "YYYY-MM-DD"
fmt.Println(date.String())  // "2024-03-15"

// Access components
year := date.Year()
month := date.Month()
day := date.Day()
```

### JSON Marshaling

```go
type Person struct {
    Name      string              `json:"name"`
    BirthDate customtype.Date     `json:"birth_date"`
}

person := Person{
    Name:      "Alice",
    BirthDate: customtype.Date{Time: time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC)},
}

jsonData, _ := json.Marshal(person)
// {
//   "name": "Alice",
//   "birth_date": "1990-05-15"
// }
```

### JSON Unmarshaling

```go
// Format 1: "YYYY-MM-DD"
jsonData := []byte(`{"birth_date": "1990-05-15"}`)

// Format 2: "YYYY/MM/DD"
jsonData := []byte(`{"birth_date": "1990/05/15"}`)

// Format 3: Compact "YYYYMMDD"
jsonData := []byte(`{"birth_date": "19900515"}`)

type Person struct {
    BirthDate customtype.Date `json:"birth_date"`
}

var person Person
json.Unmarshal(jsonData, &person)
// All formats parse successfully
```

**Null Handling:**

```go
// Null value
jsonData := []byte(`{"birth_date": null}`)
var person Person
json.Unmarshal(jsonData, &person)
// person.BirthDate.IsZero() == true
```

### Custom Formatting

```go
date := customtype.Date{Time: time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)}

fmt.Println(date.FormatYMD("YYYY-MM-DD"))        // "2024-03-15"
fmt.Println(date.FormatYMD("DD/MM/YYYY"))        // "15/03/2024"
fmt.Println(date.FormatYMD("MM/DD/YYYY"))        // "03/15/2024"
fmt.Println(date.FormatYMD("DD MMMM YYYY"))      // "15 March 2024"
fmt.Println(date.FormatYMD("MMMM DD, YYYY"))     // "March 15, 2024"
fmt.Println(date.FormatYMD("MMM DD, YY"))        // "Mar 15, 24"
```

### Date Operations

```go
date := customtype.Date{Time: time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)}

// Add/subtract days
tomorrow := customtype.Date{Time: date.AddDate(0, 0, 1)}
yesterday := customtype.Date{Time: date.AddDate(0, 0, -1)}

// Compare dates
if date1.Before(date2.Time) {
    fmt.Println("date1 is earlier")
}

// Calculate difference
duration := date2.Sub(date1.Time)
days := int(duration.Hours() / 24)
```

## Decimal Type

High-precision decimal numbers using shopspring/decimal library.

### Basic Usage

```go
import "github.com/primadi/lokstra/common/customtype"

// From string
price, err := customtype.NewDecimalFromString("19.99")
if err != nil {
    // Handle parsing error
}

// From float
amount := customtype.NewDecimalFromFloat(123.456)

// String representation (fixed scale)
fmt.Println(price.StringFixed(2))  // "19.99"
```

### Default Scale

```go
// Default scale is 2 decimal places
customtype.DefaultScale = 2

// Change default scale
customtype.DefaultScale = 4

dec, _ := customtype.NewDecimalFromString("123.456789")
// Rounded to 4 decimal places: 123.4568
```

### JSON Marshaling

```go
type Product struct {
    Name  string              `json:"name"`
    Price customtype.Decimal  `json:"price"`
}

product := Product{
    Name:  "Book",
    Price: customtype.NewDecimalFromFloat(19.99),
}

jsonData, _ := json.Marshal(product)
// {
//   "name": "Book",
//   "price": 19.99
// }
```

### JSON Unmarshaling

Accepts both string and number formats:

```go
// Format 1: String
jsonData := []byte(`{"price": "19.99"}`)

// Format 2: Number
jsonData := []byte(`{"price": 19.99}`)

// Format 3: Null
jsonData := []byte(`{"price": null}`)

type Product struct {
    Price customtype.Decimal `json:"price"`
}

var product Product
json.Unmarshal(jsonData, &product)
```

### Arithmetic Operations

```go
price := customtype.NewDecimalFromFloat(19.99)
quantity := customtype.NewDecimalFromFloat(3)

// Multiply
total := customtype.Decimal{
    Decimal: price.Mul(quantity.Decimal),
    Scale:   customtype.DefaultScale,
}
// total = 59.97

// Add
discount := customtype.NewDecimalFromFloat(5.00)
finalPrice := customtype.Decimal{
    Decimal: total.Sub(discount.Decimal),
    Scale:   customtype.DefaultScale,
}
// finalPrice = 54.97

// Divide
unitPrice := customtype.Decimal{
    Decimal: total.Div(quantity.Decimal),
    Scale:   customtype.DefaultScale,
}
// unitPrice = 19.99

// Compare
if price.GreaterThan(discount.Decimal) {
    fmt.Println("Price is greater than discount")
}
```

### Database Integration

Decimal implements `sql.Scanner` and `driver.Valuer`:

```go
type Product struct {
    ID    int                `db:"id"`
    Name  string             `db:"name"`
    Price customtype.Decimal `db:"price"`
}

// Query from database
var product Product
err := db.Get(&product, "SELECT id, name, price FROM products WHERE id = ?", 1)

// Insert into database
_, err := db.Exec(
    "INSERT INTO products (name, price) VALUES (?, ?)",
    product.Name,
    product.Price,
)
```

**SQL Value Conversion:**

```go
// Value() - Convert to SQL value
price := customtype.NewDecimalFromFloat(19.99)
sqlValue, _ := price.Value()  // "19.99" (string)

// Scan() - Read from SQL result
var price customtype.Decimal
price.Scan("19.99")            // From string
price.Scan([]byte("19.99"))    // From bytes
```

## JSON Marshaling

### Complete Example

```go
type Invoice struct {
    InvoiceNumber string              `json:"invoice_number"`
    IssueDate     customtype.Date     `json:"issue_date"`
    DueDate       customtype.Date     `json:"due_date"`
    CreatedAt     customtype.DateTime `json:"created_at"`
    Items         []InvoiceItem       `json:"items"`
    Total         customtype.Decimal  `json:"total"`
}

type InvoiceItem struct {
    Description string              `json:"description"`
    Quantity    customtype.Decimal  `json:"quantity"`
    UnitPrice   customtype.Decimal  `json:"unit_price"`
    Amount      customtype.Decimal  `json:"amount"`
}

// Create invoice
invoice := Invoice{
    InvoiceNumber: "INV-2024-001",
    IssueDate:     customtype.Date{Time: time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)},
    DueDate:       customtype.Date{Time: time.Date(2024, 4, 15, 0, 0, 0, 0, time.UTC)},
    CreatedAt:     customtype.DateTime{Time: time.Now()},
    Items: []InvoiceItem{
        {
            Description: "Widget A",
            Quantity:    customtype.NewDecimalFromFloat(10),
            UnitPrice:   customtype.NewDecimalFromFloat(19.99),
            Amount:      customtype.NewDecimalFromFloat(199.90),
        },
    },
    Total: customtype.NewDecimalFromFloat(199.90),
}

// Marshal to JSON
jsonData, _ := json.Marshal(invoice)
```

**Result:**

```json
{
  "invoice_number": "INV-2024-001",
  "issue_date": "2024-03-15",
  "due_date": "2024-04-15",
  "created_at": "2024-03-15 14:30:45",
  "items": [
    {
      "description": "Widget A",
      "quantity": 10.00,
      "unit_price": 19.99,
      "amount": 199.90
    }
  ],
  "total": 199.90
}
```

## Database Integration

### Table Definition

```sql
CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    order_number VARCHAR(50) NOT NULL,
    order_date DATE NOT NULL,
    created_at TIMESTAMP NOT NULL,
    total_amount DECIMAL(10, 2) NOT NULL,
    discount_amount DECIMAL(10, 2) DEFAULT 0,
    final_amount DECIMAL(10, 2) NOT NULL
);
```

### Go Struct

```go
type Order struct {
    ID             int                 `db:"id"`
    OrderNumber    string              `db:"order_number"`
    OrderDate      customtype.Date     `db:"order_date"`
    CreatedAt      customtype.DateTime `db:"created_at"`
    TotalAmount    customtype.Decimal  `db:"total_amount"`
    DiscountAmount customtype.Decimal  `db:"discount_amount"`
    FinalAmount    customtype.Decimal  `db:"final_amount"`
}
```

### CRUD Operations

```go
// Create
order := Order{
    OrderNumber:    "ORD-2024-001",
    OrderDate:      customtype.Date{Time: time.Now()},
    CreatedAt:      customtype.DateTime{Time: time.Now()},
    TotalAmount:    customtype.NewDecimalFromFloat(199.90),
    DiscountAmount: customtype.NewDecimalFromFloat(10.00),
    FinalAmount:    customtype.NewDecimalFromFloat(189.90),
}

_, err := db.Exec(`
    INSERT INTO orders (order_number, order_date, created_at, total_amount, discount_amount, final_amount)
    VALUES (?, ?, ?, ?, ?, ?)
`, order.OrderNumber, order.OrderDate, order.CreatedAt, order.TotalAmount, order.DiscountAmount, order.FinalAmount)

// Read
var order Order
err := db.Get(&order, "SELECT * FROM orders WHERE id = ?", 1)

// Update
_, err := db.Exec(`
    UPDATE orders 
    SET total_amount = ?, discount_amount = ?, final_amount = ?
    WHERE id = ?
`, order.TotalAmount, order.DiscountAmount, order.FinalAmount, order.ID)
```

## Best Practices

### Date/DateTime Usage

```go
✓ DO: Use Date for date-only fields
type Person struct {
    BirthDate customtype.Date `json:"birth_date"`  // No time component needed
}

✗ DON'T: Use DateTime when you don't need time
type Person struct {
    BirthDate customtype.DateTime `json:"birth_date"`  // BAD: Unnecessary time component
}

✓ DO: Use DateTime for timestamps
type Record struct {
    CreatedAt customtype.DateTime `json:"created_at"`  // Full timestamp
    UpdatedAt customtype.DateTime `json:"updated_at"`
}

✓ DO: Check for zero values
if record.CreatedAt.IsZero() {
    // Handle unset datetime
}
```

### Decimal Usage

```go
✓ DO: Use Decimal for monetary values
type Product struct {
    Price customtype.Decimal `json:"price" db:"price"`
}

✗ DON'T: Use float64 for money
type Product struct {
    Price float64 `json:"price"`  // BAD: Precision issues
}

✓ DO: Set appropriate default scale
func init() {
    customtype.DefaultScale = 2  // For currency (2 decimal places)
}

✓ DO: Normalize after arithmetic operations
result := customtype.Decimal{
    Decimal: price.Mul(quantity.Decimal),
    Scale:   customtype.DefaultScale,
}
```

### JSON Parsing

```go
✓ DO: Handle parsing errors
var record Record
if err := json.Unmarshal(jsonData, &record); err != nil {
    return fmt.Errorf("failed to parse JSON: %w", err)
}

✓ DO: Check for zero values after unmarshaling
if record.CreatedAt.IsZero() {
    record.CreatedAt = customtype.DateTime{Time: time.Now()}
}

✓ DO: Use consistent date formats in API
// Always use ISO 8601 formats
// Date: "YYYY-MM-DD"
// DateTime: "YYYY-MM-DD HH:mm:SS"
```

### Database Integration

```go
✓ DO: Use appropriate SQL types
// PostgreSQL
price DECIMAL(10, 2)
order_date DATE
created_at TIMESTAMP

✓ DO: Handle NULL values
type Product struct {
    Price *customtype.Decimal `db:"price"`  // Pointer for nullable
}

✓ DO: Set default scale before database operations
customtype.DefaultScale = 2
```

## Examples

### E-commerce Product

```go
type Product struct {
    ID          int                 `json:"id" db:"id"`
    SKU         string              `json:"sku" db:"sku"`
    Name        string              `json:"name" db:"name"`
    Price       customtype.Decimal  `json:"price" db:"price"`
    CostPrice   customtype.Decimal  `json:"cost_price" db:"cost_price"`
    Stock       int                 `json:"stock" db:"stock"`
    CreatedAt   customtype.DateTime `json:"created_at" db:"created_at"`
    UpdatedAt   customtype.DateTime `json:"updated_at" db:"updated_at"`
}

func CreateProduct(name string, price, costPrice float64) *Product {
    return &Product{
        SKU:       generateSKU(),
        Name:      name,
        Price:     customtype.NewDecimalFromFloat(price),
        CostPrice: customtype.NewDecimalFromFloat(costPrice),
        Stock:     0,
        CreatedAt: customtype.DateTime{Time: time.Now()},
        UpdatedAt: customtype.DateTime{Time: time.Now()},
    }
}

func (p *Product) CalculateMargin() customtype.Decimal {
    margin := p.Price.Sub(p.CostPrice.Decimal)
    return customtype.Decimal{
        Decimal: margin,
        Scale:   customtype.DefaultScale,
    }
}
```

### Event Scheduling

```go
type Event struct {
    ID          int                 `json:"id"`
    Title       string              `json:"title"`
    EventDate   customtype.Date     `json:"event_date"`
    StartTime   customtype.DateTime `json:"start_time"`
    EndTime     customtype.DateTime `json:"end_time"`
    CreatedAt   customtype.DateTime `json:"created_at"`
}

func CreateEvent(title string, eventDate time.Time, startHour, endHour int) *Event {
    return &Event{
        Title:     title,
        EventDate: customtype.Date{Time: eventDate},
        StartTime: customtype.DateTime{
            Time: time.Date(eventDate.Year(), eventDate.Month(), eventDate.Day(),
                startHour, 0, 0, 0, time.UTC),
        },
        EndTime: customtype.DateTime{
            Time: time.Date(eventDate.Year(), eventDate.Month(), eventDate.Day(),
                endHour, 0, 0, 0, time.UTC),
        },
        CreatedAt: customtype.DateTime{Time: time.Now()},
    }
}
```

### Financial Calculation

```go
type Transaction struct {
    ID          int                 `json:"id" db:"id"`
    Type        string              `json:"type" db:"type"`  // debit/credit
    Amount      customtype.Decimal  `json:"amount" db:"amount"`
    Fee         customtype.Decimal  `json:"fee" db:"fee"`
    NetAmount   customtype.Decimal  `json:"net_amount" db:"net_amount"`
    Date        customtype.Date     `json:"date" db:"date"`
    ProcessedAt customtype.DateTime `json:"processed_at" db:"processed_at"`
}

func CreateTransaction(txType string, amount, feePercent float64) *Transaction {
    amountDec := customtype.NewDecimalFromFloat(amount)
    feePercentDec := customtype.NewDecimalFromFloat(feePercent / 100)
    
    fee := customtype.Decimal{
        Decimal: amountDec.Mul(feePercentDec.Decimal),
        Scale:   customtype.DefaultScale,
    }
    
    netAmount := customtype.Decimal{
        Decimal: amountDec.Sub(fee.Decimal),
        Scale:   customtype.DefaultScale,
    }
    
    return &Transaction{
        Type:        txType,
        Amount:      amountDec,
        Fee:         fee,
        NetAmount:   netAmount,
        Date:        customtype.Date{Time: time.Now()},
        ProcessedAt: customtype.DateTime{Time: time.Now()},
    }
}
```

### API Response with Custom Types

```go
type OrderResponse struct {
    OrderNumber string              `json:"order_number"`
    OrderDate   customtype.Date     `json:"order_date"`
    Items       []OrderItem         `json:"items"`
    Subtotal    customtype.Decimal  `json:"subtotal"`
    Tax         customtype.Decimal  `json:"tax"`
    Total       customtype.Decimal  `json:"total"`
    CreatedAt   customtype.DateTime `json:"created_at"`
}

type OrderItem struct {
    ProductName string              `json:"product_name"`
    Quantity    customtype.Decimal  `json:"quantity"`
    UnitPrice   customtype.Decimal  `json:"unit_price"`
    Total       customtype.Decimal  `json:"total"`
}

func GetOrderHandler(w http.ResponseWriter, r *http.Request) {
    // Fetch order from database
    order := getOrder(orderID)
    
    // Convert to response
    response := OrderResponse{
        OrderNumber: order.OrderNumber,
        OrderDate:   order.OrderDate,
        Items:       convertItems(order.Items),
        Subtotal:    order.Subtotal,
        Tax:         order.Tax,
        Total:       order.Total,
        CreatedAt:   order.CreatedAt,
    }
    
    // Marshal and send
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}
```

## Related Documentation

- [Helpers Overview](index) - All helper packages
- [Cast Package](cast) - Type conversion utilities
- [Validator Package](validator) - Struct validation

---

**Next:** [JSON Package](json) - JSON utilities
