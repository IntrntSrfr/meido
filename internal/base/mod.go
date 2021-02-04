package base

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
