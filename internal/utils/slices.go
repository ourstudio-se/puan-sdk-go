package utils

import (
	"cmp"
	"slices"

	"github.com/go-errors/errors"
)

func Without[T comparable](sliceA []T, sliceB []T) []T {
	var result []T
	excludeMap := make(map[T]struct{}, len(sliceB))

	for _, item := range sliceB {
		excludeMap[item] = struct{}{}
	}

	for _, item := range sliceA {
		if _, found := excludeMap[item]; !found {
			result = append(result, item)
		}
	}

	return result
}

func ContainsDuplicates[T comparable](elements []T) bool {
	seen := make(map[T]any)
	for _, e := range elements {
		if _, ok := seen[e]; ok {
			return true
		}
		seen[e] = nil
	}

	return false
}

func Reverse[T any](elements []T) []T {
	result := make([]T, len(elements))
	for i := range elements {
		result[len(elements)-1-i] = elements[i]
	}

	return result
}

func Contains[T comparable](elements []T, element T) bool {
	for _, e := range elements {
		if e == element {
			return true
		}
	}

	return false
}

func ContainsAny[T comparable](a []T, b []T) bool {
	for _, e := range b {
		if Contains(a, e) {
			return true
		}
	}

	return false
}

func ContainsAll[T comparable](a []T, b []T) bool {
	for _, e := range b {
		if !Contains(a, e) {
			return false
		}
	}

	return true
}

func IndexOf[T comparable](elements []T, element T) (int, error) {
	for i, e := range elements {
		if e == element {
			return i, nil
		}
	}

	return -1, errors.New("element not found")
}

func Dedupe[T comparable](elements []T) []T {
	seen := make(map[T]struct{}, len(elements))
	var result []T

	for _, e := range elements {
		if _, ok := seen[e]; !ok {
			seen[e] = struct{}{}
			result = append(result, e)
		}
	}

	return result
}

func Filter[T any](slice []T, predicate func(T) bool) []T {
	var result []T
	for _, e := range slice {
		if predicate(e) {
			result = append(result, e)
		}
	}
	return result
}

func Sort[T cmp.Ordered](slice []T) []T {
	sorted := make([]T, len(slice))
	copy(sorted, slice)

	slices.Sort(sorted)
	return sorted
}

func SortedBy[T any, K cmp.Ordered](in []T, key func(T) K) []T {
	out := make([]T, len(in))
	copy(out, in)

	slices.SortFunc(out, func(a, b T) int {
		return cmp.Compare(key(a), key(b))
	})

	return out
}
