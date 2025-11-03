package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/primadi/lokstra/core/app"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/response"
	"github.com/primadi/lokstra/core/router"
	"github.com/primadi/lokstra/core/server"
)

// Param struct untuk parameter binding
type Param struct {
	ID   int    `path:"id"`
	Name string `query:"name"`
	Key  string `header:"X-API-Key"`
	Data string `json:"data"`
}

// ============================================================================
// GROUP 1: No Input - func()
// ============================================================================

// Handler01: func() error
func Handler01() error {
	fmt.Println("Handler 01: func() error")
	return nil
}

// Handler02: func() any
func Handler02() map[string]string {
	return map[string]string{"message": "Handler 02: func() any"}
}

// Handler03: func() (any, error)
func Handler03() (map[string]string, error) {
	return map[string]string{"message": "Handler 03: func() (any, error)"}, nil
}

// Handler04: func() *response.Response - Using helper constructor (recommended)
func Handler04() *response.Response {
	return response.NewJsonResponse(map[string]string{
		"message": "Handler 04: func() *response.Response",
		"method":  "NewJsonResponse - recommended for JSON",
	})
}

// Handler05: func() response.Response - Returning HTML
func Handler05() response.Response {
	r := response.NewResponse()
	r.Html("<h1>Handler 05: HTML Response</h1><p>func() response.Response</p>")
	return *r
}

// Handler06: func() *response.ApiHelper - Using helper constructor (recommended)
func Handler06() *response.ApiHelper {
	return response.NewApiOk(map[string]string{
		"message": "Handler 06: func() *response.ApiHelper",
		"method":  "NewApiOk - recommended for API responses",
	})
}

// Handler07: func() response.ApiHelper - Manual method
func Handler07() response.ApiHelper {
	return *response.NewApiOk(map[string]string{
		"message": "Handler 07: func() response.ApiHelper",
		"note":    "Returning value, not pointer",
	})
}

// Handler08: func() (*response.Response, error) - Plain text response
func Handler08() (*response.Response, error) {
	return response.NewTextResponse("Handler 08: Plain text response\nfunc() (*response.Response, error)"), nil
}

// Handler09: func() (response.Response, error) - Raw bytes (CSV example)
func Handler09() (response.Response, error) {
	csv := "name,value\nHandler09,CSV Response"
	r := response.NewResponse()
	r.Raw("text/csv", []byte(csv))
	return *r, nil
}

// Handler10: func() (*response.ApiHelper, error) - Using helper constructor
func Handler10() (*response.ApiHelper, error) {
	return response.NewApiOk(map[string]string{
		"message": "Handler 10: func() (*response.ApiHelper, error)",
		"method":  "NewApiOk",
	}), nil
}

// Handler11: func() (response.ApiHelper, error) - Return value with error
func Handler11() (response.ApiHelper, error) {
	return *response.NewApiOk(map[string]string{
		"message": "Handler 11: func() (response.ApiHelper, error)",
		"method":  "NewApiOk dereferenced",
	}), nil
}

// ============================================================================
// GROUP 2: Context Only - func(*request.Context)
// ============================================================================

// Handler12: func(*request.Context) error
func Handler12(ctx *request.Context) error {
	fmt.Printf("Handler 12: %s %s\n", ctx.R.Method, ctx.R.URL.Path)
	return nil
}

// Handler13: func(*request.Context) any
func Handler13(ctx *request.Context) map[string]string {
	return map[string]string{
		"message": "Handler 13: func(*request.Context) any",
		"method":  ctx.R.Method,
		"path":    ctx.R.URL.Path,
	}
}

// Handler14: func(*request.Context) (any, error)
func Handler14(ctx *request.Context) (map[string]string, error) {
	return map[string]string{
		"message": "Handler 14: func(*request.Context) (any, error)",
		"method":  ctx.R.Method,
	}, nil
}

// Handler15: func(*request.Context) *response.Response - Using ctx.Resp
func Handler15(ctx *request.Context) *response.Response {
	// Alternative way: using ctx.Resp helper (same as response.NewResponse())
	r := response.NewResponse()
	r.WithStatus(200)
	r.Json(map[string]string{
		"message": "Handler 15: Using chainable methods",
		"method":  ctx.R.Method,
	})
	return r
}

// Handler16: func(*request.Context) response.Response - Stream response example
func Handler16(ctx *request.Context) response.Response {
	r := response.NewResponse()
	r.Stream("text/event-stream", func(w http.ResponseWriter) error {
		fmt.Fprintf(w, "data: Handler 16: Stream response\n\n")
		fmt.Fprintf(w, "data: Path: %s\n\n", ctx.R.URL.Path)
		return nil
	})
	return *r
}

// Handler17: func(*request.Context) *response.ApiHelper - Using helper
func Handler17(ctx *request.Context) *response.ApiHelper {
	return response.NewApiOk(map[string]string{
		"message": "Handler 17: func(*request.Context) *response.ApiHelper",
		"method":  ctx.R.Method,
	})
}

// Handler18: func(*request.Context) response.ApiHelper - Return value
func Handler18(ctx *request.Context) response.ApiHelper {
	return *response.NewApiOk(map[string]string{
		"message": "Handler 18: func(*request.Context) response.ApiHelper",
		"path":    ctx.R.URL.Path,
	})
}

// Handler19: func(*request.Context) (*response.Response, error) - HTML with custom status
func Handler19(ctx *request.Context) (*response.Response, error) {
	html := fmt.Sprintf("<h2>Handler 19</h2><p>Method: %s</p><p>Using NewHtmlResponse</p>", ctx.R.Method)
	r := response.NewHtmlResponse(html)
	r.WithStatus(200)
	return r, nil
}

// Handler20: func(*request.Context) (response.Response, error) - Raw bytes with custom content-type
func Handler20(ctx *request.Context) (response.Response, error) {
	xml := fmt.Sprintf("<?xml version=\"1.0\"?><handler><name>Handler20</name><path>%s</path></handler>", ctx.R.URL.Path)
	return *response.NewRawResponse("application/xml", []byte(xml)), nil
}

// Handler21: func(*request.Context) (*response.ApiHelper, error) - Using helper
func Handler21(ctx *request.Context) (*response.ApiHelper, error) {
	return response.NewApiOk(map[string]string{
		"message": "Handler 21: func(*request.Context) (*response.ApiHelper, error)",
		"method":  ctx.R.Method,
	}), nil
}

// Handler22: func(*request.Context) (response.ApiHelper, error) - Return value
func Handler22(ctx *request.Context) (response.ApiHelper, error) {
	return *response.NewApiOk(map[string]string{
		"message": "Handler 22: func(*request.Context) (response.ApiHelper, error)",
		"path":    ctx.R.URL.Path,
	}), nil
}

// ============================================================================
// GROUP 3: Context + Pointer Param - func(*request.Context, *Param)
// ============================================================================

// Handler23: func(*request.Context, *Param) error
func Handler23(ctx *request.Context, p *Param) error {
	fmt.Printf("Handler 23: ID=%d, Name=%s, Key=%s\n", p.ID, p.Name, p.Key)
	return nil
}

// Handler24: func(*request.Context, *Param) any
func Handler24(ctx *request.Context, p *Param) map[string]any {
	return map[string]any{
		"message": "Handler 24: func(*request.Context, *Param) any",
		"id":      p.ID,
		"name":    p.Name,
		"key":     p.Key,
	}
}

// Handler25: func(*request.Context, *Param) (any, error)
func Handler25(ctx *request.Context, p *Param) (map[string]any, error) {
	return map[string]any{
		"message": "Handler 25: func(*request.Context, *Param) (any, error)",
		"id":      p.ID,
		"name":    p.Name,
	}, nil
}

// Handler26: func(*request.Context, *Param) *response.Response
func Handler26(ctx *request.Context, p *Param) *response.Response {
	r := response.NewResponse()
	r.RespStatusCode = 200
	r.RespData = map[string]any{
		"message": "Handler 26: func(*request.Context, *Param) *response.Response",
		"id":      p.ID,
		"name":    p.Name,
	}
	return r
}

// Handler27: func(*request.Context, *Param) response.Response
func Handler27(ctx *request.Context, p *Param) response.Response {
	return response.Response{
		RespStatusCode: 200,
		RespData: map[string]any{
			"message": "Handler 27: func(*request.Context, *Param) response.Response",
			"id":      p.ID,
			"key":     p.Key,
		},
	}
}

// Handler28: func(*request.Context, *Param) *response.ApiHelper - Using helper
func Handler28(ctx *request.Context, p *Param) *response.ApiHelper {
	return response.NewApiOk(map[string]any{
		"message": "Handler 28: func(*request.Context, *Param) *response.ApiHelper",
		"id":      p.ID,
		"name":    p.Name,
	})
}

// Handler29: func(*request.Context, *Param) response.ApiHelper - Return value
func Handler29(ctx *request.Context, p *Param) response.ApiHelper {
	return *response.NewApiOk(map[string]any{
		"message": "Handler 29: func(*request.Context, *Param) response.ApiHelper",
		"id":      p.ID,
		"key":     p.Key,
	})
}

// Handler30: func(*request.Context, *Param) (*response.Response, error)
func Handler30(ctx *request.Context, p *Param) (*response.Response, error) {
	r := response.NewResponse()
	r.RespStatusCode = 200
	r.RespData = map[string]any{
		"message": "Handler 30: func(*request.Context, *Param) (*response.Response, error)",
		"id":      p.ID,
		"name":    p.Name,
	}
	return r, nil
}

// Handler31: func(*request.Context, *Param) (response.Response, error)
func Handler31(ctx *request.Context, p *Param) (response.Response, error) {
	return response.Response{
		RespStatusCode: 200,
		RespData: map[string]any{
			"message": "Handler 31: func(*request.Context, *Param) (response.Response, error)",
			"id":      p.ID,
			"key":     p.Key,
		},
	}, nil
}

// Handler32: func(*request.Context, *Param) (*response.ApiHelper, error)
func Handler32(ctx *request.Context, p *Param) (*response.ApiHelper, error) {
	api := response.NewApiHelper()
	api.Ok(map[string]any{
		"message": "Handler 32: func(*request.Context, *Param) (*response.ApiHelper, error)",
		"id":      p.ID,
		"name":    p.Name,
	})
	return api, nil
}

// Handler33: func(*request.Context, *Param) (response.ApiHelper, error)
func Handler33(ctx *request.Context, p *Param) (response.ApiHelper, error) {
	api := response.NewApiHelper()
	api.Ok(map[string]any{
		"message": "Handler 33: func(*request.Context, *Param) (response.ApiHelper, error)",
		"id":      p.ID,
		"key":     p.Key,
	})
	return *api, nil
}

// ============================================================================
// GROUP 4: Context + Value Param - func(*request.Context, Param)
// ============================================================================

// Handler34: func(*request.Context, Param) error
func Handler34(ctx *request.Context, p Param) error {
	fmt.Printf("Handler 34: ID=%d, Name=%s, Key=%s\n", p.ID, p.Name, p.Key)
	return nil
}

// Handler35: func(*request.Context, Param) any
func Handler35(ctx *request.Context, p Param) map[string]any {
	return map[string]any{
		"message": "Handler 35: func(*request.Context, Param) any",
		"id":      p.ID,
		"name":    p.Name,
	}
}

// Handler36: func(*request.Context, Param) (any, error)
func Handler36(ctx *request.Context, p Param) (map[string]any, error) {
	return map[string]any{
		"message": "Handler 36: func(*request.Context, Param) (any, error)",
		"id":      p.ID,
		"name":    p.Name,
	}, nil
}

// Handler37: func(*request.Context, Param) *response.Response
func Handler37(ctx *request.Context, p Param) *response.Response {
	r := response.NewResponse()
	r.RespStatusCode = 200
	r.RespData = map[string]any{
		"message": "Handler 37: func(*request.Context, Param) *response.Response",
		"id":      p.ID,
		"name":    p.Name,
	}
	return r
}

// Handler38: func(*request.Context, Param) response.Response
func Handler38(ctx *request.Context, p Param) response.Response {
	return response.Response{
		RespStatusCode: 200,
		RespData: map[string]any{
			"message": "Handler 38: func(*request.Context, Param) response.Response",
			"id":      p.ID,
			"key":     p.Key,
		},
	}
}

// Handler39: func(*request.Context, Param) *response.ApiHelper
func Handler39(ctx *request.Context, p Param) *response.ApiHelper {
	api := response.NewApiHelper()
	api.Ok(map[string]any{
		"message": "Handler 39: func(*request.Context, Param) *response.ApiHelper",
		"id":      p.ID,
		"name":    p.Name,
	})
	return api
}

// Handler40: func(*request.Context, Param) response.ApiHelper
func Handler40(ctx *request.Context, p Param) response.ApiHelper {
	api := response.NewApiHelper()
	api.Ok(map[string]any{
		"message": "Handler 40: func(*request.Context, Param) response.ApiHelper",
		"id":      p.ID,
		"key":     p.Key,
	})
	return *api
}

// Handler41: func(*request.Context, Param) (*response.Response, error)
func Handler41(ctx *request.Context, p Param) (*response.Response, error) {
	r := response.NewResponse()
	r.RespStatusCode = 200
	r.RespData = map[string]any{
		"message": "Handler 41: func(*request.Context, Param) (*response.Response, error)",
		"id":      p.ID,
		"name":    p.Name,
	}
	return r, nil
}

// Handler42: func(*request.Context, Param) (response.Response, error)
func Handler42(ctx *request.Context, p Param) (response.Response, error) {
	return response.Response{
		RespStatusCode: 200,
		RespData: map[string]any{
			"message": "Handler 42: func(*request.Context, Param) (response.Response, error)",
			"id":      p.ID,
			"key":     p.Key,
		},
	}, nil
}

// Handler43: func(*request.Context, Param) (*response.ApiHelper, error)
func Handler43(ctx *request.Context, p Param) (*response.ApiHelper, error) {
	api := response.NewApiHelper()
	api.Ok(map[string]any{
		"message": "Handler 43: func(*request.Context, Param) (*response.ApiHelper, error)",
		"id":      p.ID,
		"name":    p.Name,
	})
	return api, nil
}

// Handler44: func(*request.Context, Param) (response.ApiHelper, error)
func Handler44(ctx *request.Context, p Param) (response.ApiHelper, error) {
	api := response.NewApiHelper()
	api.Ok(map[string]any{
		"message": "Handler 44: func(*request.Context, Param) (response.ApiHelper, error)",
		"id":      p.ID,
		"key":     p.Key,
	})
	return *api, nil
}

// ============================================================================
// GROUP 5: Pointer Param Only - func(*Param)
// ============================================================================

// Handler45: func(*Param) error
func Handler45(p *Param) error {
	fmt.Printf("Handler 45: ID=%d, Name=%s\n", p.ID, p.Name)
	return nil
}

// Handler46: func(*Param) any
func Handler46(p *Param) map[string]any {
	return map[string]any{
		"message": "Handler 46: func(*Param) any",
		"id":      p.ID,
		"name":    p.Name,
	}
}

// Handler47: func(*Param) (any, error)
func Handler47(p *Param) (map[string]any, error) {
	return map[string]any{
		"message": "Handler 47: func(*Param) (any, error)",
		"id":      p.ID,
		"name":    p.Name,
	}, nil
}

// Handler48: func(*Param) *response.Response
func Handler48(p *Param) *response.Response {
	r := response.NewResponse()
	r.RespStatusCode = 200
	r.RespData = map[string]any{
		"message": "Handler 48: func(*Param) *response.Response",
		"id":      p.ID,
		"name":    p.Name,
	}
	return r
}

// Handler49: func(*Param) response.Response
func Handler49(p *Param) response.Response {
	return response.Response{
		RespStatusCode: 200,
		RespData: map[string]any{
			"message": "Handler 49: func(*Param) response.Response",
			"id":      p.ID,
			"key":     p.Key,
		},
	}
}

// Handler50: func(*Param) *response.ApiHelper
func Handler50(p *Param) *response.ApiHelper {
	api := response.NewApiHelper()
	api.Ok(map[string]any{
		"message": "Handler 50: func(*Param) *response.ApiHelper",
		"id":      p.ID,
		"name":    p.Name,
	})
	return api
}

// Handler51: func(*Param) response.ApiHelper
func Handler51(p *Param) response.ApiHelper {
	api := response.NewApiHelper()
	api.Ok(map[string]any{
		"message": "Handler 51: func(*Param) response.ApiHelper",
		"id":      p.ID,
		"key":     p.Key,
	})
	return *api
}

// Handler52: func(*Param) (*response.Response, error)
func Handler52(p *Param) (*response.Response, error) {
	r := response.NewResponse()
	r.RespStatusCode = 200
	r.RespData = map[string]any{
		"message": "Handler 52: func(*Param) (*response.Response, error)",
		"id":      p.ID,
		"name":    p.Name,
	}
	return r, nil
}

// Handler53: func(*Param) (response.Response, error)
func Handler53(p *Param) (response.Response, error) {
	return response.Response{
		RespStatusCode: 200,
		RespData: map[string]any{
			"message": "Handler 53: func(*Param) (response.Response, error)",
			"id":      p.ID,
			"key":     p.Key,
		},
	}, nil
}

// Handler54: func(*Param) (*response.ApiHelper, error)
func Handler54(p *Param) (*response.ApiHelper, error) {
	api := response.NewApiHelper()
	api.Ok(map[string]any{
		"message": "Handler 54: func(*Param) (*response.ApiHelper, error)",
		"id":      p.ID,
		"name":    p.Name,
	})
	return api, nil
}

// Handler55: func(*Param) (response.ApiHelper, error)
func Handler55(p *Param) (response.ApiHelper, error) {
	api := response.NewApiHelper()
	api.Ok(map[string]any{
		"message": "Handler 55: func(*Param) (response.ApiHelper, error)",
		"id":      p.ID,
		"key":     p.Key,
	})
	return *api, nil
}

// ============================================================================
// GROUP 6: Value Param Only - func(Param)
// ============================================================================

// Handler56: func(Param) error
func Handler56(p Param) error {
	fmt.Printf("Handler 56: ID=%d, Name=%s\n", p.ID, p.Name)
	return nil
}

// Handler57: func(Param) any
func Handler57(p Param) map[string]any {
	return map[string]any{
		"message": "Handler 57: func(Param) any",
		"id":      p.ID,
		"name":    p.Name,
	}
}

// Handler58: func(Param) (any, error)
func Handler58(p Param) (map[string]any, error) {
	return map[string]any{
		"message": "Handler 58: func(Param) (any, error)",
		"id":      p.ID,
		"name":    p.Name,
	}, nil
}

// Handler59: func(Param) *response.Response
func Handler59(p Param) *response.Response {
	r := response.NewResponse()
	r.RespStatusCode = 200
	r.RespData = map[string]any{
		"message": "Handler 59: func(Param) *response.Response",
		"id":      p.ID,
		"name":    p.Name,
	}
	return r
}

// Handler60: func(Param) response.Response
func Handler60(p Param) response.Response {
	return response.Response{
		RespStatusCode: 200,
		RespData: map[string]any{
			"message": "Handler 60: func(Param) response.Response",
			"id":      p.ID,
			"key":     p.Key,
		},
	}
}

// Handler61: func(Param) *response.ApiHelper
func Handler61(p Param) *response.ApiHelper {
	api := response.NewApiHelper()
	api.Ok(map[string]any{
		"message": "Handler 61: func(Param) *response.ApiHelper",
		"id":      p.ID,
		"name":    p.Name,
	})
	return api
}

// Handler62: func(Param) response.ApiHelper
func Handler62(p Param) response.ApiHelper {
	api := response.NewApiHelper()
	api.Ok(map[string]any{
		"message": "Handler 62: func(Param) response.ApiHelper",
		"id":      p.ID,
		"key":     p.Key,
	})
	return *api
}

// Handler63: func(Param) (*response.Response, error)
func Handler63(p Param) (*response.Response, error) {
	r := response.NewResponse()
	r.RespStatusCode = 200
	r.RespData = map[string]any{
		"message": "Handler 63: func(Param) (*response.Response, error)",
		"id":      p.ID,
		"name":    p.Name,
	}
	return r, nil
}

// Handler64: func(Param) (response.Response, error)
func Handler64(p Param) (response.Response, error) {
	return response.Response{
		RespStatusCode: 200,
		RespData: map[string]any{
			"message": "Handler 64: func(Param) (response.Response, error)",
			"id":      p.ID,
			"key":     p.Key,
		},
	}, nil
}

// Handler65: func(Param) (*response.ApiHelper, error)
func Handler65(p Param) (*response.ApiHelper, error) {
	api := response.NewApiHelper()
	api.Ok(map[string]any{
		"message": "Handler 65: func(Param) (*response.ApiHelper, error)",
		"id":      p.ID,
		"name":    p.Name,
	})
	return api, nil
}

// Handler66: func(Param) (response.ApiHelper, error)
func Handler66(p Param) (response.ApiHelper, error) {
	api := response.NewApiHelper()
	api.Ok(map[string]any{
		"message": "Handler 66: func(Param) (response.ApiHelper, error)",
		"id":      p.ID,
		"key":     p.Key,
	})
	return *api, nil
}

// ============================================================================
// SPECIAL FORMS (+3)
// ============================================================================

// Handler67: http.Handler
type Handler67 struct{}

func (h Handler67) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Handler 67: http.Handler",
		"method":  r.Method,
		"path":    r.URL.Path,
	})
}

// Handler68: http.HandlerFunc
func Handler68(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Handler 68: http.HandlerFunc",
		"method":  r.Method,
		"path":    r.URL.Path,
	})
}

// Handler69: request.HandlerFunc (sama dengan func(*request.Context) error)
// Ini sebenarnya sama dengan Handler12, tapi diberi nama berbeda untuk kelengkapan
var Handler69 request.HandlerFunc = func(ctx *request.Context) error {
	return ctx.Resp.WithStatus(200).Json(map[string]string{
		"message": "Handler 69: request.HandlerFunc",
		"method":  ctx.R.Method,
		"path":    ctx.R.URL.Path,
	})
}

func main() {
	// Create router
	r := router.New("all-handler-forms")

	// ========== GROUP 1: No Input ==========
	r.GET("/h01", Handler01)
	r.GET("/h02", Handler02)
	r.GET("/h03", Handler03)
	r.GET("/h04", Handler04)
	r.GET("/h05", Handler05)
	r.GET("/h06", Handler06)
	r.GET("/h07", Handler07)
	r.GET("/h08", Handler08)
	r.GET("/h09", Handler09)
	r.GET("/h10", Handler10)
	r.GET("/h11", Handler11)

	// ========== GROUP 2: Context Only ==========
	r.GET("/h12", Handler12)
	r.GET("/h13", Handler13)
	r.GET("/h14", Handler14)
	r.GET("/h15", Handler15)
	r.GET("/h16", Handler16)
	r.GET("/h17", Handler17)
	r.GET("/h18", Handler18)
	r.GET("/h19", Handler19)
	r.GET("/h20", Handler20)
	r.GET("/h21", Handler21)
	r.GET("/h22", Handler22)

	// ========== GROUP 3: Context + *Param ==========
	r.GET("/h23/:id", Handler23)
	r.GET("/h24/:id", Handler24)
	r.GET("/h25/:id", Handler25)
	r.GET("/h26/:id", Handler26)
	r.GET("/h27/:id", Handler27)
	r.GET("/h28/:id", Handler28)
	r.GET("/h29/:id", Handler29)
	r.GET("/h30/:id", Handler30)
	r.GET("/h31/:id", Handler31)
	r.GET("/h32/:id", Handler32)
	r.GET("/h33/:id", Handler33)

	// ========== GROUP 4: Context + Param ==========
	r.GET("/h34/:id", Handler34)
	r.GET("/h35/:id", Handler35)
	r.GET("/h36/:id", Handler36)
	r.GET("/h37/:id", Handler37)
	r.GET("/h38/:id", Handler38)
	r.GET("/h39/:id", Handler39)
	r.GET("/h40/:id", Handler40)
	r.GET("/h41/:id", Handler41)
	r.GET("/h42/:id", Handler42)
	r.GET("/h43/:id", Handler43)
	r.GET("/h44/:id", Handler44)

	// ========== GROUP 5: *Param Only ==========
	r.GET("/h45/:id", Handler45)
	r.GET("/h46/:id", Handler46)
	r.GET("/h47/:id", Handler47)
	r.GET("/h48/:id", Handler48)
	r.GET("/h49/:id", Handler49)
	r.GET("/h50/:id", Handler50)
	r.GET("/h51/:id", Handler51)
	r.GET("/h52/:id", Handler52)
	r.GET("/h53/:id", Handler53)
	r.GET("/h54/:id", Handler54)
	r.GET("/h55/:id", Handler55)

	// ========== GROUP 6: Param Only ==========
	r.GET("/h56/:id", Handler56)
	r.GET("/h57/:id", Handler57)
	r.GET("/h58/:id", Handler58)
	r.GET("/h59/:id", Handler59)
	r.GET("/h60/:id", Handler60)
	r.GET("/h61/:id", Handler61)
	r.GET("/h62/:id", Handler62)
	r.GET("/h63/:id", Handler63)
	r.GET("/h64/:id", Handler64)
	r.GET("/h65/:id", Handler65)
	r.GET("/h66/:id", Handler66)

	// ========== SPECIAL FORMS ==========
	r.GET("/h67", Handler67{})
	r.GET("/h68", http.HandlerFunc(Handler68))
	r.GET("/h69", Handler69)

	// Print summary
	fmt.Println("=== ALL HANDLER FORMS DEMO ===")
	fmt.Println("Total: 69 handler forms registered")
	fmt.Println("Groups 1-2: No params (h01-h22)")
	fmt.Println("Groups 3-6: With params (h23-h66)")
	fmt.Println("Special: http.Handler, http.HandlerFunc, request.HandlerFunc (h67-h69)")
	fmt.Println("\nServer starting on :8080...")

	// Create app with router and address, then start server
	a := app.New("demo-app", ":8080", r)
	srv := server.New("demo-server", a)
	if err := srv.Start(); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
