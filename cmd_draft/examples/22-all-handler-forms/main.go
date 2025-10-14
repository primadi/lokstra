package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/primadi/lokstra/core/app"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/router"
)

// Request struct for testing
type UserRequest struct {
	ID   string `path:"id"`
	Name string `json:"name"`
}

type MessageRequest struct {
	Message string `json:"message"`
}

// ============================================================================
// FORM 1: func() error
// ============================================================================

func Handler1NoParams() error {
	fmt.Println("✅ Form 1: func() error")
	return nil
}

// ============================================================================
// FORM 2: func() (data, error)
// ============================================================================

func Handler2NoParamsWithData() (any, error) {
	fmt.Println("✅ Form 2: func() (data, error)")
	return map[string]string{
		"message": "Form 2 works!",
		"form":    "func() (data, error)",
	}, nil
}

// ============================================================================
// FORM 3: func(*T) error
// ============================================================================

func Handler3StructOnly(req *UserRequest) error {
	fmt.Printf("✅ Form 3: func(*T) error - ID: %s, Name: %s\n", req.ID, req.Name)
	return nil
}

// ============================================================================
// FORM 4: func(*T) (data, error)
// ============================================================================

func Handler4StructWithData(req *UserRequest) (any, error) {
	fmt.Printf("✅ Form 4: func(*T) (data, error) - ID: %s, Name: %s\n", req.ID, req.Name)
	return map[string]any{
		"message": "Form 4 works!",
		"form":    "func(*T) (data, error)",
		"request": req,
	}, nil
}

// ============================================================================
// FORM 5: func(*request.Context) error
// ============================================================================

func Handler5ContextOnly(ctx *request.Context) error {
	fmt.Println("✅ Form 5: func(*request.Context) error")
	id := ctx.Req.PathParam("id", "")
	fmt.Printf("   Path param ID: %s\n", id)
	return nil
}

// ============================================================================
// FORM 6: func(*request.Context) (data, error)
// ============================================================================

func Handler6ContextWithData(ctx *request.Context) (any, error) {
	fmt.Println("✅ Form 6: func(*request.Context) (data, error)")
	id := ctx.Req.PathParam("id", "")
	return map[string]string{
		"message": "Form 6 works!",
		"form":    "func(*request.Context) (data, error)",
		"id":      id,
	}, nil
}

// ============================================================================
// FORM 7: func(*request.Context, *T) error
// ============================================================================

func Handler7ContextAndStruct(ctx *request.Context, req *UserRequest) error {
	fmt.Printf("✅ Form 7: func(*request.Context, *T) error - ID: %s, Name: %s\n", req.ID, req.Name)
	fmt.Printf("   Request method: %s\n", ctx.R.Method)
	return nil
}

// ============================================================================
// FORM 8: func(*request.Context, *T) (data, error)
// ============================================================================

func Handler8ContextAndStructWithData(ctx *request.Context, req *UserRequest) (any, error) {
	fmt.Printf("✅ Form 8: func(*request.Context, *T) (data, error) - ID: %s, Name: %s\n", req.ID, req.Name)
	return map[string]any{
		"message": "Form 8 works!",
		"form":    "func(*request.Context, *T) (data, error)",
		"method":  ctx.R.Method,
		"request": req,
	}, nil
}

// ============================================================================
// FORM 9: http.HandlerFunc (standard Go handler)
// ============================================================================

func Handler9StandardHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println("✅ Form 9: http.HandlerFunc (standard Go handler)")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := `{"message":"Form 9 works!","form":"http.HandlerFunc"}`
	w.Write([]byte(response))
}

func main() {
	// Create router and register all 9 handler forms
	r := router.New("")

	r.GET("/form1", Handler1NoParams)
	r.GET("/form2", Handler2NoParamsWithData)
	r.GET("/form3/{id}", Handler3StructOnly)
	r.GET("/form4/{id}", Handler4StructWithData)
	r.GET("/form5/{id}", Handler5ContextOnly)
	r.GET("/form6/{id}", Handler6ContextWithData)
	r.POST("/form7/{id}", Handler7ContextAndStruct)
	r.POST("/form8/{id}", Handler8ContextAndStructWithData)
	r.GET("/form9", Handler9StandardHTTP)

	// Print test instructions
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("🧪 All Handler Forms Test Server (9 Forms)")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("\n📋 Test Commands (run in separate terminal):")
	fmt.Println("\n# Form 1: func() error")
	fmt.Println("curl http://localhost:3000/form1")

	fmt.Println("\n# Form 2: func() (data, error)")
	fmt.Println("curl http://localhost:3000/form2")

	fmt.Println("\n# Form 3: func(*T) error")
	fmt.Println(`curl "http://localhost:3000/form3/user123?name=Alice"`)

	fmt.Println("\n# Form 4: func(*T) (data, error)")
	fmt.Println(`curl "http://localhost:3000/form4/user456?name=Bob"`)

	fmt.Println("\n# Form 5: func(*request.Context) error")
	fmt.Println("curl http://localhost:3000/form5/ctx789")

	fmt.Println("\n# Form 6: func(*request.Context) (data, error)")
	fmt.Println("curl http://localhost:3000/form6/ctx999")

	fmt.Println("\n# Form 7: func(*request.Context, *T) error")
	fmt.Println(`curl -X POST "http://localhost:3000/form7/post123?name=Charlie"`)

	fmt.Println("\n# Form 8: func(*request.Context, *T) (data, error)")
	fmt.Println(`curl -X POST -H "Content-Type: application/json" -d "{\"name\":\"Diana\"}" http://localhost:3000/form8/post456`)

	fmt.Println("\n# Form 9: http.HandlerFunc (standard Go handler)")
	fmt.Println("curl http://localhost:3000/form9")

	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("\n🚀 Server starting on http://localhost:3000")

	// Create application and start
	application := app.New("handler-forms-test", ":3000", r)
	if err := application.Run(30 * time.Second); err != nil {
		log.Fatal(err)
	}
}
