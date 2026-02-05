package service

import (
	"fmt"
	"sync"
	"time"

	"github.com/primadi/lokstra/core/proxy"
)

// ========================================
// Models
// ========================================

type WeatherData struct {
	City        string  `json:"city"`
	Temperature float64 `json:"temperature"`
	Condition   string  `json:"condition"`
	Humidity    int     `json:"humidity"`
	WindSpeed   float64 `json:"wind_speed"`
	Timestamp   string  `json:"timestamp"`
}

type ForecastData struct {
	City     string          `json:"city"`
	Forecast []DailyForecast `json:"forecast"`
}

type DailyForecast struct {
	Date      string  `json:"date"`
	TempHigh  float64 `json:"temp_high"`
	TempLow   float64 `json:"temp_low"`
	Condition string  `json:"condition"`
}

type WeatherReport struct {
	ID          string        `json:"id"`
	City        string        `json:"city"`
	Current     *WeatherData  `json:"current"`
	Forecast    *ForecastData `json:"forecast,omitempty"`
	RequestedAt time.Time     `json:"requested_at"`
}

type GetWeatherReportParams struct {
	City            string `query:"city"`
	IncludeForecast bool   `query:"forecast"`
	ForecastDays    int    `query:"days"`
}

// ========================================
// Weather Service
// ========================================

// WeatherService demonstrates using proxy.Router for quick API access
// No service wrapper needed - direct HTTP calls to external API
type WeatherService struct {
	weatherAPI *proxy.Router
}

var (
	reports   = make(map[string]*WeatherReport)
	reportsMu sync.RWMutex
	reportID  = 1
)

// Create fetches weather from external API using proxy.Router
// Using standard REST method name (Create) for auto-routing: POST /weather-reports
func (s *WeatherService) Create(p *GetWeatherReportParams) (*WeatherReport, error) {
	if p.City == "" {
		return nil, fmt.Errorf("city is required")
	}

	// Default forecast days
	if p.ForecastDays == 0 {
		p.ForecastDays = 5
	}

	reportsMu.Lock()
	id := fmt.Sprintf("report_%d", reportID)
	reportID++
	reportsMu.Unlock()

	report := &WeatherReport{
		ID:          id,
		City:        p.City,
		RequestedAt: time.Now(),
	}

	// Fetch current weather using proxy.Router (no service wrapper!)
	var current WeatherData
	err := s.weatherAPI.DoJSON(
		"GET",
		fmt.Sprintf("/weather/%s", p.City),
		nil,
		nil,
		&current,
	)

	if err != nil {
		return nil, err
	}

	report.Current = &current

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
			return nil, err
		}

		report.Forecast = &forecast
	}

	// Repository report
	reportsMu.Lock()
	reports[id] = report
	reportsMu.Unlock()

	return report, nil
}

// ========================================
// Factory
// ========================================

func WeatherServiceFactory(deps map[string]any, config map[string]any) any {
	// Get URL from config and create proxy.Router
	url, ok := config["weather-api-url"].(string)
	if !ok {
		panic("weather-api-url is not a string")
	}

	return &WeatherService{
		weatherAPI: proxy.NewRemoteRouter(url),
	}
}
