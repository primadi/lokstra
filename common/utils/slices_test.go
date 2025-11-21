package utils

import (
	"testing"
)

func TestToAnySlice(t *testing.T) {
	t.Run("Convert string slice to any slice", func(t *testing.T) {
		input := []string{"hello", "world", "test"}
		result := ToAnySlice(input)

		if len(result) != 3 {
			t.Errorf("len(result) = %d, want 3", len(result))
		}
		if result[0] != "hello" {
			t.Errorf("result[0] = %v, want 'hello'", result[0])
		}
		if result[1] != "world" {
			t.Errorf("result[1] = %v, want 'world'", result[1])
		}
		if result[2] != "test" {
			t.Errorf("result[2] = %v, want 'test'", result[2])
		}
	})

	t.Run("Convert int slice to any slice", func(t *testing.T) {
		input := []int{1, 2, 3, 4, 5}
		result := ToAnySlice(input)

		if len(result) != 5 {
			t.Errorf("len(result) = %d, want 5", len(result))
		}
		for i, v := range input {
			if result[i] != v {
				t.Errorf("result[%d] = %v, want %v", i, result[i], v)
			}
		}
	})

	t.Run("Convert empty slice", func(t *testing.T) {
		input := []string{}
		result := ToAnySlice(input)

		if len(result) != 0 {
			t.Errorf("len(result) = %d, want 0", len(result))
		}
		if result == nil {
			t.Error("result should not be nil")
		}
	})

	t.Run("Convert nil slice", func(t *testing.T) {
		var input []string
		result := ToAnySlice(input)

		if len(result) != 0 {
			t.Errorf("len(result) = %d, want 0", len(result))
		}
		if result == nil {
			t.Error("result should not be nil")
		}
	})

	t.Run("Convert bool slice", func(t *testing.T) {
		input := []bool{true, false, true}
		result := ToAnySlice(input)

		if len(result) != 3 {
			t.Errorf("len(result) = %d, want 3", len(result))
		}
		if result[0] != true {
			t.Errorf("result[0] = %v, want true", result[0])
		}
		if result[1] != false {
			t.Errorf("result[1] = %v, want false", result[1])
		}
		if result[2] != true {
			t.Errorf("result[2] = %v, want true", result[2])
		}
	})

	t.Run("Convert float slice", func(t *testing.T) {
		input := []float64{1.1, 2.2, 3.3}
		result := ToAnySlice(input)

		if len(result) != 3 {
			t.Errorf("len(result) = %d, want 3", len(result))
		}
		if result[0] != 1.1 {
			t.Errorf("result[0] = %v, want 1.1", result[0])
		}
		if result[1] != 2.2 {
			t.Errorf("result[1] = %v, want 2.2", result[1])
		}
		if result[2] != 3.3 {
			t.Errorf("result[2] = %v, want 3.3", result[2])
		}
	})

	t.Run("Convert custom struct slice", func(t *testing.T) {
		type Person struct {
			Name string
			Age  int
		}

		input := []Person{
			{Name: "John", Age: 30},
			{Name: "Jane", Age: 25},
		}
		result := ToAnySlice(input)

		if len(result) != 2 {
			t.Errorf("len(result) = %d, want 2", len(result))
		}
		if result[0] != (Person{Name: "John", Age: 30}) {
			t.Errorf("result[0] = %v, want %v", result[0], Person{Name: "John", Age: 30})
		}
		if result[1] != (Person{Name: "Jane", Age: 25}) {
			t.Errorf("result[1] = %v, want %v", result[1], Person{Name: "Jane", Age: 25})
		}
	})

	t.Run("Convert interface slice", func(t *testing.T) {
		input := []any{"string", 42, true}
		result := ToAnySlice(input)

		if len(result) != 3 {
			t.Errorf("len(result) = %d, want 3", len(result))
		}
		if result[0] != "string" {
			t.Errorf("result[0] = %v, want 'string'", result[0])
		}
		if result[1] != 42 {
			t.Errorf("result[1] = %v, want 42", result[1])
		}
		if result[2] != true {
			t.Errorf("result[2] = %v, want true", result[2])
		}
	})
}

func TestSlicesConcat(t *testing.T) {
	t.Run("Concatenate string slices", func(t *testing.T) {
		slice1 := []string{"a", "b"}
		slice2 := []string{"c", "d"}
		slice3 := []string{"e"}

		result := SlicesConcat(slice1, slice2, slice3)

		expected := []string{"a", "b", "c", "d", "e"}
		if len(result) != len(expected) {
			t.Errorf("len(result) = %d, want %d", len(result), len(expected))
		}
		for i, v := range expected {
			if result[i] != v {
				t.Errorf("result[%d] = %v, want %v", i, result[i], v)
			}
		}
	})

	t.Run("Concatenate int slices", func(t *testing.T) {
		slice1 := []int{1, 2}
		slice2 := []int{3, 4, 5}
		slice3 := []int{6}

		result := SlicesConcat(slice1, slice2, slice3)

		expected := []int{1, 2, 3, 4, 5, 6}
		if len(result) != len(expected) {
			t.Errorf("len(result) = %d, want %d", len(result), len(expected))
		}
		for i, v := range expected {
			if result[i] != v {
				t.Errorf("result[%d] = %v, want %v", i, result[i], v)
			}
		}
	})

	t.Run("Concatenate with empty slice", func(t *testing.T) {
		slice1 := []string{"a", "b"}
		slice2 := []string{}
		slice3 := []string{"c"}

		result := SlicesConcat(slice1, slice2, slice3)

		expected := []string{"a", "b", "c"}
		if len(result) != len(expected) {
			t.Errorf("len(result) = %d, want %d", len(result), len(expected))
		}
		for i, v := range expected {
			if result[i] != v {
				t.Errorf("result[%d] = %v, want %v", i, result[i], v)
			}
		}
	})

	t.Run("Concatenate all empty slices", func(t *testing.T) {
		slice1 := []string{}
		slice2 := []string{}
		slice3 := []string{}

		result := SlicesConcat(slice1, slice2, slice3)

		if result == nil {
			t.Error("result should not be nil")
		}
		if len(result) != 0 {
			t.Errorf("len(result) = %d, want 0", len(result))
		}
	})

	t.Run("Concatenate single slice", func(t *testing.T) {
		slice1 := []string{"single", "slice"}

		result := SlicesConcat(slice1)

		expected := []string{"single", "slice"}
		if len(result) != len(expected) {
			t.Errorf("len(result) = %d, want %d", len(result), len(expected))
		}
		for i, v := range expected {
			if result[i] != v {
				t.Errorf("result[%d] = %v, want %v", i, result[i], v)
			}
		}
	})

	t.Run("Concatenate no slices", func(t *testing.T) {
		result := SlicesConcat[string]()

		if result != nil {
			t.Error("result should be nil")
		}
	})

	t.Run("Concatenate with nil slice", func(t *testing.T) {
		slice1 := []string{"a", "b"}
		var slice2 []string // nil slice
		slice3 := []string{"c"}

		result := SlicesConcat(slice1, slice2, slice3)

		expected := []string{"a", "b", "c"}
		if len(result) != len(expected) {
			t.Errorf("len(result) = %d, want %d", len(result), len(expected))
		}
		for i, v := range expected {
			if result[i] != v {
				t.Errorf("result[%d] = %v, want %v", i, result[i], v)
			}
		}
	})

	t.Run("Concatenate large slices", func(t *testing.T) {
		// Create large slices for performance testing
		slice1 := make([]int, 1000)
		slice2 := make([]int, 1000)

		for i := 0; i < 1000; i++ {
			slice1[i] = i
			slice2[i] = i + 1000
		}

		result := SlicesConcat(slice1, slice2)

		if len(result) != 2000 {
			t.Errorf("len(result) = %d, want 2000", len(result))
		}
		if result[0] != 0 {
			t.Errorf("result[0] = %d, want 0", result[0])
		}
		if result[999] != 999 {
			t.Errorf("result[999] = %d, want 999", result[999])
		}
		if result[1000] != 1000 {
			t.Errorf("result[1000] = %d, want 1000", result[1000])
		}
		if result[1999] != 1999 {
			t.Errorf("result[1999] = %d, want 1999", result[1999])
		}
	})

	t.Run("Concatenate custom struct slices", func(t *testing.T) {
		type Person struct {
			Name string
			Age  int
		}

		slice1 := []Person{{Name: "John", Age: 30}}
		slice2 := []Person{{Name: "Jane", Age: 25}, {Name: "Bob", Age: 35}}

		result := SlicesConcat(slice1, slice2)

		expected := []Person{
			{Name: "John", Age: 30},
			{Name: "Jane", Age: 25},
			{Name: "Bob", Age: 35},
		}
		if len(result) != len(expected) {
			t.Errorf("len(result) = %d, want %d", len(result), len(expected))
		}
		for i, v := range expected {
			if result[i] != v {
				t.Errorf("result[%d] = %v, want %v", i, result[i], v)
			}
		}
	})

	t.Run("Concatenate many slices", func(t *testing.T) {
		slices := make([][]int, 10)
		expectedTotal := 0

		for i := 0; i < 10; i++ {
			slices[i] = []int{i, i + 10}
			expectedTotal += 2
		}

		result := SlicesConcat(slices...)

		if len(result) != expectedTotal {
			t.Errorf("len(result) = %d, want %d", len(result), expectedTotal)
		}
		if result[0] != 0 {
			t.Errorf("result[0] = %d, want 0", result[0])
		}
		if result[1] != 10 {
			t.Errorf("result[1] = %d, want 10", result[1])
		}
		if result[2] != 1 {
			t.Errorf("result[2] = %d, want 1", result[2])
		}
		if result[3] != 11 {
			t.Errorf("result[3] = %d, want 11", result[3])
		}
	})
}

func BenchmarkToAnySlice(b *testing.B) {
	slice := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		slice[i] = "item"
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ToAnySlice(slice)
	}
}

func BenchmarkSlicesConcat(b *testing.B) {
	slice1 := make([]int, 500)
	slice2 := make([]int, 500)

	for i := 0; i < 500; i++ {
		slice1[i] = i
		slice2[i] = i + 500
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SlicesConcat(slice1, slice2)
	}
}
