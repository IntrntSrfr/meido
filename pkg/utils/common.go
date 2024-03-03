package utils

import "strconv"

var (
	ColorCritical  = 0xc80000
	ColorGreen     = 0x00c800
	ColorLightBlue = 0x00bbe0
	ColorInfo      = 0xFEFEFE
)

func Clamp(lower, upper, n int) int {
	return max(lower, min(upper, n))
}

func IsNumber(str string) bool {
	if _, err := strconv.Atoi(str); err != nil {
		return false
	}
	return true
}
