package mio

type CommandRan struct {
	Command *ModuleCommand
	Message *DiscordMessage
}

type CommandPanicked struct {
	Command    *ModuleCommand
	Message    *DiscordMessage
	StackTrace string
}
