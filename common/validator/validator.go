package validator

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/primadi/lokstra/core/response/api_formatter"
)

// ValidatorFunc is a function that validates a field value
// Returns error if validation fails, nil if valid
type ValidatorFunc func(fieldName string, fieldValue reflect.Value, ruleValue string) error

var (
	// validatorRegistry stores registered validator functions
	validatorRegistry     = make(map[string]ValidatorFunc)
	validatorRegistryLock sync.RWMutex

	validatorMetaCache sync.Map // map[reflect.Type]*validatorMeta
)

// RegisterValidator registers a custom validator function
// name: validator name (e.g., "uuid", "url")
// fn: validator function
func RegisterValidator(name string, fn ValidatorFunc) {
	validatorRegistryLock.Lock()
	defer validatorRegistryLock.Unlock()
	validatorRegistry[name] = fn
}

// getValidator retrieves a validator function by name
func getValidator(name string) (ValidatorFunc, bool) {
	validatorRegistryLock.RLock()
	defer validatorRegistryLock.RUnlock()
	fn, ok := validatorRegistry[name]
	return fn, ok
}

func init() {
	// Register built-in validators
	RegisterValidator("required", validateRequired)
	RegisterValidator("email", validateEmail)
	RegisterValidator("min", validateMin)
	RegisterValidator("max", validateMax)
	RegisterValidator("gt", validateGt)
	RegisterValidator("gte", validateGte)
	RegisterValidator("lt", validateLt)
	RegisterValidator("lte", validateLte)
	RegisterValidator("oneof", validateOneOf)
}

// validatorFieldMeta contains cached metadata for a single field
type validatorFieldMeta struct {
	FieldIndex []int  // Field index path for nested structs
	FieldName  string // Name to use in error messages (from json tag or field name)
	Rules      []validationRule
}

// validationRule represents a single validation rule
type validationRule struct {
	Name  string // e.g., "required", "min", "email"
	Value string // e.g., "5" for min=5
}

// validatorMeta contains cached validation metadata for a struct type
type validatorMeta struct {
	Type   reflect.Type
	Fields []validatorFieldMeta
}

// getOrBuildValidatorMeta gets cached metadata or builds it
func getOrBuildValidatorMeta(t reflect.Type) *validatorMeta {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return &validatorMeta{Type: t, Fields: nil}
	}

	// Check cache
	if meta, ok := validatorMetaCache.Load(t); ok {
		return meta.(*validatorMeta)
	}

	// Build new metadata
	meta := &validatorMeta{
		Type:   t,
		Fields: make([]validatorFieldMeta, 0),
	}

	numField := t.NumField()
	for i := 0; i < numField; i++ {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get validate tag
		validateTag := field.Tag.Get("validate")
		if validateTag == "" {
			continue
		}

		// Get field name for error message (prefer json tag)
		fieldName := field.Name
		if jsonTag := field.Tag.Get("json"); jsonTag != "" {
			jsonTag = strings.Split(jsonTag, ",")[0] // Remove options like omitempty
			if jsonTag != "" && jsonTag != "-" {
				fieldName = jsonTag
			}
		}

		// Parse validation rules
		rules := parseValidationRules(validateTag)
		if len(rules) == 0 {
			continue
		}

		meta.Fields = append(meta.Fields, validatorFieldMeta{
			FieldIndex: field.Index,
			FieldName:  fieldName,
			Rules:      rules,
		})
	}

	// Store in cache
	validatorMetaCache.Store(t, meta)

	return meta
}

// parseValidationRules parses a validate tag into rules
func parseValidationRules(validateTag string) []validationRule {
	if validateTag == "" {
		return nil
	}

	ruleParts := strings.Split(validateTag, ",")
	rules := make([]validationRule, 0, len(ruleParts))

	for _, rulePart := range ruleParts {
		rulePart = strings.TrimSpace(rulePart)
		if rulePart == "" {
			continue
		}

		parts := strings.SplitN(rulePart, "=", 2)
		rule := validationRule{
			Name: parts[0],
		}
		if len(parts) > 1 {
			rule.Value = parts[1]
		}

		rules = append(rules, rule)
	}

	return rules
}

// ValidateStruct validates a struct based on validate tags
// Returns a slice of FieldError if validation fails, or error if there's a system error
// Uses cached metadata for performance
func ValidateStruct(structData any) ([]api_formatter.FieldError, error) {
	if structData == nil {
		return nil, fmt.Errorf("structData cannot be nil")
	}

	val := reflect.ValueOf(structData)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil, fmt.Errorf("structData pointer cannot be nil")
		}
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("structData must be a struct or pointer to struct")
	}

	// Get or build cached metadata
	meta := getOrBuildValidatorMeta(val.Type())

	var fieldErrors []api_formatter.FieldError

	// Validate each field using cached metadata
	for _, fieldMeta := range meta.Fields {
		fieldValue := val.FieldByIndex(fieldMeta.FieldIndex)

		// Validate all rules for this field
		for _, rule := range fieldMeta.Rules {
			err := validateRule(fieldMeta.FieldName, fieldValue, rule)
			if err != nil {
				fieldErrors = append(fieldErrors, api_formatter.FieldError{
					Field:   fieldMeta.FieldName,
					Message: err.Error(),
				})
				break // Stop at first error for this field
			}
		}
	}

	return fieldErrors, nil
}

func validateRule(fieldName string, fieldValue reflect.Value, rule validationRule) error {
	// Handle pointer fields
	if fieldValue.Kind() == reflect.Ptr {
		if fieldValue.IsNil() {
			// Check if required
			if rule.Name == "required" {
				return fmt.Errorf("%s is required", fieldName)
			}
			// For optional pointer fields, skip validation if nil
			if rule.Name == "omitempty" {
				return nil
			}
			// Skip validation for nil pointers unless required
			return nil
		}
		fieldValue = fieldValue.Elem()
	}

	// Special handling for omitempty
	if rule.Name == "omitempty" {
		return nil
	}

	// Look up validator function from registry
	validatorFn, ok := getValidator(rule.Name)
	if !ok {
		// Unknown validator - skip (could log warning in production)
		return nil
	}

	// Call the validator function
	return validatorFn(fieldName, fieldValue, rule.Value)
}

func validateRequired(fieldName string, fieldValue reflect.Value, ruleValue string) error {
	switch fieldValue.Kind() {
	case reflect.String:
		if fieldValue.String() == "" {
			return fmt.Errorf("%s is required", fieldName)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// For numbers, 0 is considered empty for required check (you might want to change this)
		if fieldValue.Int() == 0 {
			return fmt.Errorf("%s is required", fieldName)
		}
	case reflect.Float32, reflect.Float64:
		if fieldValue.Float() == 0 {
			return fmt.Errorf("%s is required", fieldName)
		}
	case reflect.Bool:
		// Bool is always valid for required
	case reflect.Slice, reflect.Map, reflect.Array:
		if fieldValue.Len() == 0 {
			return fmt.Errorf("%s is required", fieldName)
		}
	}
	return nil
}

func validateEmail(fieldName string, fieldValue reflect.Value, ruleValue string) error {
	if fieldValue.Kind() != reflect.String {
		return nil
	}

	email := fieldValue.String()
	if email == "" {
		return nil // Use required tag to check for empty
	}

	// Simple email validation
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return fmt.Errorf("%s must be a valid email address", fieldName)
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return fmt.Errorf("%s must be a valid email address", fieldName)
	}

	return nil
}

func validateMin(fieldName string, fieldValue reflect.Value, ruleValue string) error {
	minVal, err := strconv.ParseFloat(ruleValue, 64)
	if err != nil {
		return nil // Invalid rule value, skip
	}

	switch fieldValue.Kind() {
	case reflect.String:
		if float64(len(fieldValue.String())) < minVal {
			return fmt.Errorf("%s must be at least %.0f characters", fieldName, minVal)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if float64(fieldValue.Int()) < minVal {
			return fmt.Errorf("%s must be at least %.0f", fieldName, minVal)
		}
	case reflect.Float32, reflect.Float64:
		if fieldValue.Float() < minVal {
			return fmt.Errorf("%s must be at least %.2f", fieldName, minVal)
		}
	case reflect.Slice, reflect.Map, reflect.Array:
		if float64(fieldValue.Len()) < minVal {
			return fmt.Errorf("%s must have at least %.0f items", fieldName, minVal)
		}
	}
	return nil
}

func validateMax(fieldName string, fieldValue reflect.Value, ruleValue string) error {
	maxVal, err := strconv.ParseFloat(ruleValue, 64)
	if err != nil {
		return nil
	}

	switch fieldValue.Kind() {
	case reflect.String:
		if float64(len(fieldValue.String())) > maxVal {
			return fmt.Errorf("%s must be at most %.0f characters", fieldName, maxVal)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if float64(fieldValue.Int()) > maxVal {
			return fmt.Errorf("%s must be at most %.0f", fieldName, maxVal)
		}
	case reflect.Float32, reflect.Float64:
		if fieldValue.Float() > maxVal {
			return fmt.Errorf("%s must be at most %.2f", fieldName, maxVal)
		}
	case reflect.Slice, reflect.Map, reflect.Array:
		if float64(fieldValue.Len()) > maxVal {
			return fmt.Errorf("%s must have at most %.0f items", fieldName, maxVal)
		}
	}
	return nil
}

func validateGt(fieldName string, fieldValue reflect.Value, ruleValue string) error {
	gtVal, err := strconv.ParseFloat(ruleValue, 64)
	if err != nil {
		return nil
	}

	switch fieldValue.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if float64(fieldValue.Int()) <= gtVal {
			return fmt.Errorf("%s must be greater than %.0f", fieldName, gtVal)
		}
	case reflect.Float32, reflect.Float64:
		if fieldValue.Float() <= gtVal {
			return fmt.Errorf("%s must be greater than %.2f", fieldName, gtVal)
		}
	}
	return nil
}

func validateGte(fieldName string, fieldValue reflect.Value, ruleValue string) error {
	gteVal, err := strconv.ParseFloat(ruleValue, 64)
	if err != nil {
		return nil
	}

	switch fieldValue.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if float64(fieldValue.Int()) < gteVal {
			return fmt.Errorf("%s must be greater than or equal to %.0f", fieldName, gteVal)
		}
	case reflect.Float32, reflect.Float64:
		if fieldValue.Float() < gteVal {
			return fmt.Errorf("%s must be greater than or equal to %.2f", fieldName, gteVal)
		}
	}
	return nil
}

func validateLt(fieldName string, fieldValue reflect.Value, ruleValue string) error {
	ltVal, err := strconv.ParseFloat(ruleValue, 64)
	if err != nil {
		return nil
	}

	switch fieldValue.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if float64(fieldValue.Int()) >= ltVal {
			return fmt.Errorf("%s must be less than %.0f", fieldName, ltVal)
		}
	case reflect.Float32, reflect.Float64:
		if fieldValue.Float() >= ltVal {
			return fmt.Errorf("%s must be less than %.2f", fieldName, ltVal)
		}
	}
	return nil
}

func validateLte(fieldName string, fieldValue reflect.Value, ruleValue string) error {
	lteVal, err := strconv.ParseFloat(ruleValue, 64)
	if err != nil {
		return nil
	}

	switch fieldValue.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if float64(fieldValue.Int()) > lteVal {
			return fmt.Errorf("%s must be less than or equal to %.0f", fieldName, lteVal)
		}
	case reflect.Float32, reflect.Float64:
		if fieldValue.Float() > lteVal {
			return fmt.Errorf("%s must be less than or equal to %.2f", fieldName, lteVal)
		}
	}
	return nil
}

func validateOneOf(fieldName string, fieldValue reflect.Value, ruleValue string) error {
	if fieldValue.Kind() != reflect.String {
		return nil
	}

	value := fieldValue.String()
	if value == "" {
		return nil // Use required tag to check for empty
	}

	validValues := strings.Fields(ruleValue)
	for _, validValue := range validValues {
		if value == validValue {
			return nil
		}
	}

	return fmt.Errorf("%s must be one of: %s", fieldName, strings.Join(validValues, ", "))
}
