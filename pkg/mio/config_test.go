package mio

import (
	"testing"
)

func TestConfigBase_Set(t *testing.T) {
	config := NewConfig()
	testKey := "testKey"
	testValue := "testValue"

	config.Set(testKey, testValue)
	if got := config.GetString(testKey); got != testValue {
		t.Errorf("Set() followed by GetString() = %v, want %v", got, testValue)
	}
}

func TestConfigBase_GetString(t *testing.T) {
	config := NewConfig()
	testKey := "testKey"
	testValue := "testValue"

	config.Set(testKey, testValue)
	if got := config.GetString(testKey); got != testValue {
		t.Errorf("GetString() = %v, want %v", got, testValue)
	}

	if got := config.GetString("nonExistentKey"); got != "" {
		t.Errorf("GetString() with non-existent key = %v, want empty string", got)
	}
}

func TestConfigBase_GetInt(t *testing.T) {
	config := NewConfig()
	testKey := "testKey"
	testValue := 42

	config.Set(testKey, testValue)
	if got := config.GetInt(testKey); got != testValue {
		t.Errorf("GetInt() = %v, want %v", got, testValue)
	}

	if got := config.GetInt("nonExistentKey"); got != -1 {
		t.Errorf("GetInt() with non-existent key = %v, want -1", got)
	}
}

func TestConfigBase_GetStringSlice(t *testing.T) {
	config := NewConfig()
	testKey := "testKey"
	testValue := []string{"one", "two", "three"}

	config.Set(testKey, testValue)
	if got := config.GetStringSlice(testKey); len(got) != len(testValue) {
		t.Errorf("GetStringSlice() = %v, want %v", got, testValue)
	}

	if got := config.GetStringSlice("nonExistentKey"); len(got) != 0 {
		t.Errorf("GetStringSlice() with non-existent key = %v, want empty slice", got)
	}
}
