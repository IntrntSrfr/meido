package utils

// Color codes
var (
	ColorCritical = 0xc80000
	ColorGreen    = 0x00c800
	//ColorBlue      = 0x0000c8
	ColorLightBlue = 0x00bbe0
	ColorInfo      = 0xFEFEFE
)

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
