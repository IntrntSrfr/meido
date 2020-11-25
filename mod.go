package meidov2

type Mod interface {
	Name() string
	Save() error
	Load() error
	Commands() map[string]*ModCommand
	Passives() []*ModPassive
	Hook(*Bot) error
	RegisterCommand(*ModCommand)
	AllowedTypes() MessageType
	AllowDMs() bool
}

type ModCommand struct {
	Mod           Mod
	Name          string
	Description   string
	Triggers      []string
	Usage         string
	Cooldown      int
	RequiredPerms int
	RequiresOwner bool
	AllowedTypes  MessageType
	AllowDMs      bool
	Enabled       bool
	Run           func(*DiscordMessage)
}

type ModPassive struct {
	Mod          Mod
	Name         string
	Description  string
	AllowedTypes MessageType
	Enabled      bool
	Run          func(*DiscordMessage)
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func Clamp(lower, upper, n int) int {
	return Max(lower, Min(upper, n))
}
