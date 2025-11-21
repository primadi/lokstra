package utils

import (
	"testing"
)

func TestCamelToSnake(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Basic cases
		{
			name:     "Simple camelCase",
			input:    "camelCase",
			expected: "camel_case",
		},
		{
			name:     "PascalCase",
			input:    "PascalCase",
			expected: "pascal_case",
		},
		{
			name:     "Single word lowercase",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "Single word uppercase",
			input:    "HELLO",
			expected: "hello",
		},
		{
			name:     "Single character lowercase",
			input:    "a",
			expected: "a",
		},
		{
			name:     "Single character uppercase",
			input:    "A",
			expected: "a",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},

		// Multiple words
		{
			name:     "Three words camelCase",
			input:    "getUserName",
			expected: "get_user_name",
		},
		{
			name:     "Three words PascalCase",
			input:    "GetUserName",
			expected: "get_user_name",
		},
		{
			name:     "Four words",
			input:    "getUserFullName",
			expected: "get_user_full_name",
		},

		// Acronyms and abbreviations
		{
			name:     "Acronym at start",
			input:    "HTTPResponse",
			expected: "http_response",
		},
		{
			name:     "Acronym in middle",
			input:    "parseHTTPResponse",
			expected: "parse_http_response",
		},
		{
			name:     "Acronym at end",
			input:    "responseHTTP",
			expected: "response_http",
		},
		{
			name:     "Multiple acronyms",
			input:    "HTTPSProxy",
			expected: "https_proxy",
		},
		{
			name:     "Long acronym",
			input:    "XMLHttpRequest",
			expected: "xml_http_request",
		},

		// Numbers
		{
			name:     "Numbers in middle",
			input:    "user2Name",
			expected: "user2_name",
		},
		{
			name:     "Numbers at end",
			input:    "userName2",
			expected: "user_name2",
		},
		{
			name:     "Numbers at start",
			input:    "2userName",
			expected: "2user_name",
		},
		{
			name:     "Multiple numbers",
			input:    "user2Name3",
			expected: "user2_name3",
		},

		// Edge cases with consecutive capitals
		{
			name:     "Consecutive capitals at start",
			input:    "HTMLParser",
			expected: "html_parser",
		},
		{
			name:     "Consecutive capitals in middle",
			input:    "parseHTMLDocument",
			expected: "parse_html_document",
		},
		{
			name:     "Consecutive capitals at end",
			input:    "documentHTML",
			expected: "document_html",
		},
		{
			name:     "All uppercase",
			input:    "ALLUPPERCASE",
			expected: "alluppercase",
		},

		// Special patterns
		{
			name:     "ID suffix",
			input:    "userID",
			expected: "user_id",
		},
		{
			name:     "URL pattern",
			input:    "baseURL",
			expected: "base_url",
		},
		{
			name:     "API pattern",
			input:    "restAPI",
			expected: "rest_api",
		},
		{
			name:     "Mixed pattern",
			input:    "getUserIDFromAPI",
			expected: "get_user_id_from_api",
		},

		// Real-world examples
		{
			name:     "Database field",
			input:    "createdAt",
			expected: "created_at",
		},
		{
			name:     "Database field with ID",
			input:    "userAccountID",
			expected: "user_account_id",
		},
		{
			name:     "Service method",
			input:    "validateUserCredentials",
			expected: "validate_user_credentials",
		},
		{
			name:     "Configuration key",
			input:    "maxConnectionPoolSize",
			expected: "max_connection_pool_size",
		},
		{
			name:     "Event handler",
			input:    "onUserLoginSuccess",
			expected: "on_user_login_success",
		},

		// Complex technical terms
		{
			name:     "OAuth pattern",
			input:    "oAuthToken",
			expected: "o_auth_token",
		},
		{
			name:     "JSON pattern",
			input:    "parseJSONData",
			expected: "parse_json_data",
		},
		{
			name:     "SQL pattern",
			input:    "executeSQLQuery",
			expected: "execute_sql_query",
		},
		{
			name:     "TCP/IP pattern",
			input:    "tcpIPConnection",
			expected: "tcp_ip_connection",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CamelToSnake(tt.input)
			if result != tt.expected {
				t.Errorf("CamelToSnake(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCamelToSnake_EdgeCases(t *testing.T) {
	// Test specific edge cases that might break the algorithm
	edgeCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Single uppercase at start",
			input:    "A",
			expected: "a",
		},
		{
			name:     "Two characters - both uppercase",
			input:    "AB",
			expected: "ab",
		},
		{
			name:     "Two characters - camelCase",
			input:    "aB",
			expected: "a_b",
		},
		{
			name:     "Three characters - all uppercase",
			input:    "ABC",
			expected: "abc",
		},
		{
			name:     "Three characters - PascalCase end",
			input:    "AbC",
			expected: "ab_c",
		},
		{
			name:     "Alternating case",
			input:    "aBcDeFg",
			expected: "a_bc_de_fg",
		},
	}

	for _, tt := range edgeCases {
		t.Run(tt.name, func(t *testing.T) {
			result := CamelToSnake(tt.input)
			if result != tt.expected {
				t.Errorf("CamelToSnake(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCamelToSnake_HelperFunctions(t *testing.T) {
	// Test helper functions individually
	t.Run("isUpper function", func(t *testing.T) {
		if !isUpper('A') {
			t.Error("isUpper('A') should be true")
		}
		if !isUpper('Z') {
			t.Error("isUpper('Z') should be true")
		}
		if isUpper('a') {
			t.Error("isUpper('a') should be false")
		}
		if isUpper('z') {
			t.Error("isUpper('z') should be false")
		}
		if isUpper('1') {
			t.Error("isUpper('1') should be false")
		}
		if isUpper(' ') {
			t.Error("isUpper(' ') should be false")
		}
		if isUpper('_') {
			t.Error("isUpper('_') should be false")
		}
	})

	t.Run("toLower function", func(t *testing.T) {
		if toLower('A') != 'a' {
			t.Errorf("toLower('A') = %q, want 'a'", toLower('A'))
		}
		if toLower('Z') != 'z' {
			t.Errorf("toLower('Z') = %q, want 'z'", toLower('Z'))
		}
		if toLower('a') != 'a' {
			t.Errorf("toLower('a') = %q, want 'a'", toLower('a'))
		}
		if toLower('z') != 'z' {
			t.Errorf("toLower('z') = %q, want 'z'", toLower('z'))
		}
		if toLower('1') != '1' {
			t.Errorf("toLower('1') = %q, want '1'", toLower('1'))
		}
		if toLower(' ') != ' ' {
			t.Errorf("toLower(' ') = %q, want ' '", toLower(' '))
		}
		if toLower('_') != '_' {
			t.Errorf("toLower('_') = %q, want '_'", toLower('_'))
		}
	})
}

func TestCamelToSnake_Performance(t *testing.T) {
	// Test with longer strings to ensure performance
	longCamelCase := "thisIsAVeryLongCamelCaseStringWithManyWordsToTestPerformanceOfTheConversionFunction"
	expected := "this_is_a_very_long_camel_case_string_with_many_words_to_test_performance_of_the_conversion_function"

	result := CamelToSnake(longCamelCase)
	if result != expected {
		t.Errorf("CamelToSnake(%q) = %q, want %q", longCamelCase, result, expected)
	}
}

func BenchmarkCamelToSnake(b *testing.B) {
	testCases := []string{
		"camelCase",
		"PascalCase",
		"getUserName",
		"HTTPResponse",
		"parseJSONData",
		"thisIsAVeryLongCamelCaseStringWithManyWords",
	}

	for _, tc := range testCases {
		b.Run(tc, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				CamelToSnake(tc)
			}
		})
	}
}

func ExampleCamelToSnake() {
	// Basic usage examples
	result1 := CamelToSnake("camelCase")
	result2 := CamelToSnake("PascalCase")
	result3 := CamelToSnake("getUserName")
	result4 := CamelToSnake("HTTPResponse")

	println(result1) // camel_case
	println(result2) // pascal_case
	println(result3) // get_user_name
	println(result4) // http_response
}
