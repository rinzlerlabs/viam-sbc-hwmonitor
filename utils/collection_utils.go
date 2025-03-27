package utils

func Values[T comparable, U any](m map[T]U) []U {
	values := make([]U, 0, len(m))
	for _, value := range m {
		values = append(values, value)
	}
	return values
}

func Keys[T comparable, U any](m map[T]U) []T {
	keys := make([]T, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}

func Invert[T comparable, U comparable](m map[T]U) map[U]T {
	inverted := make(map[U]T, len(m))
	for key, value := range m {
		inverted[value] = key
	}
	return inverted
}
