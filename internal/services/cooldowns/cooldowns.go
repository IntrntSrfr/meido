package cooldowns

import (
	"sync"
	"time"
)

type CooldownHandler struct {
	sync.Mutex
	m map[string]time.Time
}

func NewCooldownHandler() *CooldownHandler {
	return &CooldownHandler{
		m: make(map[string]time.Time),
	}
}

// IsOnCooldown checks whether a command is on cooldown.
// Returns the value from the CooldownCache
func (c *CooldownHandler) IsOnCooldown(key string) (time.Time, bool) {
	c.Lock()
	defer c.Unlock()
	t, ok := c.m[key]
	return t, ok
}

// SetOnCooldown sets a command on cooldown, adding it to the CooldownCache.
func (c *CooldownHandler) SetOnCooldown(key string, dur time.Duration) {

	c.Lock()
	c.m[key] = time.Now().Add(time.Second * dur)
	c.Unlock()

	go func() {
		time.AfterFunc(time.Second*dur, func() {
			c.Lock()
			delete(c.m, key)
			c.Unlock()
		})
	}()
}
