package utils

import "testing"

func TestMin(t *testing.T) {
	a, b := 5, 10
	want := 5
	if got := Min(a, b); got != want {
		t.Errorf("Min() = %v, want %v", got, want)
	}
}

func TestMax(t *testing.T) {
	a, b := 5, 10
	want := 10
	if got := Max(a, b); got != want {
		t.Errorf("Max() = %v, want %v", got, want)
	}
}

func TestIsNumber(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "number",
			input: "163454407999094786",
			want:  true,
		},
		{
			name:  "string",
			input: "asdfasdf",
			want:  false,
		},
		{
			name:  "string and number",
			input: "1234 asdf 1234",
			want:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNumber(tt.input); got != tt.want {
				t.Errorf("IsNumber() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClamp(t *testing.T) {
	type args struct {
		lower int
		upper int
		n     int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "0, 10, 15",
			args: args{0, 10, 15},
			want: 10,
		},
		{
			name: "0, 10, 5",
			args: args{0, 10, 5},
			want: 5,
		},
		{
			name: "-15, 10, -35",
			args: args{-15, 10, -35},
			want: -15,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Clamp(tt.args.lower, tt.args.upper, tt.args.n); got != tt.want {
				t.Errorf("Clamp() = %v, want %v", got, tt.want)
			}
		})
	}
}
