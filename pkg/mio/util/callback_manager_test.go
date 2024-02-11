package util

import (
	"testing"
)

func TestCooldownService_Make(t *testing.T) {
	handler := NewCallbackManager()
	key := "testKey"
	ch, err := handler.Make(key)
	if err != nil {
		t.Errorf("Unexpected error when making a channel: %s", err)
	}
	if ch == nil {
		t.Errorf("Expected a non-nil channel")
	}

	_, err = handler.Make(key)
	if err != ErrCallbackAlreadyExists {
		t.Errorf("Expected ErrCallbackAlreadyExists error, got: %s", err)
	}
}

func TestCooldownService_Get(t *testing.T) {
	handler := NewCallbackManager()
	key := "testKey"
	handler.Make(key)

	ch, err := handler.Get(key)
	if err != nil {
		t.Errorf("Unexpected error when getting a channel: %s", err)
	}
	if ch == nil {
		t.Errorf("Expected a non-nil channel")
	}

	_, err = handler.Get("nonExistentKey")
	if err != ErrCallbackNotFound {
		t.Errorf("Expected ErrCallbackNotFound error, got: %s", err)
	}
}

func TestCooldownService_Delete(t *testing.T) {
	handler := NewCallbackManager()
	key := "testKey"

	handler.Make(key)
	handler.Delete(key)
	if _, ok := handler.ch[key]; ok {
		t.Errorf("Channel should have been deleted")
	}
}
