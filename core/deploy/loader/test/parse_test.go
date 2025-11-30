package loader_test

import (
	"testing"

	"github.com/primadi/lokstra/core/deploy/loader/internal"
)

func TestParseKeyDefault(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedKey     string
		expectedDefault string
	}{
		// Without quotes - FIRST ':' is separator
		{
			name:            "simple key without default",
			input:           "db.host",
			expectedKey:     "db.host",
			expectedDefault: "",
		},
		{
			name:            "simple key with default",
			input:           "db.host:localhost",
			expectedKey:     "db.host",
			expectedDefault: "localhost",
		},
		{
			name:            "vault path with colon - ambiguous",
			input:           "secret/data/db:password",
			expectedKey:     "secret/data/db",
			expectedDefault: "password",
		},
		{
			name:            "URL as default value",
			input:           "DB_URL:postgresql://localhost:5432/db",
			expectedKey:     "DB_URL",
			expectedDefault: "postgresql://localhost:5432/db",
		},
		{
			name:            "ARN without default - ambiguous interpretation",
			input:           "arn:aws:secretsmanager:us-east-1:123456789:secret:db-password",
			expectedKey:     "arn", // Only first part before FIRST ':'
			expectedDefault: "aws:secretsmanager:us-east-1:123456789:secret:db-password",
		},

		// With single quotes - ':' inside quotes is part of key
		{
			name:            "quoted key without default",
			input:           `'db.host'`,
			expectedKey:     "db.host",
			expectedDefault: "",
		},
		{
			name:            "quoted key with default",
			input:           `'db.host':localhost`,
			expectedKey:     "db.host",
			expectedDefault: "localhost",
		},
		{
			name:            "quoted key with colons - no default",
			input:           `'secret/data/db:password'`,
			expectedKey:     "secret/data/db:password",
			expectedDefault: "",
		},
		{
			name:            "quoted key with colons - with default",
			input:           `'secret/data/db:password':fallback`,
			expectedKey:     "secret/data/db:password",
			expectedDefault: "fallback",
		},
		{
			name:            "quoted ARN",
			input:           `'arn:aws:secretsmanager:us-east-1:123456789:secret:db-password'`,
			expectedKey:     "arn:aws:secretsmanager:us-east-1:123456789:secret:db-password",
			expectedDefault: "",
		},
		{
			name:            "quoted ARN with default",
			input:           `'arn:aws:secretsmanager:us-east-1:123456789:secret:db-password':fallback`,
			expectedKey:     "arn:aws:secretsmanager:us-east-1:123456789:secret:db-password",
			expectedDefault: "fallback",
		},
		{
			name:            "quoted config key with colon",
			input:           `'db:url'`,
			expectedKey:     "db:url",
			expectedDefault: "",
		},
		{
			name:            "quoted config key with colon and URL default",
			input:           `'db:url':postgresql://localhost:5432/db`,
			expectedKey:     "db:url",
			expectedDefault: "postgresql://localhost:5432/db",
		},

		// Edge cases
		{
			name:            "empty string",
			input:           "",
			expectedKey:     "",
			expectedDefault: "",
		},
		{
			name:            "only colon",
			input:           ":",
			expectedKey:     "",
			expectedDefault: "",
		},
		{
			name:            "malformed quote - no closing",
			input:           `'unclosed`,
			expectedKey:     `'unclosed`,
			expectedDefault: "",
		},
		{
			name:            "empty quoted key",
			input:           `''`,
			expectedKey:     "",
			expectedDefault: "",
		},
		{
			name:            "empty quoted key with default",
			input:           `'':default`,
			expectedKey:     "",
			expectedDefault: "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, defaultValue := internal.ParseKeyDefault(tt.input)

			if key != tt.expectedKey {
				t.Errorf("parseKeyDefault(%q): key = %q, want %q", tt.input, key, tt.expectedKey)
			}

			if defaultValue != tt.expectedDefault {
				t.Errorf("parseKeyDefault(%q): defaultValue = %q, want %q", tt.input, defaultValue, tt.expectedDefault)
			}
		})
	}
}

func TestParseKeyDefault_RealWorldExamples(t *testing.T) {
	// Real-world scenarios
	tests := []struct {
		scenario        string
		input           string
		expectedKey     string
		expectedDefault string
		explanation     string
	}{
		{
			scenario:        "AWS ARN without quotes",
			input:           "arn:aws:secretsmanager:us-east-1:123456789:secret:db-password",
			expectedKey:     "arn",
			expectedDefault: "aws:secretsmanager:us-east-1:123456789:secret:db-password",
			explanation:     "WRONG - need quotes to preserve full ARN as key",
		},
		{
			scenario:        "AWS ARN with quotes",
			input:           `'arn:aws:secretsmanager:us-east-1:123456789:secret:db-password'`,
			expectedKey:     "arn:aws:secretsmanager:us-east-1:123456789:secret:db-password",
			expectedDefault: "",
			explanation:     "CORRECT - single quotes preserve full ARN",
		},
		{
			scenario:        "Postgres DSN as default",
			input:           "DB_DSN:postgresql://user:pass@localhost:5432/mydb?sslmode=disable",
			expectedKey:     "DB_DSN",
			expectedDefault: "postgresql://user:pass@localhost:5432/mydb?sslmode=disable",
			explanation:     "CORRECT - DSN is default value, key has no colons",
		},
		{
			scenario:        "Vault path simple",
			input:           "secret/data/myapp/db-password",
			expectedKey:     "secret/data/myapp/db-password",
			expectedDefault: "",
			explanation:     "CORRECT - no colons in path",
		},
		{
			scenario:        "Vault path with colons - no quotes",
			input:           "secret/data/myapp:db-password",
			expectedKey:     "secret/data/myapp",
			expectedDefault: "db-password",
			explanation:     "AMBIGUOUS - could be key or key:default",
		},
		{
			scenario:        "Vault path with colons - with quotes",
			input:           `'secret/data/myapp:db-password'`,
			expectedKey:     "secret/data/myapp:db-password",
			expectedDefault: "",
			explanation:     "CORRECT - single quotes preserve path with colons",
		},
		{
			scenario:        "K8s resource path",
			input:           "configmap/app-config/db-host",
			expectedKey:     "configmap/app-config/db-host",
			expectedDefault: "",
			explanation:     "CORRECT - no colons in path",
		},
		{
			scenario:        "Config key nested with dots",
			input:           "database.primary.host",
			expectedKey:     "database.primary.host",
			expectedDefault: "",
			explanation:     "CORRECT - dots are fine, no colons",
		},
		{
			scenario:        "Config key with colon - needs quotes",
			input:           `'db:primary:host'`,
			expectedKey:     "db:primary:host",
			expectedDefault: "",
			explanation:     "CORRECT - single quotes for keys with colons",
		},
	}

	for _, tt := range tests {
		t.Run(tt.scenario, func(t *testing.T) {
			key, defaultValue := internal.ParseKeyDefault(tt.input)

			if key != tt.expectedKey {
				t.Errorf("%s\nInput: %q\nKey: got %q, want %q\nExplanation: %s",
					tt.scenario, tt.input, key, tt.expectedKey, tt.explanation)
			}

			if defaultValue != tt.expectedDefault {
				t.Errorf("%s\nInput: %q\nDefault: got %q, want %q\nExplanation: %s",
					tt.scenario, tt.input, defaultValue, tt.expectedDefault, tt.explanation)
			}
		})
	}
}
