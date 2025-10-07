package custom_resolver

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	lokstraConfig "github.com/primadi/lokstra/core/config"
)

// JSONSecretsResolver resolves secrets from AWS Secrets Manager JSON values
type JSONSecretsResolver struct {
	client *secretsmanager.Client
	cache  map[string]map[string]interface{}
}

// NewJSONSecretsResolver creates resolver for JSON secrets
func NewJSONSecretsResolver(region string) (*JSONSecretsResolver, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS SDK config: %w", err)
	}

	return &JSONSecretsResolver{
		client: secretsmanager.NewFromConfig(cfg),
		cache:  make(map[string]map[string]interface{}),
	}, nil
}

// Resolve retrieves JSON field from AWS secret
// Key format: "secret-name.field-path"
// Example: "database-creds.username" → gets "username" field from "database-creds" secret
func (r *JSONSecretsResolver) Resolve(source string, key string, defaultValue string) (string, bool) {
	// Parse key: "secret-name.field-path"
	secretName, fieldPath, found := parseJSONKey(key)
	if !found {
		log.Printf("Invalid JSON secret key format: %s (expected: secret-name.field)", key)
		return defaultValue, false
	}

	// Check cache
	if secret, ok := r.cache[secretName]; ok {
		if value, ok := getJSONField(secret, fieldPath); ok {
			return fmt.Sprint(value), true
		}
	}

	// Get secret from AWS
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	}

	result, err := r.client.GetSecretValue(context.TODO(), input)
	if err != nil {
		log.Printf("Failed to get secret %s from AWS: %v", secretName, err)
		return defaultValue, false
	}

	if result.SecretString == nil {
		log.Printf("Secret %s is not a JSON string", secretName)
		return defaultValue, false
	}

	// Parse JSON
	var secret map[string]interface{}
	if err := json.Unmarshal([]byte(*result.SecretString), &secret); err != nil {
		log.Printf("Failed to parse JSON secret %s: %v", secretName, err)
		return defaultValue, false
	}

	// Cache the secret
	r.cache[secretName] = secret

	// Get field value
	if value, ok := getJSONField(secret, fieldPath); ok {
		return fmt.Sprint(value), true
	}

	return defaultValue, false
}

var _ lokstraConfig.VariableResolver = (*JSONSecretsResolver)(nil)

// Helper to parse "secret-name.field-path" format
func parseJSONKey(key string) (secretName, fieldPath string, ok bool) {
	for i := 0; i < len(key); i++ {
		if key[i] == '.' {
			return key[:i], key[i+1:], true
		}
	}
	return "", "", false
}

// Helper to get nested JSON field
func getJSONField(data map[string]interface{}, path string) (interface{}, bool) {
	value, ok := data[path]
	return value, ok
}

// Example usage
func ExampleJSONSecretsResolver() {
	// Register JSON secrets resolver
	jsonResolver, err := NewJSONSecretsResolver("us-east-1")
	if err != nil {
		log.Fatal(err)
	}
	lokstraConfig.AddVariableResolver("AWSJSON", jsonResolver)

	// AWS Secret "database-creds" contains:
	// {
	//   "host": "db.example.com",
	//   "port": "5432",
	//   "username": "admin",
	//   "password": "secret123",
	//   "database": "production"
	// }
	//
	// YAML config can reference individual fields:
	//
	// services:
	//   - name: postgres
	//     type: dbpool_pg
	//     config:
	//       host: ${@AWSJSON:database-creds.host}
	//       port: ${@AWSJSON:database-creds.port}
	//       user: ${@AWSJSON:database-creds.username}
	//       password: ${@AWSJSON:database-creds.password}
	//       database: ${@AWSJSON:database-creds.database}
}
