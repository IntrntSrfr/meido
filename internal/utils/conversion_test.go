package utils

import "testing"

func TestCelsiusToFahrenheit(t *testing.T) {
	tests := []struct {
		name string
		args float64
		want float64
	}{
		{"25c to f", 25.0, 77.0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := CelsiusToFahrenheit(tc.args); got != tc.want {
				t.Errorf("CelsiusToFahrenheit(%v) = %v; expected %v", tc.args, got, tc.want)
			}
		})
	}
}
