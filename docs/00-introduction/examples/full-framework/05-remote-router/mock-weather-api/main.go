package main

import (
	"fmt"
	"log"
	"time"

	"github.com/primadi/lokstra"
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

type GetWeatherRequest struct {
	City string `path:"city"`
}

type GetForecastRequest struct {
	City string `path:"city"`
	Days int    `query:"days"`
}

// ========================================
// Mock Data
// ========================================

var mockWeather = map[string]*WeatherData{
	"jakarta": {
		City:        "Jakarta",
		Temperature: 32.5,
		Condition:   "Partly Cloudy",
		Humidity:    75,
		WindSpeed:   12.5,
	},
	"bandung": {
		City:        "Bandung",
		Temperature: 26.0,
		Condition:   "Sunny",
		Humidity:    60,
		WindSpeed:   8.0,
	},
	"surabaya": {
		City:        "Surabaya",
		Temperature: 31.0,
		Condition:   "Rainy",
		Humidity:    80,
		WindSpeed:   15.0,
	},
}

// ========================================
// Handlers
// ========================================

func getCurrentWeather(req *GetWeatherRequest) (*WeatherData, error) {
	weather, exists := mockWeather[req.City]
	if !exists {
		return nil, fmt.Errorf("weather data not found for city: %s", req.City)
	}

	// Add timestamp
	result := *weather
	result.Timestamp = time.Now().Format(time.RFC3339)

	log.Printf("🌤️  Weather requested: %s - %.1f°C %s", req.City, result.Temperature, result.Condition)

	return &result, nil
}

func getForecast(req *GetForecastRequest) (*ForecastData, error) {
	// Default 5 days
	if req.Days == 0 {
		req.Days = 5
	}

	if req.Days > 10 {
		return nil, fmt.Errorf("maximum forecast days is 10")
	}

	_, exists := mockWeather[req.City]
	if !exists {
		return nil, fmt.Errorf("weather data not found for city: %s", req.City)
	}

	// Generate mock forecast
	forecast := &ForecastData{
		City:     req.City,
		Forecast: make([]DailyForecast, req.Days),
	}

	baseTemp := mockWeather[req.City].Temperature
	conditions := []string{"Sunny", "Partly Cloudy", "Cloudy", "Rainy", "Stormy"}

	for i := 0; i < req.Days; i++ {
		date := time.Now().AddDate(0, 0, i+1).Format("2006-01-02")
		forecast.Forecast[i] = DailyForecast{
			Date:      date,
			TempHigh:  baseTemp + float64(i%3),
			TempLow:   baseTemp - float64(3+i%2),
			Condition: conditions[i%len(conditions)],
		}
	}

	log.Printf("📅 Forecast requested: %s - %d days", req.City, req.Days)

	return forecast, nil
}

// ========================================
// Main
// ========================================

func main() {
	// Create router
	r := lokstra.NewRouter("weather-api")

	// Routes
	r.GET("/weather/{city}", getCurrentWeather)
	r.GET("/forecast/{city}", getForecast)

	// Start server
	app := lokstra.NewApp("weather-api", ":9001", r)

	fmt.Println("==========================================================")
	fmt.Println("🌦️  Mock Weather API (Lokstra)")
	fmt.Println("==========================================================")
	fmt.Println()
	fmt.Println("Running on: http://localhost:9001")
	fmt.Println()
	fmt.Println("Endpoints:")
	fmt.Println("  GET /weather/{city}             - Current weather")
	fmt.Println("  GET /forecast/{city}?days=5     - Weather forecast")
	fmt.Println()
	fmt.Println("Available cities: jakarta, bandung, surabaya")
	fmt.Println()
	fmt.Println("==========================================================")
	fmt.Println()

	if err := app.Run(30 * time.Second); err != nil {
		log.Fatalf("Failed to run app: %v", err)
	}
}
