package response

// import (
// 	"lokstra/common/response/response_iface"
// 	"lokstra/core"
// 	"strings"
// )

// registry formatter
// var formatterRegistry = map[string]response_iface.ResponseFormatter{}

// RegisterFormatter allows adding a new formatter for a MIME type
// func RegisterFormatter(mime string, f response_iface.ResponseFormatter) {
// 	formatterRegistry[mime] = f
// }

// SelectFormatter selects based on Accept header (e.g. "application/json")
// func SelectFormatter(accept string) response_iface.ResponseFormatter {
// 	for _, mime := range parseAcceptHeader(accept) {
// 		if f, ok := formatterRegistry[mime]; ok {
// 			return f
// 		}
// 	}
// 	return core.GlobalRuntime().GetResponseFormatter()
// }

// func parseAcceptHeader(accept string) []string {
// 	parts := strings.Split(accept, ",")
// 	result := make([]string, 0, len(parts))
// 	for _, p := range parts {
// 		mime := strings.TrimSpace(strings.Split(p, ";")[0])
// 		if mime != "" {
// 			result = append(result, mime)
// 		}
// 	}
// 	return result
// }
