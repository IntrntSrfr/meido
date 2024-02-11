package util

import (
	"testing"
	"time"
)

func TestCooldownManager_Set(t *testing.T) {
	handler := NewCooldownManager()
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

func TestCooldownManager_Check(t *testing.T) {
	handler := NewCooldownManager()
	key := "testKey"
	dur := time.Millisecond * 25

	handler.Set(key, dur)
	_, ok := handler.Check(key)
	if !ok {
		t.Errorf("Expected the key to be on cooldown")
	}
	time.Sleep(time.Millisecond * 50)
	_, ok = handler.Check(key)
	if ok {
		t.Errorf("Expected the key to not be on cooldown")
	}
}

func TestCooldownManager_Remove(t *testing.T) {
	handler := NewCooldownManager()
	key := "testKey"

	handler.Set(key, 5*time.Second)
	handler.Remove(key)
	if _, ok := handler.m[key]; ok {
		t.Errorf("Expected the key to be removed from cooldown")
	}
}
