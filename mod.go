package meidov2

type Mod interface {
	Hook(*Bot) error
	Message(*DiscordMessage)
	Settings(*DiscordMessage)
	Help(*DiscordMessage)
	Save() error
	Load() error
	Commands() map[string]ModCommand
}

type ModCommand interface {
	Name() string
	Aliases() []string
	Triggers() []string
	Description() string
	Usage() string
	RequiredPerms() int
	OwnerOnly() bool
	Enabled() bool
	Cooldown() int
	Run(*DiscordMessage)
}

/*
type ModCommand struct {
	Name string
	Aliases []string
	Triggers []string
	RequiredPerms int
	OwnerOnly bool
	Enabled bool
	Run func(*Message)
}
*/
