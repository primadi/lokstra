package utils

import "slices"

func ToAnySlice[T any](input []T) []any {
	result := make([]any, len(input))
	for i, v := range input {
		result[i] = v
	}
	return result
}

func SlicesConcat[T any](s ...[]T) []T {
	if len(s) == 0 {
		return nil
	}

	totalLength := 0
	for _, slice := range s {
		totalLength += len(slice)
	}

	// if totalLength == 0 return empty slice (non-nil)
	if totalLength == 0 {
		return []T{}
	}

	return slices.Concat(s...)
}

func AppendSorted[T any](slice []T, value T, less func(a, b T) bool) []T {
	if len(slice) == 0 {
		return []T{value}
	}

	// Find the correct insertion point
	for i := range slice {
		if less(value, slice[i]) {
			slice = append(slice, value)
			copy(slice[i+1:], slice[i:])
			slice[i] = value
			return slice
		}
	}

	// Append to the end
	return append(slice, value)
}

func AppendSortedOptimize[T any](slice []T, value T, less func(a, b T) bool) []T {
	low, high := 0, len(slice)

	if high < 16 {
		return AppendSorted(slice, value, less)
	}

	// Binary search for the correct position
	for low < high {
		mid := (low + high) / 2
		if less(value, slice[mid]) {
			high = mid
		} else {
			low = mid + 1
		}
	}

	// Sisipkan value di posisi 'low'
	slice = append(slice, value)     // tambah ruang di akhir
	copy(slice[low+1:], slice[low:]) // geser elemen ke kanan
	slice[low] = value               // masukkan value

	return slice
}

// create a new slice by appending value(s) to existing slice
func NewSliceAndAppend[T any](slice []T, value ...T) []T {
	newSlice := make([]T, len(slice)+len(value))
	copy(newSlice, slice)
	copy(newSlice[len(slice):], value)
	return newSlice
}

// create a new slice by concatenating multiple slices
func SliceConcat[T any](slices ...[]T) []T {
	total := 0
	for _, s := range slices {
		total += len(s)
	}
	result := make([]T, total)
	off := 0
	for _, s := range slices {
		copy(result[off:], s)
		off += len(s)
	}
	return result
}
