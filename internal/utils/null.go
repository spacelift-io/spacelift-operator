package utils

func AddressOf[T any](v T) *T {
	return &v
}
