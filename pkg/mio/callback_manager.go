package mio

import (
	"errors"
	"sync"
)

var (
	ErrCallbackAlreadyExists = errors.New("callback for this key already exists")
	ErrCallbackNotFound      = errors.New("callback for this key not found")
)

type CallbackManager struct {
	sync.Mutex
	ch map[string]chan *DiscordMessage
}

func NewCallbackManager() *CallbackManager {
	return &CallbackManager{
		ch: make(map[string]chan *DiscordMessage),
	}
}

// Make makes a channel for future communication with a running command
func (c *CallbackManager) Make(key string) (chan *DiscordMessage, error) {
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
func (c *CallbackManager) Get(key string) (chan *DiscordMessage, error) {
	c.Lock()
	defer c.Unlock()
	ch, ok := c.ch[key]
	if !ok {
		return nil, ErrCallbackNotFound
	}
	return ch, nil
}

// Delete removes a channel for communication with a running command
func (c *CallbackManager) Delete(key string) {
	c.Lock()
	defer c.Unlock()
	close(c.ch[key])
	delete(c.ch, key)
}
