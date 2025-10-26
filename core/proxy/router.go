package proxy

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"strings"

	"github.com/primadi/lokstra/common/json"
)

type Router struct {
	handler   http.Handler
	remoteUrl string
}

func NewRemoteRouter(remoteUrl string) *Router {
	return &Router{
		remoteUrl: remoteUrl,
	}
}

func NewLocalRouter(handler http.Handler) *Router {
	return &Router{
		handler: handler,
	}
}

func (p *Router) Serve(req *http.Request) (*http.Response, error) {
	if p.handler != nil {
		rw := httptest.NewRecorder()
		p.handler.ServeHTTP(rw, req)
		return rw.Result(), nil
	}

	remoteReq := cloneRequest(req, p.remoteUrl)
	return http.DefaultClient.Do(remoteReq)
}

// DoJSON performs any method with JSON body and returns decoded response.
func (p *Router) DoJSON(method, path string, headers map[string]string,
	reqBody, respBody any) error {
	if reqBody == nil {
		reqBody = map[string]any{}
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := p.Serve(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		data, _ := io.ReadAll(resp.Body)
		rawErr := fmt.Errorf("http %d: %s", resp.StatusCode, string(data))
		return ParseRouterError(rawErr)
	}

	if respBody != nil {
		return json.NewDecoder(resp.Body).Decode(respBody)
	}
	return nil
}

// ------------------------------
// HELPERS
// ------------------------------

func cloneRequest(req *http.Request, remoteBase string) *http.Request {
	newReq := req.Clone(req.Context())

	// rebuild target URL
	u, _ := url.Parse(remoteBase)
	u.Path = path.Join(u.Path, req.URL.Path)
	u.RawQuery = req.URL.RawQuery
	newReq.URL = u

	return newReq
}

func ParseRouterError(err error) error {
	errMsg := err.Error()

	// Check if this is a proxy error (format: "http XXX: {json}")
	if !strings.HasPrefix(errMsg, "http ") {
		return err
	}

	// Find the JSON part after "http XXX: "
	idx := strings.Index(errMsg, ": ")
	if idx == -1 {
		return err
	}

	jsonPart := errMsg[idx+2:]

	// Try to parse as error response
	var errorResp struct {
		Status string `json:"status"`
		Error  struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal([]byte(jsonPart), &errorResp); err == nil {
		// Successfully parsed, return just the message
		return fmt.Errorf("%s", errorResp.Error.Message)
	}

	// If parsing fails, return original error
	return err
}
