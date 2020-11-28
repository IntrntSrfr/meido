package meidov2

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
	Run           func(*DiscordMessage) `json:"-"`
}

type ModPassive struct {
	Mod          Mod
	Name         string
	Description  string
	AllowedTypes MessageType
	Enabled      bool
	Run          func(*DiscordMessage) `json:"-"`
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
