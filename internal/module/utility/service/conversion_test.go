package service

import (
	"fmt"
	"testing"
)

func TestConversion(t *testing.T) {
	tests := []struct {
		name     string
		unitFrom string
		unitTo   string
		value    float64
		want     float64
		wantErr  bool
	}{
		{name: "cm to m", unitFrom: "cm", unitTo: "m", value: 100, want: 1, wantErr: false},
		{name: "s to h", unitFrom: "s", unitTo: "m", value: 1, want: 1.0 / 60.0, wantErr: false},
		{name: "cm to day", unitFrom: "cm", unitTo: "day", value: 1, want: 10, wantErr: true},
		{name: "c to f", unitFrom: "c", unitTo: "f", value: 0, want: 32, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Convert(tt.value, tt.unitFrom, tt.unitTo)
			fmt.Println(got, err)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Convert() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			if got != tt.want {
				t.Errorf("Convert() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convertTemperature(t *testing.T) {
	tests := []struct {
		name     string
		unitFrom string
		unitTo   string
		value    float64
		want     float64
		wantErr  bool
	}{
		{name: "c to f", unitFrom: "c", unitTo: "f", value: 0, want: 32, wantErr: false},
		{name: "c to k", unitFrom: "c", unitTo: "k", value: 0, want: 273.15, wantErr: false},
		{name: "c to c", unitFrom: "c", unitTo: "c", value: 0, want: 0, wantErr: false},
		{name: "f to c", unitFrom: "f", unitTo: "c", value: 32, want: 0, wantErr: false},
		{name: "f to k", unitFrom: "f", unitTo: "k", value: 32, want: 273.15, wantErr: false},
		{name: "f to f", unitFrom: "f", unitTo: "f", value: 32, want: 32, wantErr: false},
		{name: "k to c", unitFrom: "k", unitTo: "c", value: 273.15, want: 0, wantErr: false},
		{name: "k to f", unitFrom: "k", unitTo: "f", value: 273.15, want: 32, wantErr: false},
		{name: "k to k", unitFrom: "k", unitTo: "k", value: 273.15, want: 273.15, wantErr: false},
		{name: "c to day", unitFrom: "c", unitTo: "day", value: 0, want: 0, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertTemperature(tt.value, tt.unitFrom, tt.unitTo)
			if tt.wantErr {
				if err == nil {
					t.Errorf("convertTemperature() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			if got != tt.want {
				t.Errorf("convertTemperature() got = %v, want %v", got, tt.want)
			}
		})
	}
}
