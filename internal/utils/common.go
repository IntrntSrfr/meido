package utils

// Min returns the minimum of two numbers.
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Max returns the maximum of two numbers.
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Clamp clamps a number between a lower and upper limit.
func Clamp(lower, upper, n int) int {
	return Max(lower, Min(upper, n))
}
