package meidov2

import "github.com/jmoiron/sqlx"

type Mod interface {
	Hook(*Bot, *sqlx.DB, chan *DiscordMessage) error
	Message(*DiscordMessage)
	Settings(*DiscordMessage)
	Help(*DiscordMessage)
	Save() error
	Load() error
	Commands() []ModCommand
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
