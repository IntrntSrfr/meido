package utils

import "testing"

func TestIsNumber(t *testing.T) {
	testCases := []struct {
		str  string
		want bool
	}{
		{"123", true},
		{"abc", false},
		{"", false},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			result := IsNumber(tc.str)
			if result != tc.want {
				t.Errorf("IsNumber(%s) = %v; expected %v", tc.str, result, tc.want)
			}
		})
	}
}

func TestClamp(t *testing.T) {
	testCases := []struct {
		lower, upper, n int
		want            int
	}{
		{0, 10, -5, 0},
		{0, 10, 15, 10},
		{0, 10, 5, 5},
		{5, 5, 7, 5},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			result := Clamp(tc.lower, tc.upper, tc.n)
			if result != tc.want {
				t.Errorf("Clamp(%d, %d, %d) = %d; expected %d", tc.lower, tc.upper, tc.n, result, tc.want)
			}
		})
	}
}
