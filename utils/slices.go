package utils

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
