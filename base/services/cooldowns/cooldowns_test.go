package cooldowns

import (
	"testing"
	"time"
)

func TestBotCooldownService_Check(t *testing.T) {
	c := &BotCooldownService{
		m: map[string]time.Time{"future": time.Now()},
	}
	_, ok := c.Check("future")
	if !ok {
		t.Errorf("Check() got = %v, want %v", ok, true)
	}
}

func TestBotCooldownService_Remove(t *testing.T) {
	c := &BotCooldownService{
		m: map[string]time.Time{"future": time.Now()},
	}
	c.Remove("future")
	if len(c.m) != 0 {
		t.Errorf("Check() got = %v, want %v", len(c.m), 0)
	}
}

func TestBotCooldownService_Set(t *testing.T) {
	c := &BotCooldownService{
		m: make(map[string]time.Time),
	}
	c.Set("future", time.Second*5)
	if len(c.m) != 1 {
		t.Errorf("Check() got = %v, want %v", len(c.m), 1)
	}
}
