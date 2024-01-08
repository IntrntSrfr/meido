package mio

import (
	"testing"
	"time"
)

func TestSetMethod(t *testing.T) {
	handler := NewCooldownHandler()
	key := "testKey"
	dur := 5 * time.Second

	handler.Set(key, dur)
	if _, ok := handler.m[key]; !ok {
		t.Errorf("Expected the key to be set with a cooldown")
	}

	handler.Set("zeroDurKey", 0)
	if _, ok := handler.m["zeroDurKey"]; ok {
		t.Errorf("Expected the key with zero duration not to be set")
	}
}

func TestCheckMethod(t *testing.T) {
	handler := NewCooldownHandler()
	key := "testKey"
	dur := 1 * time.Second

	handler.Set(key, dur)
	_, ok := handler.Check(key)
	if !ok {
		t.Errorf("Expected the key to be on cooldown")
	}
}

func TestRemoveMethod(t *testing.T) {
	handler := NewCooldownHandler()
	key := "testKey"

	handler.Set(key, 5*time.Second)
	handler.Remove(key)
	if _, ok := handler.m[key]; ok {
		t.Errorf("Expected the key to be removed from cooldown")
	}
}
