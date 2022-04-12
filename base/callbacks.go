package base

import (
	"errors"
	"sync"
)

type CallbackService interface {
	Make(key string) (chan *DiscordMessage, error)
	Get(key string) (chan *DiscordMessage, error)
	Delete(key string)
}

var (
	ErrCallbackAlreadyExists = errors.New("callback for this key already exists")
	ErrCallbackNotFound      = errors.New("callback for this key not found")
)

type BotCallbackService struct {
	sync.Mutex
	ch map[string]chan *DiscordMessage
}

func NewCallbackHandler() *BotCallbackService {
	return &BotCallbackService{
		ch: make(map[string]chan *DiscordMessage),
	}
}

// Make makes a channel for future communication with a running command
func (c *BotCallbackService) Make(key string) (chan *DiscordMessage, error) {
	ch := make(chan *DiscordMessage)
	c.Lock()
	defer c.Unlock()
	if _, ok := c.ch[key]; ok {
		return nil, ErrCallbackAlreadyExists
	}
	c.ch[key] = ch
	return ch, nil
}

// Get gets a channel for communication with a running command
func (c *BotCallbackService) Get(key string) (chan *DiscordMessage, error) {
	c.Lock()
	defer c.Unlock()
	ch, ok := c.ch[key]
	if !ok {
		return nil, ErrCallbackNotFound
	}
	return ch, nil
}

// Delete removes a channel for communication with a running command
func (c *BotCallbackService) Delete(key string) {
	c.Lock()
	defer c.Unlock()
	close(c.ch[key])
	delete(c.ch, key)
}
