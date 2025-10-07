package custom_resolver

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"

	lokstraConfig "github.com/primadi/lokstra/core/config"
)

type AWSSecretsResolver struct {
	client *secretsmanager.Client
	cache  map[string]string
}

// creates a new AWS Secrets Manager resolver
func NewAWSSecretsResolver(region string) (*AWSSecretsResolver, error) {
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS SDK config: %w", err)
	}

	return &AWSSecretsResolver{
		client: secretsmanager.NewFromConfig(cfg),
		cache:  make(map[string]string),
	}, nil
}

// retrieves secret from AWS Secrets Manager
func (r *AWSSecretsResolver) Resolve(source string, key string, defaultValue string) (string, bool) {
	// Check cache first
	if value, ok := r.cache[key]; ok {
		return value, true
	}

	// Get secret from AWS
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(key),
	}

	result, err := r.client.GetSecretValue(context.TODO(), input)
	if err != nil {
		log.Printf("Failed to get secret %s from AWS: %v (using default)", key, err)
		return defaultValue, false
	}

	var secretValue string
	if result.SecretString != nil {
		secretValue = *result.SecretString
	} else {
		// Binary secret
		secretValue = string(result.SecretBinary)
	}

	// Cache the value
	r.cache[key] = secretValue
	return secretValue, true
}

var _ lokstraConfig.VariableResolver = (*AWSSecretsResolver)(nil)

// Example usage
func ExampleAWSSecretsResolver() {
	// Register AWS Secrets Manager resolver
	awsResolver, err := NewAWSSecretsResolver("us-east-1")
	if err != nil {
		log.Fatal(err)
	}
	lokstraConfig.AddVariableResolver("AWS", awsResolver)

	// Now you can use ${@AWS:secret-name} in your YAML configs
	// Example: config/production.yaml
	//
	// services:
	//   - name: postgres
	//     type: dbpool_pg
	//     config:
	//       dsn: ${@AWS:database-connection-string}
	//
	//   - name: api
	//     config:
	//       apiKey: ${@AWS:api-key-secret}
	//       jwtSecret: ${@AWS:jwt-signing-key}
}
