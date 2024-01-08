package utils

import (
	"testing"
	"time"
)

func TestIDToTimestamp(t *testing.T) {
	tests := []struct {
		idStr string
		want  time.Time
	}{
		{"163454407999094786", time.Unix(1459040967, 0)},
		{"invalid", time.Unix(0, 0)},
	}

	for _, tc := range tests {
		if got := IDToTimestamp(tc.idStr); !got.Equal(tc.want) {
			t.Errorf("IDToTimestamp(%s) = %v; want %v", tc.idStr, got, tc.want)
		}
	}
}

func TestTrimUserID(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"ID", "163454407999094786", "163454407999094786"},
		{"<@ID>", "<@163454407999094786>", "163454407999094786"},
		{"<@!ID>", "<@!163454407999094786>", "163454407999094786"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := TrimUserID(tc.input); got != tc.want {
				t.Errorf("TrimUserID(%s) = %s; want %s", tc.input, got, tc.want)
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
		{"ID", "393558442977263619", "393558442977263619"},
		{"<#ID>", "<#393558442977263619>", "393558442977263619"},
		{"<#!ID>", "<#!393558442977263619>", "393558442977263619"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := TrimChannelID(tc.input); got != tc.want {
				t.Errorf("TrimChannelID(%s) = %s; want %s", tc.input, got, tc.want)
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
		{"ID", "394302349721731072", "394302349721731072"},
		{"<&ID>", "<&394302349721731072>", "394302349721731072"},
		{"<&!ID>", "<&!394302349721731072>", "394302349721731072"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := TrimRoleID(tc.input); got != tc.want {
				t.Errorf("TrimRoleID(%s) = %s; want %s", tc.input, got, tc.want)
			}
		})
	}
}
