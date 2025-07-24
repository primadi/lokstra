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

	// if totalLength == 0 return empty slice
	if totalLength == 0 {
		var emptySlice []T
		return emptySlice
	}

	return slices.Concat(s...)
}
