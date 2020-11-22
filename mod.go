package meidov2

type Mod interface {
	Save() error
	Load() error
	Commands() map[string]ModCommand
	Hook(*Bot) error
	RegisterCommand(ModCommand)
	Settings(*DiscordMessage)
	Help(*DiscordMessage)
	Message(*DiscordMessage)
}

type ModCommand interface {
	Name() string
	Description() string
	Triggers() []string
	Usage() string
	Cooldown() int
	RequiredPerms() int
	RequiresOwner() bool
	IsEnabled() bool
	Run(*DiscordMessage)
}

/*
type ModCommandStruct struct {
	Mod
	Name          string
	Description   string
	Triggers      []string
	Usage         string
	Cooldown      int
	RequiredPerms int
	RequiresOwner bool
	Enabled       bool
}
*/

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
