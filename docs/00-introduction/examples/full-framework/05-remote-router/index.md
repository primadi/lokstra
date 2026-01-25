# Example 07 - Remote Router (proxy.Router)

This example demonstrates how to use **`proxy.Router`** for **quick, direct HTTP calls** to external APIs **without creating service wrappers**. Perfect for simple integrations and one-off API calls.

## 📋 What You'll Learn

- ✅ Using `proxy.Router` for direct HTTP calls
- ✅ No service wrapper needed (simpler than `proxy.Service`)
- ✅ Simple URL config (no special definitions)
- ✅ Quick integration without convention/metadata
- ✅ When to use `proxy.Router` vs `proxy.Service`
- ✅ Error handling with external APIs

## 🏗️ Architecture

```
┌────────────────────────────────────────────────────────────┐
│                      Main App (:3001)                      │
│                                                            │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  WeatherService                                      │  │
│  │  - GetWeatherReport()  → POST /weather-reports      │  │
│  │                                                      │  │
│  │  Uses: proxy.Router.DoJSON()                         │  │
│  │  No wrapper! Direct HTTP calls!                      │  │
│  └─────────────────┬────────────────────────────────────┘  │
│                    │ Direct HTTP calls                     │
└────────────────────┼───────────────────────────────────────┘
                     ▼
     ┌───────────────────────────────────────────────┐
     │   Mock Weather API (:9001)                    │
     │   (Simulates OpenWeather, etc.)               │
     │                                               │
     │   GET /weather/{city}                         │
     │   GET /forecast/{city}?days=5                 │
     └───────────────────────────────────────────────┘
```

**Key difference from Example 06:**
- ❌ No `PaymentServiceRemote` wrapper
- ❌ No `RegisterServiceType` for remote
- ✅ Direct `proxy.Router.DoJSON()` calls
- ✅ Much simpler for quick integrations!

## 🚀 How to Run

### Step 1: Start Mock Weather API

```bash
cd mock-weather-api
go run main.go
```

This starts the mock weather API on `http://localhost:9001`.

### Step 2: Start Main Application

```bash
# From the example root directory
go run main.go
```

This starts the main application on `http://localhost:3001`.

### Step 3: Test with HTTP Requests

Use the `test.http` file or curl:

```bash
# Get weather report (current only)
curl -X POST "http://localhost:3001/weather-reports?city=jakarta&forecast=false"

# Get weather report with forecast
curl -X POST "http://localhost:3001/weather-reports?city=jakarta&forecast=true&days=5"

# Different city
curl -X POST "http://localhost:3001/weather-reports?city=bandung&forecast=true&days=3"
```

## 📂 Project Structure

```
07-remote-router/
├── main.go                           # Main application entry point
├── config.yaml                       # Simple URL config
├── test.http                         # HTTP test scenarios
├── index                         # This file
│
├── mock-weather-api/
│   └── main.go                       # Mock weather API
│
└── service/
    └── weather_service.go            # Service using proxy.Router
```

## 🔑 Key Concepts

### 1. Simple URL Configuration

Repository URL directly in service config - no special definitions needed:

```yaml
service-definitions:
  weather-service:
    type: weather-service-factory
    config:
      weather-api-url: "http://localhost:9001"  # Direct URL
```

**Why this is simpler:**
- ✅ No separate router-definitions section
- ✅ URL directly in service config
- ✅ Easy to override per environment
- ✅ Clear configuration intent

Factory creates the router from URL:

```go
func WeatherServiceFactory(deps map[string]any, config map[string]any) any {
    url := config["weather-api-url"].(string)
    return &WeatherService{
        weatherAPI: proxy.NewRemoteRouter(url),
    }
}
```

### 2. Service Using proxy.Router

Direct HTTP calls without wrapper:

```go
type WeatherService struct {
    weatherAPI *proxy.Router
}

func (s *WeatherService) GetWeatherReport(p *GetWeatherReportParams) (*WeatherReport, error) {
    // Direct HTTP call - no wrapper!
    var current WeatherData
    err := s.weatherAPI.DoJSON(
        "GET",
        fmt.Sprintf("/weather/%s", p.City),
        nil,    // headers
        nil,    // request body
        &current, // response body
    )
    
    if err != nil {
        return nil, fmt.Errorf("failed to fetch weather: %w", err)
    }
    
    // Optionally fetch forecast
    if p.IncludeForecast {
        var forecast ForecastData
        err := s.weatherAPI.DoJSON(
            "GET",
            fmt.Sprintf("/forecast/%s?days=%d", p.City, p.ForecastDays),
            nil,
            nil,
            &forecast,
        )
        
        if err != nil {
            return nil, fmt.Errorf("failed to fetch forecast: %w", err)
        }
        
        report.Forecast = &forecast
    }
    
    return report, nil
}
```

**Key points:**
- ✅ Direct `DoJSON()` calls
- ✅ Manual URL construction
- ✅ No service wrapper needed
- ✅ Simple and straightforward

### 3. Factory Pattern

Create `proxy.Router` from URL in config:

```go
func WeatherServiceFactory(deps map[string]any, config map[string]any) any {
    // Read URL from config
    url, ok := config["weather-api-url"].(string)
    if !ok {
        panic("weather-api-url is not a string")
    }
    
    // Create router directly
    return &WeatherService{
        weatherAPI: proxy.NewRemoteRouter(url),
    }
}
```

**Simple instantiation:**
- Read URL from config
- Create router with `proxy.NewRemoteRouter(url)`
- No framework injection needed

### 4. Service Registration

Simple registration without route overrides:

```go
lokstra_registry.RegisterServiceType("weather-service-factory",
    svc.WeatherServiceFactory, nil,
    deploy.WithResource("weather-report", "weather-reports"),
    deploy.WithConvention("rest"),
    // No route overrides needed!
)
```

Method names match REST convention:
- `GetWeatherReport()` → `GET /weather-reports/{id}` (not used in this example)
- Or accessed via `POST /weather-reports` with query params

## 🎯 proxy.Router API

### Available Methods

1. **DoJSON** - Most flexible (recommended)
```go
err := router.DoJSON(
    method string,        // "GET", "POST", etc.
    path string,          // "/weather/jakarta"
    headers map[string]string,
    requestBody any,      // nil for GET
    responseBody any,     // pointer to struct
)
```

2. **Get** - Simple GET requests
```go
resp, err := router.Get("/weather/jakarta", headers)
```

3. **PostJSON** - POST with JSON body
```go
resp, err := router.PostJSON("/endpoint", data, headers)
```

4. **Serve** - Low-level HTTP request
```go
resp, err := router.Serve(httpRequest)
```

## 🔄 Comparison: proxy.Router vs proxy.Service

| Aspect | proxy.Router (This Example) | proxy.Service (Example 06) |
|--------|----------------------------|---------------------------|
| **Use Case** | Quick API access | Structured services |
| **Setup** | Minimal (just URL) | Service wrapper + metadata |
| **Convention** | ❌ Manual paths | ✅ Auto-routing |
| **Type Safety** | ✅ Response only | ✅ Request + Response |
| **Service Wrapper** | ❌ Not needed | ✅ Required |
| **Route Overrides** | N/A | ✅ Supported |
| **Best For** | One-off calls, prototyping | Multi-endpoint services |

### When to Use proxy.Router

✅ **USE proxy.Router when:**
- Quick integration needed
- One-off API calls
- Prototyping/testing external APIs
- Simple endpoints (1-3 calls)
- No need for reusable service abstraction
- Example: Weather API, currency converter, IP geolocation

### When to Use proxy.Service

✅ **USE proxy.Service when:**
- Multiple related endpoints
- Need service abstraction
- Want typed methods
- Dependency injection required
- Complex business logic
- Example: Payment gateway, email service, SMS provider

## 💡 Real-World Examples

### Good Use Cases for proxy.Router

```go
// Weather API
weatherRouter := proxy.NewRemoteRouter("https://api.weather.com")
weatherRouter.DoJSON("GET", "/current/jakarta", nil, nil, &weather)

// Currency Converter
currencyRouter := proxy.NewRemoteRouter("https://api.exchangerate.com")
currencyRouter.DoJSON("GET", "/latest?base=USD", nil, nil, &rates)

// IP Geolocation
ipRouter := proxy.NewRemoteRouter("https://ipapi.co")
ipRouter.DoJSON("GET", "/json", nil, nil, &location)
```

### When to Upgrade to proxy.Service

If you find yourself:
- Making 5+ calls to same API
- Repeating URL construction logic
- Need to mock for testing
- Want stronger typing
- Service used by multiple parts of code

**Then**: Create a proper service wrapper with `proxy.Service` (see Example 06).

## 🎓 Learning Points

### 1. Simplicity vs Structure Trade-off

**proxy.Router**: Simple, quick, less code
```go
// One-liner!
router.DoJSON("GET", "/weather/jakarta", nil, nil, &weather)
```

**proxy.Service**: More code, better structure
```go
// Typed method
weather, err := weatherService.GetWeather(&GetWeatherParams{
    City: "jakarta",
})
```

### 2. Manual URL Construction

With `proxy.Router`, you build URLs manually:
```go
// Manual path construction
path := fmt.Sprintf("/forecast/%s?days=%d", city, days)
err := router.DoJSON("GET", path, nil, nil, &forecast)
```

With `proxy.Service`, framework does it:
```go
// Framework builds URL from metadata + method name
forecast, err := service.GetForecast(&ForecastParams{
    City: city,
    Days: days,
})
```

### 3. Error Handling

Same pattern for both:
```go
if err != nil {
    return nil, fmt.Errorf("failed to fetch data: %w", err)
}
```

Use `proxy.ParseRouterError()` for better error messages:
```go
if err != nil {
    return nil, proxy.ParseRouterError(err)
}
```

### 4. Configuration Pattern

**proxy.Router**: Simple router definition
```yaml
router-definitions:
  api-name:
    url: "https://api.example.com"
```

**proxy.Service**: Full service definition
```yaml
external-service-definitions:
  api-name:
    url: "https://api.example.com"
    type: api-service-remote-factory
```

## 🧪 Mock Weather API

Built with Lokstra, demonstrates clean API design:

```go
func getCurrentWeather(req *GetWeatherRequest) (*WeatherData, error) {
    weather, exists := mockWeather[req.City]
    if !exists {
        return nil, fmt.Errorf("weather data not found for city: %s", req.City)
    }
    
    result := *weather
    result.Timestamp = time.Now().Format(time.RFC3339)
    return &result, nil
}

func main() {
    r := lokstra.NewRouter("weather-api")
    
    r.GET("/weather/{city}", getCurrentWeather)
    r.GET("/forecast/{city}", getForecast)
    
    app := lokstra.NewApp("weather-api", ":9001", r)
    if err := app.Run(30 * time.Second); err != nil {
        fmt.Println("Error starting server:", err)
    }
}
```

**Available cities:** jakarta, bandung, surabaya

## 🔄 Next Steps

1. ✅ **Example 06** - External Services (`proxy.Service` with wrappers)
2. ✅ **Example 07** - Remote Router (You are here)

## 📚 Related Documentation

- [Architecture - Proxy Patterns](../../architecture#proxy-patterns)
- [Example 06 - External Services](../06-external-services)
- [Remote Services Guide](../../../02-framework-guide/02-service)

---

**💡 Key Takeaway:** Use `proxy.Router` for quick, simple API integrations. Upgrade to `proxy.Service` when you need structure, typing, and reusability. Choose based on complexity, not cargo-culting!
