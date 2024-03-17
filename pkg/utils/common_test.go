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
			if result := IsNumber(tc.str); result != tc.want {
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
			if result := Clamp(tc.lower, tc.upper, tc.n); result != tc.want {
				t.Errorf("Clamp(%d, %d, %d) = %d; expected %d", tc.lower, tc.upper, tc.n, result, tc.want)
			}
		})
	}
}

func TestJoinStrings(t *testing.T) {
	testCases := []struct {
		sep  string
		strs []string
		want string
	}{
		{",", []string{"a", "b", "c"}, "a,b,c"},
		{"", []string{"a", "b", "c"}, "abc"},
		{"-", []string{"a", "b", "c"}, "a-b-c"},
		{"-", []string{}, ""},
		{"-", []string{"a"}, "a"},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			if result := JoinStrings(tc.sep, tc.strs...); result != tc.want {
				t.Errorf("JoinStrings(%s, %v) = %s; expected %s", tc.sep, tc.strs, result, tc.want)
			}
		})
	}
}

func TestStringInSlice(t *testing.T) {
	testCases := []struct {
		str   string
		slice []string
		want  bool
	}{
		{"a", []string{"a", "b", "c"}, true},
		{"d", []string{"a", "b", "c"}, false},
		{"", []string{"a", "b", "c"}, false},
		{"a", []string{}, false},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			if result := StringInSlice(tc.str, tc.slice); result != tc.want {
				t.Errorf("StringInSlice(%s, %v) = %v; expected %v", tc.str, tc.slice, result, tc.want)
			}
		})
	}
}
