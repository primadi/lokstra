# Do() Method Implementation Summary

## 📋 **Jawaban: Do() sebaiknya di Handler**

### ✅ **Implementasi Selesai**

#### **Files Created/Modified:**
1. **`handler.go`** - Added `Do()` method
2. **`customstep.go`** - New file with `customStep` implementation  
3. **`do_test.go`** - Comprehensive unit tests
4. **`do_examples.go`** - Usage examples
5. **`EXECUTION_GUIDE.md`** - Updated documentation

---

## 🏗️ **Architecture Decision**

### **Handler (✅ Chosen)**
- **Fluent API**: Maintains method chaining pattern
- **Build-time**: Perfect for configuring flow pipeline
- **Consistency**: Follows existing pattern like `Done()`, `BeginTx()`
- **Separation**: Clear distinction from runtime execution

### **Context (❌ Rejected)**
- Runtime container, breaks build-time pattern
- Would disrupt fluent API chaining
- Not consistent with current architecture

### **Step (❌ Rejected)**  
- Too low-level for user interaction
- Requires manual step creation every time
- Not user-friendly

---

## 🔧 **Implementation Details**

### **Handler Method:**
```go
// Do adds a custom function step to the handler pipeline.
func (h *Handler) Do(fn func(*Context) error) *Handler {
    step := &customStep{fn: fn}
    h.steps = append(h.steps, step)
    return h
}
```

### **CustomStep Implementation:**
```go
type customStep struct {
    fn func(*Context) error
}

func (s *customStep) Run(ctx *Context) error {
    return s.fn(ctx)
}

func (s *customStep) Meta() StepMeta {
    return StepMeta{
        Name: "custom.function",
        Kind: StepNormal,
    }
}
```

---

## 📖 **Usage Examples**

### **1. Simple Custom Logic**
```go
handler := flow.NewHandler(regCtx, "simple").
    ExecSql("INSERT INTO logs (message) VALUES (?)", "Starting").
    Done().
    Do(func(ctx *Context) error {
        log.Println("Custom logic executed")
        ctx.Set("process_started", true)
        return nil
    }).
    ExecSql("UPDATE logs SET status = 'completed'").
    Done()
```

### **2. Business Logic with Validation**
```go
handler := flow.NewHandler(regCtx, "validation").
    QueryRowSql("SELECT balance FROM accounts WHERE id = ?", accountID).
    SaveAs("account_balance").
    Do(func(ctx *Context) error {
        balanceData, _ := ctx.Get("account_balance")
        balance := balanceData.(map[string]any)["balance"].(float64)
        
        if balance < 100.0 {
            ctx.Set("low_balance_warning", true)
            ctx.Set("notification_message", "Low balance warning")
        }
        
        return nil
    }).
    ExecSql("INSERT INTO notifications (message) VALUES (?)", "{{.notification_message}}").
    Done()
```

### **3. External Service Integration**
```go
handler := flow.NewHandler(regCtx, "payment").
    QueryRowSql("SELECT email, amount FROM orders WHERE id = ?", orderID).
    SaveAs("order_data").
    Do(func(ctx *Context) error {
        orderData, _ := ctx.Get("order_data")
        order := orderData.(map[string]any)
        
        // Call external payment service
        paymentResult, err := paymentService.Process(
            order["email"].(string), 
            order["amount"].(float64)
        )
        if err != nil {
            return fmt.Errorf("payment failed: %w", err)
        }
        
        ctx.Set("payment_id", paymentResult.ID)
        return nil
    }).
    ExecSql("UPDATE orders SET payment_id = ? WHERE id = ?", "{{.payment_id}}", orderID).
    Done()
```

---

## ✅ **Benefits**

1. **Flexible**: Support any custom logic within flow
2. **Consistent**: Follows established Handler pattern
3. **Chainable**: Maintains fluent API
4. **Testable**: Full unit test coverage
5. **Error Handling**: Proper error propagation
6. **Variable Access**: Full access to Context variables

---

## 🧪 **Test Coverage**

- ✅ Method creation and chaining
- ✅ Step execution and error handling  
- ✅ Variable access and modification
- ✅ Fluent API integration
- ✅ Meta information correctness

---

## 📚 **Documentation Updated**

- ✅ EXECUTION_GUIDE.md with Do() examples
- ✅ Comprehensive code examples
- ✅ Best practices guidance
- ✅ Integration patterns

---

## 🎯 **Conclusion**

`Do()` method successfully implemented in **Handler** with:
- **Proper architecture alignment**
- **Full test coverage** 
- **Complete documentation**
- **Real-world usage examples**
- **Clean integration** with existing flow system

The implementation provides maximum flexibility for custom business logic while maintaining the clean, chainable API that Lokstra Flow is designed around.
