package meidov2

type Mod interface {
	Hook(*Bot, chan *DiscordMessage) error
	Message(*DiscordMessage)
}

type ModCommand interface{
	Name()string
	Aliases() []string
	Triggers()[]string
	RequiredPerms()int
	OwnerOnly()bool
	Enabled()bool
}