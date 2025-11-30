package internal

import "strings"

// parses "key:default" format with quote escaping support
//
// Rules:
//  1. If key starts with `"`, parse until closing `"` - rest is default
//  2. If key does NOT start with `"`, FIRST `:` is separator
//
// This eliminates ambiguity when key contains ':' characters
//
// Examples WITHOUT quotes (`:` is separator):
//
//	"db.host" -> ("db.host", "")
//	"db.host:localhost" -> ("db.host", "localhost")
//	"secret/data/db:password" -> ("secret/data/db", "password")
//	"DB_URL:postgresql://localhost:5432/db" -> ("DB_URL", "postgresql://localhost:5432/db")
//
// Examples WITH quotes (`:` inside quotes is part of key):
//
//	`"db.host"` -> ("db.host", "")
//	`"db.host":localhost` -> ("db.host", "localhost")
//	`"secret/data/db:password"` -> ("secret/data/db:password", "")
//	`"secret/data/db:password":fallback` -> ("secret/data/db:password", "fallback")
//	`"arn:aws:secretsmanager:us-east-1:123:secret:db"` -> ("arn:aws:secretsmanager:us-east-1:123:secret:db", "")
//	`"db:url":postgresql://localhost:5432/db` -> ("db:url", "postgresql://localhost:5432/db")
//
// Strategy:
//   - Check if input starts with `"`
//   - If YES: find closing `"`, extract key, rest is default (after `:`)
//   - If NO: find FIRST `:`, split into key and default
func ParseKeyDefault(input string) (key string, defaultValue string) {
	if len(input) == 0 {
		return "", ""
	}

	// Strategy 1: Quoted key - parse until closing single quote
	if input[0] == '\'' {
		// Find closing single quote
		closeQuote := strings.Index(input[1:], "'")
		if closeQuote == -1 {
			// No closing quote - treat entire string as key (malformed)
			return input, ""
		}

		// Extract key (without quotes)
		key = input[1 : closeQuote+1]

		// Check if there's a default value after closing quote
		afterQuote := input[closeQuote+2:] // +2 to skip closing quote
		if len(afterQuote) > 0 && afterQuote[0] == ':' {
			// Has default value
			defaultValue = afterQuote[1:] // Skip the ':'
		}

		return key, defaultValue
	}

	// Strategy 2: Non-quoted key - FIRST ':' is separator
	firstColon := strings.Index(input, ":")

	if firstColon == -1 {
		// No ':' found - entire string is the key
		return input, ""
	}

	// Split at FIRST ':'
	key = input[:firstColon]
	defaultValue = input[firstColon+1:]

	return key, defaultValue
}
