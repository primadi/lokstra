package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/response"
	"github.com/primadi/lokstra/core/router"
)

func main() {
	log.Println("Testing Return Type 'any' Only Support")

	r := router.New("test-api")

	// Test 1: Return data only (no error)
	r.GET("/test1", func(c *request.Context) map[string]any {
		return map[string]any{
			"test":    "1",
			"message": "Data return only (no error)",
		}
	})

	// Test 2: Return *Response only (no error)
	r.GET("/test2", func(c *request.Context) *response.Response {
		resp := response.NewResponse()
		resp.WithStatus(http.StatusCreated).Json(map[string]any{
			"test":    "2",
			"message": "Response return only (no error)",
		})
		return resp
	})

	// Test 3: Return *ApiHelper only (no error)
	r.GET("/test3", func(c *request.Context) *response.ApiHelper {
		api := response.NewApiHelper()
		api.Ok(map[string]any{
			"test":    "3",
			"message": "ApiHelper return only (no error)",
		})
		return api
	})

	// Test 4: With struct param, return data only
	type test4Params struct {
		ID int `path:"id"`
	}
	r.GET("/test4/{id}", func(p *test4Params) map[string]any {
		return map[string]any{
			"test":    "4",
			"id":      p.ID,
			"message": "Struct param with data return only",
		}
	})

	// Test 5: No context, return *Response only
	r.GET("/test5", func() *response.Response {
		resp := response.NewResponse()
		resp.WithStatus(http.StatusTeapot).Text("I'm a teapot (no context, no error)")
		return resp
	})

	// Test 6: Return nil *Response (should send default success)
	r.GET("/test6", func(c *request.Context) *response.Response {
		return nil
	})

	// Test 7: Mixed - c.Resp set but return value should override
	r.GET("/test7", func(c *request.Context) *response.Response {
		// This should be IGNORED
		c.Resp.WithStatus(http.StatusOK).Json(map[string]any{
			"source": "c.Resp (should be ignored)",
		})

		// This should be USED
		resp := response.NewResponse()
		resp.WithStatus(http.StatusCreated).Json(map[string]any{
			"test":    "7",
			"source":  "return value (should be used)",
			"message": "Return value overrides c.Resp",
		})
		return resp
	})

	// Test 8: Compare with standard (data, error) pattern
	r.GET("/test8", func(c *request.Context) (map[string]any, error) {
		return map[string]any{
			"test":    "8",
			"message": "Standard (data, error) pattern",
		}, nil
	})

	r.PrintRoutes()

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("✅ Return Type 'any' Only - Test Server Started")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("\nTest endpoints:")
	fmt.Println("  GET http://localhost:3000/test1 - Data only")
	fmt.Println("  GET http://localhost:3000/test2 - *Response only")
	fmt.Println("  GET http://localhost:3000/test3 - *ApiHelper only")
	fmt.Println("  GET http://localhost:3000/test4/123 - Struct param")
	fmt.Println("  GET http://localhost:3000/test5 - No context")
	fmt.Println("  GET http://localhost:3000/test6 - Nil return")
	fmt.Println("  GET http://localhost:3000/test7 - Priority test")
	fmt.Println("  GET http://localhost:3000/test8 - Standard pattern")
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println()

	log.Fatal(http.ListenAndServe(":3000", r))
}
