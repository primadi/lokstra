package validator

import (
	"sync"
	"testing"
)

// Benchmark to verify cache performance - simulates first validation (cold cache)
// This benchmark measures the cost of building metadata + validation
func BenchmarkValidateStruct_FirstCall(b *testing.B) {
	type TestStruct struct {
		Name  string `json:"name" validate:"required,min=3,max=50"`
		Email string `json:"email" validate:"required,email"`
		Age   int    `json:"age" validate:"min=1,max=120"`
		Role  string `json:"role" validate:"oneof=admin user guest"`
	}

	for b.Loop() {
		// Clear cache to simulate first call by creating new sync.Map
		// sync.Map doesn't have Clear() method, so we recreate it
		validatorMetaCache = sync.Map{}

		valid := TestStruct{
			Name:  "John Doe",
			Email: "john@example.com",
			Age:   25,
			Role:  "user",
		}
		ValidateStruct(&valid)
	}
}

func BenchmarkValidateStruct_CachedCall(b *testing.B) {
	type TestStruct struct {
		Name  string `json:"name" validate:"required,min=3,max=50"`
		Email string `json:"email" validate:"required,email"`
		Age   int    `json:"age" validate:"min=1,max=120"`
		Role  string `json:"role" validate:"oneof=admin user guest"`
	}

	// Warm up cache
	valid := TestStruct{
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   25,
		Role:  "user",
	}
	ValidateStruct(&valid)

	for b.Loop() {
		valid := TestStruct{
			Name:  "John Doe",
			Email: "john@example.com",
			Age:   25,
			Role:  "user",
		}
		ValidateStruct(&valid)
	}
}

func BenchmarkValidateStruct_Invalid(b *testing.B) {
	type TestStruct struct {
		Name  string `json:"name" validate:"required,min=3,max=50"`
		Email string `json:"email" validate:"required,email"`
		Age   int    `json:"age" validate:"min=1,max=120"`
		Role  string `json:"role" validate:"oneof=admin user guest"`
	}

	// Warm up cache
	valid := TestStruct{Name: "John", Email: "test@test.com", Age: 25, Role: "user"}
	ValidateStruct(&valid)

	for b.Loop() {
		invalid := TestStruct{
			Name:  "AB", // Too short
			Email: "not-an-email",
			Age:   150, // Too high
			Role:  "superuser",
		}
		ValidateStruct(&invalid)
	}
}

func BenchmarkValidateStruct_Complex(b *testing.B) {
	type TestStruct struct {
		Name     string  `json:"name" validate:"required,min=3,max=50"`
		Email    string  `json:"email" validate:"required,email"`
		Age      int     `json:"age" validate:"min=18,max=120"`
		Role     string  `json:"role" validate:"oneof=admin user guest"`
		Salary   float64 `json:"salary" validate:"gt=0"`
		Active   bool    `json:"active"`
		Phone    *string `json:"phone" validate:"omitempty,min=10"`
		Bio      *string `json:"bio" validate:"omitempty,max=500"`
		Tags     []string
		Metadata map[string]string
	}

	phone := "1234567890"
	bio := "Software engineer"

	// Warm up cache
	valid := TestStruct{
		Name:   "John Doe",
		Email:  "john@example.com",
		Age:    30,
		Role:   "user",
		Salary: 50000,
		Active: true,
		Phone:  &phone,
		Bio:    &bio,
	}
	ValidateStruct(&valid)

	for b.Loop() {
		valid := TestStruct{
			Name:   "John Doe",
			Email:  "john@example.com",
			Age:    30,
			Role:   "user",
			Salary: 50000,
			Active: true,
			Phone:  &phone,
			Bio:    &bio,
		}
		ValidateStruct(&valid)
	}
}

// Benchmark to simulate concurrent request validation (realistic scenario)
func BenchmarkValidateStruct_Concurrent(b *testing.B) {
	type TestStruct struct {
		Name  string `json:"name" validate:"required,min=3,max=50"`
		Email string `json:"email" validate:"required,email"`
		Age   int    `json:"age" validate:"min=1,max=120"`
		Role  string `json:"role" validate:"oneof=admin user guest"`
	}

	// Warm up cache
	valid := TestStruct{Name: "John", Email: "test@test.com", Age: 25, Role: "user"}
	ValidateStruct(&valid)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			valid := TestStruct{
				Name:  "John Doe",
				Email: "john@example.com",
				Age:   25,
				Role:  "user",
			}
			ValidateStruct(&valid)
		}
	})
}
