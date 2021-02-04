package base

import (
	"strconv"
	"strings"
	"time"
)

// Mod represents a collection of commands and passives.
type Mod interface {
	Name() string
	Save() error
	Load() error
	Passives() []*ModPassive
	Commands() map[string]*ModCommand
	AllowedTypes() MessageType
	AllowDMs() bool
	Hook(*Bot) error
	RegisterCommand(*ModCommand)
}

// ModCommand represents a command for a Mod.
type ModCommand struct {
	Mod           Mod
	Name          string
	Description   string
	Triggers      []string
	Usage         string
	Cooldown      int
	CooldownUser  bool
	RequiredPerms int64
	RequiresOwner bool
	CheckBotPerms bool
	AllowedTypes  MessageType
	AllowDMs      bool
	Enabled       bool
	Run           func(*DiscordMessage) `json:"-"`
}

// ModPassive represents a passive for a Mod.
type ModPassive struct {
	Mod          Mod
	Name         string
	Description  string
	AllowedTypes MessageType
	Enabled      bool
	Run          func(*DiscordMessage) `json:"-"`
}

// Min returns the minimum of two numbers. Convenience function.
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Max returns the maximum of two numbers. Convenience function.
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Clamp clamps a number between a lower and upper limit. Convenience function.
func Clamp(lower, upper, n int) int {
	return Max(lower, Min(upper, n))
}

// IDToTimestamp converts a discord ID to a timestamp
func IDToTimestamp(idStr string) time.Time {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return time.Now()
	}

	id = ((id >> 22) + 1420070400000) / 1000
	return time.Unix(id, 0)
}

func TrimUserID(id string) string {

	id = strings.TrimPrefix(id, "<@!")
	id = strings.TrimSuffix(id, ">")

	return id
}
