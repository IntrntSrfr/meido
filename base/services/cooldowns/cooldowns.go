package cooldowns

import (
	"sync"
	"time"
)

type CooldownService interface {
	Set(key string, dur time.Duration)
	Check(key string) (time.Duration, bool)
	Remove(key string)
}

type BotCooldownService struct {
	sync.Mutex
	m map[string]time.Time
}

func NewCooldownHandler() *BotCooldownService {
	return &BotCooldownService{
		m: make(map[string]time.Time),
	}
}

// Set sets a command on cooldown, adding it to the map.
func (c *BotCooldownService) Set(key string, dur time.Duration) {
	if dur == 0 {
		return
	}

	c.Lock()
	c.m[key] = time.Now().Add(time.Second * dur)
	c.Unlock()

	go func() {
		time.AfterFunc(time.Second*dur, func() {
			c.Remove(key)
		})
	}()
}

// Check checks whether a command is on cooldown. If cooldown exists,
// the time left is returned.
func (c *BotCooldownService) Check(key string) (time.Duration, bool) {
	c.Lock()
	defer c.Unlock()
	t, ok := c.m[key]
	return time.Until(t), ok
}

func (c *BotCooldownService) Remove(key string) {
	c.Lock()
	defer c.Unlock()
	delete(c.m, key)
}
