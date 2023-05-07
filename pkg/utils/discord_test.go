package utils

import "testing"

func TestTrimUserID(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "ID",
			input: "163454407999094786",
			want:  "163454407999094786",
		},
		{
			name:  "<@ID>",
			input: "<@163454407999094786>",
			want:  "163454407999094786",
		},
		{
			name:  "<@!ID>",
			input: "<@!163454407999094786>",
			want:  "163454407999094786",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TrimUserID(tt.input); got != tt.want {
				t.Errorf("TrimUserID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrimChannelID(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "ID",
			input: "393558442977263619",
			want:  "393558442977263619",
		},
		{
			name:  "<#ID>",
			input: "<#393558442977263619>",
			want:  "393558442977263619",
		},
		{
			name:  "<#!ID>",
			input: "<#!393558442977263619>",
			want:  "393558442977263619",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TrimChannelID(tt.input); got != tt.want {
				t.Errorf("TrimChannelID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrimRoleID(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "ID",
			input: "394302349721731072",
			want:  "394302349721731072",
		},
		{
			name:  "<&ID>",
			input: "<&394302349721731072>",
			want:  "394302349721731072",
		},
		{
			name:  "<&!ID>",
			input: "<&!394302349721731072>",
			want:  "394302349721731072",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TrimRoleID(tt.input); got != tt.want {
				t.Errorf("TrimRoleID() = %v, want %v", got, tt.want)
			}
		})
	}
}
