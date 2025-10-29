package main

import (
	"fmt"
	"log"
	"os"

	"github.com/primadi/lokstra"
)

// GetSecret retrieves secret from environment
func GetSecret(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("secret %s not found", key)
	}
	return value, nil
}

// Status handler shows which secrets are configured
func StatusHandler() map[string]any {
	dbPass, _ := GetSecret("DB_PASSWORD")
	apiKey, _ := GetSecret("API_KEY")
	jwtSecret, _ := GetSecret("JWT_SECRET")

	return map[string]any{
		"db_password_set": dbPass != "",
		"api_key_set":     apiKey != "",
		"jwt_secret_set":  jwtSecret != "",
		"warning":         "Never expose actual secret values!",
	}
}

// Home handler
func HomeHandler() string {
	return `
	<html>
	<body>
		<h1>Secrets Management Example</h1>
		<p>Demonstrates secure handling of sensitive configuration</p>
		<h2>Endpoints</h2>
		<ul>
			<li><a href="/status">Status</a> - Check which secrets are configured</li>
		</ul>
		<h2>Best Practices</h2>
		<ul>
			<li>✅ Store secrets in environment variables</li>
			<li>✅ Use secret management services (Vault, AWS Secrets Manager)</li>
			<li>❌ Never commit secrets to version control</li>
			<li>❌ Never log secret values</li>
			<li>❌ Never expose secrets in API responses</li>
		</ul>
		<h2>Running with Secrets</h2>
		<pre>
DB_PASSWORD=secret123 \
API_KEY=apikey456 \
JWT_SECRET=jwtsecret789 \
go run main.go
		</pre>
	</body>
	</html>
	`
}

func main() {
	// Validate required secrets
	log.Println("Validating secrets...")

	requiredSecrets := []string{"DB_PASSWORD", "API_KEY", "JWT_SECRET"}
	for _, secret := range requiredSecrets {
		if _, err := GetSecret(secret); err != nil {
			log.Printf("⚠️  Warning: %v", err)
		} else {
			log.Printf("✅ Secret %s is configured", secret)
		}
	}

	// Create router
	router := lokstra.NewRouter("main")
	router.GET("/", HomeHandler)
	router.GET("/status", StatusHandler)

	// Create and run app
	app := lokstra.NewApp("secrets-mgmt", ":3050", router)

	log.Println("Starting server on :3050")
	if err := app.Run(0); err != nil {
		log.Fatal(err)
	}
}
