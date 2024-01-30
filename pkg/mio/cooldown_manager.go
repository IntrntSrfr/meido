package mio

import (
	"sync"
	"time"
)

type CooldownManager struct {
	sync.Mutex
	m map[string]time.Time
}

func NewCooldownManager() *CooldownManager {
	return &CooldownManager{
		m: make(map[string]time.Time),
	}
}

// Set sets a command on cooldown, adding it to the map.
func (c *CooldownManager) Set(key string, dur time.Duration) {
	if dur == 0 {
		return
	}

	c.Lock()
	c.m[key] = time.Now().Add(dur)
	c.Unlock()

	go func() {
		time.AfterFunc(dur, func() {
			c.Remove(key)
		})
	}()
}

// Check checks whether a command is on cooldown. If cooldown exists,
// the time left is returned.
func (c *CooldownManager) Check(key string) (time.Duration, bool) {
	c.Lock()
	defer c.Unlock()
	t, ok := c.m[key]
	return time.Until(t), ok
}

func (c *CooldownManager) Remove(key string) {
	c.Lock()
	defer c.Unlock()
	delete(c.m, key)
}
