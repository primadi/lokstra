package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToAnySlice(t *testing.T) {
	t.Run("Convert string slice to any slice", func(t *testing.T) {
		input := []string{"hello", "world", "test"}
		result := ToAnySlice(input)

		assert.Len(t, result, 3)
		assert.Equal(t, "hello", result[0])
		assert.Equal(t, "world", result[1])
		assert.Equal(t, "test", result[2])
	})

	t.Run("Convert int slice to any slice", func(t *testing.T) {
		input := []int{1, 2, 3, 4, 5}
		result := ToAnySlice(input)

		assert.Len(t, result, 5)
		for i, v := range input {
			assert.Equal(t, v, result[i])
		}
	})

	t.Run("Convert empty slice", func(t *testing.T) {
		input := []string{}
		result := ToAnySlice(input)

		assert.Len(t, result, 0)
		assert.NotNil(t, result)
	})

	t.Run("Convert nil slice", func(t *testing.T) {
		var input []string
		result := ToAnySlice(input)

		assert.Len(t, result, 0)
		assert.NotNil(t, result)
	})

	t.Run("Convert bool slice", func(t *testing.T) {
		input := []bool{true, false, true}
		result := ToAnySlice(input)

		assert.Len(t, result, 3)
		assert.Equal(t, true, result[0])
		assert.Equal(t, false, result[1])
		assert.Equal(t, true, result[2])
	})

	t.Run("Convert float slice", func(t *testing.T) {
		input := []float64{1.1, 2.2, 3.3}
		result := ToAnySlice(input)

		assert.Len(t, result, 3)
		assert.Equal(t, 1.1, result[0])
		assert.Equal(t, 2.2, result[1])
		assert.Equal(t, 3.3, result[2])
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

		assert.Len(t, result, 2)
		assert.Equal(t, Person{Name: "John", Age: 30}, result[0])
		assert.Equal(t, Person{Name: "Jane", Age: 25}, result[1])
	})

	t.Run("Convert interface slice", func(t *testing.T) {
		input := []any{"string", 42, true}
		result := ToAnySlice(input)

		assert.Len(t, result, 3)
		assert.Equal(t, "string", result[0])
		assert.Equal(t, 42, result[1])
		assert.Equal(t, true, result[2])
	})
}

func TestSlicesConcat(t *testing.T) {
	t.Run("Concatenate string slices", func(t *testing.T) {
		slice1 := []string{"a", "b"}
		slice2 := []string{"c", "d"}
		slice3 := []string{"e"}

		result := SlicesConcat(slice1, slice2, slice3)

		expected := []string{"a", "b", "c", "d", "e"}
		assert.Equal(t, expected, result)
	})

	t.Run("Concatenate int slices", func(t *testing.T) {
		slice1 := []int{1, 2}
		slice2 := []int{3, 4, 5}
		slice3 := []int{6}

		result := SlicesConcat(slice1, slice2, slice3)

		expected := []int{1, 2, 3, 4, 5, 6}
		assert.Equal(t, expected, result)
	})

	t.Run("Concatenate with empty slice", func(t *testing.T) {
		slice1 := []string{"a", "b"}
		slice2 := []string{}
		slice3 := []string{"c"}

		result := SlicesConcat(slice1, slice2, slice3)

		expected := []string{"a", "b", "c"}
		assert.Equal(t, expected, result)
	})

	t.Run("Concatenate all empty slices", func(t *testing.T) {
		slice1 := []string{}
		slice2 := []string{}
		slice3 := []string{}

		result := SlicesConcat(slice1, slice2, slice3)

		assert.NotNil(t, result)
		assert.Len(t, result, 0)
	})

	t.Run("Concatenate single slice", func(t *testing.T) {
		slice1 := []string{"single", "slice"}

		result := SlicesConcat(slice1)

		expected := []string{"single", "slice"}
		assert.Equal(t, expected, result)
	})

	t.Run("Concatenate no slices", func(t *testing.T) {
		result := SlicesConcat[string]()

		assert.Nil(t, result)
	})

	t.Run("Concatenate with nil slice", func(t *testing.T) {
		slice1 := []string{"a", "b"}
		var slice2 []string // nil slice
		slice3 := []string{"c"}

		result := SlicesConcat(slice1, slice2, slice3)

		expected := []string{"a", "b", "c"}
		assert.Equal(t, expected, result)
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

		assert.Len(t, result, 2000)
		assert.Equal(t, 0, result[0])
		assert.Equal(t, 999, result[999])
		assert.Equal(t, 1000, result[1000])
		assert.Equal(t, 1999, result[1999])
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
		assert.Equal(t, expected, result)
	})

	t.Run("Concatenate many slices", func(t *testing.T) {
		slices := make([][]int, 10)
		expectedTotal := 0

		for i := 0; i < 10; i++ {
			slices[i] = []int{i, i + 10}
			expectedTotal += 2
		}

		result := SlicesConcat(slices...)

		assert.Len(t, result, expectedTotal)
		assert.Equal(t, 0, result[0])
		assert.Equal(t, 10, result[1])
		assert.Equal(t, 1, result[2])
		assert.Equal(t, 11, result[3])
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
