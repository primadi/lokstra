# All Handler Forms - Complete List

> **66 Standard Forms + 3 Special Forms = 69 Total**

## Kombinasi Handler Forms

### Formula
- **6 Input Variations** × **11 Output Variations** = **66 combinations**
- Plus **3 special forms** (http.Handler, http.HandlerFunc, request.HandlerFunc)
- **Total: 69 handler forms**

---

## Input Variations (6)

1. `()` - No input
2. `(*request.Context)` - Context only
3. `(*request.Context, *Param)` - Context + pointer param
4. `(*request.Context, Param)` - Context + value param
5. `(*Param)` - Pointer param only
6. `(Param)` - Value param only

---

## Output Variations (11)

1. `error`
2. `any` - Any type (primitives, struct, slice, map)
3. `(any, error)`
4. `*response.Response` - Unopinionated response helper
5. `response.Response`
6. `*response.ApiHelper` - Opinionated JSON API response
7. `response.ApiHelper`
8. `(*response.Response, error)`
9. `(response.Response, error)`
10. `(*response.ApiHelper, error)`
11. `(response.ApiHelper, error)`

---

## Complete Matrix (66 Forms)

### Group 1: No Input - `func()`
1. `func() error`
2. `func() any`
3. `func() (any, error)`
4. `func() *response.Response`
5. `func() response.Response`
6. `func() *response.ApiHelper`
7. `func() response.ApiHelper`
8. `func() (*response.Response, error)`
9. `func() (response.Response, error)`
10. `func() (*response.ApiHelper, error)`
11. `func() (response.ApiHelper, error)`

### Group 2: Context Only - `func(*request.Context)`
12. `func(*request.Context) error`
13. `func(*request.Context) any`
14. `func(*request.Context) (any, error)`
15. `func(*request.Context) *response.Response`
16. `func(*request.Context) response.Response`
17. `func(*request.Context) *response.ApiHelper`
18. `func(*request.Context) response.ApiHelper`
19. `func(*request.Context) (*response.Response, error)`
20. `func(*request.Context) (response.Response, error)`
21. `func(*request.Context) (*response.ApiHelper, error)`
22. `func(*request.Context) (response.ApiHelper, error)`

### Group 3: Context + Pointer Param - `func(*request.Context, *Param)`
23. `func(*request.Context, *Param) error`
24. `func(*request.Context, *Param) any`
25. `func(*request.Context, *Param) (any, error)`
26. `func(*request.Context, *Param) *response.Response`
27. `func(*request.Context, *Param) response.Response`
28. `func(*request.Context, *Param) *response.ApiHelper`
29. `func(*request.Context, *Param) response.ApiHelper`
30. `func(*request.Context, *Param) (*response.Response, error)`
31. `func(*request.Context, *Param) (response.Response, error)`
32. `func(*request.Context, *Param) (*response.ApiHelper, error)`
33. `func(*request.Context, *Param) (response.ApiHelper, error)`

### Group 4: Context + Value Param - `func(*request.Context, Param)`
34. `func(*request.Context, Param) error`
35. `func(*request.Context, Param) any`
36. `func(*request.Context, Param) (any, error)`
37. `func(*request.Context, Param) *response.Response`
38. `func(*request.Context, Param) response.Response`
39. `func(*request.Context, Param) *response.ApiHelper`
40. `func(*request.Context, Param) response.ApiHelper`
41. `func(*request.Context, Param) (*response.Response, error)`
42. `func(*request.Context, Param) (response.Response, error)`
43. `func(*request.Context, Param) (*response.ApiHelper, error)`
44. `func(*request.Context, Param) (response.ApiHelper, error)`

### Group 5: Pointer Param Only - `func(*Param)`
45. `func(*Param) error`
46. `func(*Param) any`
47. `func(*Param) (any, error)`
48. `func(*Param) *response.Response`
49. `func(*Param) response.Response`
50. `func(*Param) *response.ApiHelper`
51. `func(*Param) response.ApiHelper`
52. `func(*Param) (*response.Response, error)`
53. `func(*Param) (response.Response, error)`
54. `func(*Param) (*response.ApiHelper, error)`
55. `func(*Param) (response.ApiHelper, error)`

### Group 6: Value Param Only - `func(Param)`
56. `func(Param) error`
57. `func(Param) any`
58. `func(Param) (any, error)`
59. `func(Param) *response.Response`
60. `func(Param) response.Response`
61. `func(Param) *response.ApiHelper`
62. `func(Param) response.ApiHelper`
63. `func(Param) (*response.Response, error)`
64. `func(Param) (response.Response, error)`
65. `func(Param) (*response.ApiHelper, error)`
66. `func(Param) (response.ApiHelper, error)`

### Special Forms (+3)
67. `http.Handler`
68. `http.HandlerFunc`
69. `request.HandlerFunc` (sama dengan `func(*request.Context) error`)

---

## Notes

### Param Struct
```go
type Param struct {
    ID   int    `path:"id"`
    Name string `query:"name"`
    Key  string `header:"X-API-Key"`
    Data string `json:"data"`
}
```

### Response Types
- **`any`**: Primitive, struct, slice, map - Lokstra serializes automatically
- **`response.Response`**: Unopinionated helper (`ctx.Resp`) - full control
- **`response.ApiHelper`**: Opinionated JSON API (`ctx.Api`) - structured format

### Pointer vs Value
- **Technically different** but **functionally same purpose**
- Pointer: `*Param` - for large structs or optional fields
- Value: `Param` - for small structs or required fields

---

## Most Common Forms

| Rank | Form | Use Case |
|------|------|----------|
| 1 | `func(*request.Context, Param) (any, error)` | REST APIs |
| 2 | `func(*request.Context) error` | Middleware |
| 3 | `func(*request.Context, *Param) error` | Large param structs |
| 4 | `func(Param) (any, error)` | Simple handlers |
| 5 | `func(*request.Context) any` | Quick responses |

---

## Implementation Status

✅ Structure documented  
📝 Working code examples - **In Progress**  
📝 Test cases - **In Progress**

The main.go file will be updated with all 69 working handler examples.
