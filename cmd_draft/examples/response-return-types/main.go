package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/response"
	"github.com/primadi/lokstra/core/response/api_formatter"
	"github.com/primadi/lokstra/core/router"
)

func main() {
	// Setup API formatter
	// this is default behavior, you can omit this line
	// api_formatter.SetGlobalFormatter(api_formatter.NewApiResponseFormatter())

	r := router.New("response-return-types-demo")

	// ===================================
	// Group 1: Regular Data Returns
	// ===================================
	r.Group("/regular", func(g1 router.Router) {
		g1.GET("/user", func(c *request.Context) (map[string]any, error) {
			return map[string]any{
				"id":   1,
				"name": "John Doe",
			}, nil
		})

		g1.GET("/users", func(c *request.Context) ([]map[string]any, error) {
			return []map[string]any{
				{"id": 1, "name": "John"},
				{"id": 2, "name": "Jane"},
			}, nil
		})
	})

	// ===================================
	// Group 2: Response Pointer Returns
	// ===================================
	r.Group("/response", func(g2 router.Router) {
		// Custom status code
		g2.POST("/created", func(c *request.Context) (*response.Response, error) {
			resp := response.NewResponse()
			resp.WithStatus(http.StatusCreated).Json(map[string]string{
				"message": "Resource created",
				"id":      "123",
			})
			return resp, nil
		})

		// Plain text response
		g2.GET("/text", func(c *request.Context) (*response.Response, error) {
			resp := response.NewResponse()
			resp.WithStatus(http.StatusOK).Text("Hello, this is plain text!")
			return resp, nil
		})

		// HTML response
		g2.GET("/html", func(c *request.Context) (*response.Response, error) {
			resp := response.NewResponse()
			html := `
			<!DOCTYPE html>
			<html>
			<head><title>Lokstra Response</title></head>
			<body>
				<h1>Custom HTML Response</h1>
				<p>Generated at: ` + time.Now().Format(time.RFC3339) + `</p>
			</body>
			</html>
		`
			resp.WithStatus(http.StatusOK).Html(html)
			return resp, nil
		})

		// Streaming response
		g2.GET("/stream", func(c *request.Context) (*response.Response, error) {
			resp := response.NewResponse()
			resp.Stream("text/event-stream", func(w http.ResponseWriter) error {
				for i := 1; i <= 5; i++ {
					fmt.Fprintf(w, "data: Event %d at %s\n\n", i, time.Now().Format("15:04:05"))
					if f, ok := w.(http.Flusher); ok {
						f.Flush()
					}
					time.Sleep(500 * time.Millisecond)
				}
				return nil
			})
			return resp, nil
		})

		// Custom headers
		g2.GET("/custom-headers", func(c *request.Context) (*response.Response, error) {
			resp := response.NewResponse()
			resp.RespHeaders = map[string][]string{
				"X-Custom-Header":  {"custom-value"},
				"X-Request-ID":     {"req-123456"},
				"X-Rate-Limit":     {"1000"},
				"X-Rate-Remaining": {"999"},
			}
			resp.WithStatus(http.StatusOK).Json(map[string]string{
				"message": "Response with custom headers",
			})
			return resp, nil
		})

		// Nil response (default success)
		g2.GET("/nil", func(c *request.Context) (*response.Response, error) {
			return nil, nil // Will send default success
		})
	})

	// ===================================
	// Group 3: ApiHelper Returns
	// ===================================
	r.Group("/api-helper", func(g3 router.Router) {
		// Standard success
		g3.GET("/success", func(c *request.Context) (*response.ApiHelper, error) {
			api := response.NewApiHelper()
			api.Ok(map[string]string{
				"status": "success",
				"data":   "operation completed",
			})
			return api, nil
		})

		// Created with message
		g3.POST("/create", func(c *request.Context) (*response.ApiHelper, error) {
			api := response.NewApiHelper()
			resource := map[string]any{
				"id":   "new-123",
				"name": "New Resource",
			}
			api.Created(resource, "Resource created successfully")
			return api, nil
		})

		// List with pagination
		g3.GET("/list", func(c *request.Context) (*response.ApiHelper, error) {
			api := response.NewApiHelper()

			items := []map[string]any{
				{"id": 1, "name": "Item 1"},
				{"id": 2, "name": "Item 2"},
				{"id": 3, "name": "Item 3"},
			}

			meta := &api_formatter.ListMeta{
				Page:       1,
				PageSize:   10,
				Total:      50,
				TotalPages: 5,
			}

			api.OkList(items, meta)
			return api, nil
		})

		// Error responses
		g3.GET("/not-found", func(c *request.Context) (*response.ApiHelper, error) {
			api := response.NewApiHelper()
			api.NotFound("Resource not found")
			return api, nil
		})

		g3.GET("/unauthorized", func(c *request.Context) (*response.ApiHelper, error) {
			api := response.NewApiHelper()
			api.Unauthorized("Invalid credentials")
			return api, nil
		})

		g3.GET("/validation-error", func(c *request.Context) (*response.ApiHelper, error) {
			api := response.NewApiHelper()
			fields := []api_formatter.FieldError{
				{Field: "email", Message: "Invalid email format"},
				{Field: "password", Message: "Password too short"},
			}
			api.ValidationError("Validation failed", fields)
			return api, nil
		})

		// ApiHelper with custom headers
		g3.GET("/custom-headers", func(c *request.Context) (*response.ApiHelper, error) {
			api := response.NewApiHelper()

			// Access underlying Response for custom headers
			resp := api.Resp()
			resp.RespHeaders = map[string][]string{
				"X-API-Version": {"v1.0"},
				"X-Resource-ID": {"res-456"},
			}

			api.Ok(map[string]string{
				"message": "API response with custom headers",
			})
			return api, nil
		})
	})
	// ===================================
	// Group 4: Error Handling Priority
	// ===================================
	r.Group("/error-priority", func(g4 router.Router) {
		// Error takes precedence over Response
		g4.GET("/response-error", func(c *request.Context) (*response.Response, error) {
			resp := response.NewResponse()
			resp.WithStatus(http.StatusOK).Json(map[string]string{
				"message": "This will be IGNORED",
			})

			// Error takes precedence!
			return resp, errors.New("something went wrong")
		})

		// Error takes precedence over ApiHelper
		g4.GET("/api-error", func(c *request.Context) (*response.ApiHelper, error) {
			api := response.NewApiHelper()
			api.Ok(map[string]string{
				"message": "This will also be IGNORED",
			})

			// Error takes precedence!
			return api, errors.New("api error occurred")
		})
	})

	// ===================================
	// Group 5: Without Context Parameter
	// ===================================
	r.Group("/no-context", func(g5 router.Router) {
		type GreetRequest struct {
			Name string `query:"name"`
		}

		// Struct parameter only
		g5.GET("/greet", func(req *GreetRequest) (*response.Response, error) {
			if req.Name == "" {
				req.Name = "Guest"
			}

			resp := response.NewResponse()
			resp.WithStatus(http.StatusOK).Json(map[string]string{
				"greeting": "Hello, " + req.Name + "!",
			})
			return resp, nil
		})

		// No parameters at all
		g5.GET("/ping", func() (*response.Response, error) {
			resp := response.NewResponse()
			resp.WithStatus(http.StatusOK).Json(map[string]string{
				"message": "pong",
				"time":    time.Now().Format(time.RFC3339),
			})
			return resp, nil
		})
	})

	// ===================================
	// Group 6: Mixed Examples
	// ===================================
	r.Group("/mixed", func(g6 router.Router) {
		// Response value (not pointer)
		g6.GET("/response-value", func(c *request.Context) (response.Response, error) {
			resp := response.NewResponse()
			resp.WithStatus(http.StatusAccepted).Json(map[string]string{
				"status": "accepted",
			})
			return *resp, nil // Return value, not pointer
		})

		// ApiHelper value (not pointer)
		g6.GET("/api-value", func(c *request.Context) (response.ApiHelper, error) {
			api := response.NewApiHelper()
			api.Ok(map[string]int{"count": 42})
			return *api, nil // Return value, not pointer
		})

		// Conditional response type
		g6.GET("/conditional", func(c *request.Context) (*response.Response, error) {
			format := c.R.URL.Query().Get("format")

			resp := response.NewResponse()
			data := map[string]string{
				"message": "Hello World",
			}

			switch format {
			case "html":
				resp.WithStatus(http.StatusOK).Html(
					"<h1>" + data["message"] + "</h1>",
				)
			case "text":
				resp.WithStatus(http.StatusOK).Text(data["message"])
			default:
				resp.WithStatus(http.StatusOK).Json(data)
			}

			return resp, nil
		})
	})

	// ===================================
	// Start Server
	// ===================================
	fmt.Println("Server started at http://localhost:8080")
	fmt.Println("\nTry these endpoints:")
	fmt.Println("  Regular data:")
	fmt.Println("    GET  http://localhost:8080/regular/user")
	fmt.Println("    GET  http://localhost:8080/regular/users")
	fmt.Println("\n  Response returns:")
	fmt.Println("    POST http://localhost:8080/response/created")
	fmt.Println("    GET  http://localhost:8080/response/text")
	fmt.Println("    GET  http://localhost:8080/response/html")
	fmt.Println("    GET  http://localhost:8080/response/stream")
	fmt.Println("    GET  http://localhost:8080/response/custom-headers")
	fmt.Println("\n  ApiHelper returns:")
	fmt.Println("    GET  http://localhost:8080/api-helper/success")
	fmt.Println("    POST http://localhost:8080/api-helper/create")
	fmt.Println("    GET  http://localhost:8080/api-helper/list")
	fmt.Println("    GET  http://localhost:8080/api-helper/not-found")
	fmt.Println("\n  Error priority:")
	fmt.Println("    GET  http://localhost:8080/error-priority/response-error")
	fmt.Println("    GET  http://localhost:8080/error-priority/api-error")
	fmt.Println("\n  Without context:")
	fmt.Println("    GET  http://localhost:8080/no-context/greet?name=John")
	fmt.Println("    GET  http://localhost:8080/no-context/ping")
	fmt.Println("\n  Mixed examples:")
	fmt.Println("    GET  http://localhost:8080/mixed/conditional?format=html")
	fmt.Println("    GET  http://localhost:8080/mixed/conditional?format=text")
	fmt.Println("    GET  http://localhost:8080/mixed/conditional?format=json")

	http.ListenAndServe(":8080", r)
}
