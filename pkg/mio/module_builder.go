package mio

import "time"

type ModuleCommandBuilder struct {
	cmd *ModuleCommand
}

func NewModuleCommandBuilder(mod Module, name string) *ModuleCommandBuilder {
	return &ModuleCommandBuilder{
		cmd: &ModuleCommand{
			Mod:              mod,
			Name:             name,
			Triggers:         make([]string, 0),
			CooldownScope:    None,
			RequiresUserType: UserTypeAny,
			Enabled:          true,
		},
	}
}

func (b *ModuleCommandBuilder) WithDescription(text string) *ModuleCommandBuilder {
	b.cmd.Description = text
	return b
}

func (b *ModuleCommandBuilder) WithTriggers(trigs ...string) *ModuleCommandBuilder {
	b.cmd.Triggers = trigs
	return b
}

func (b *ModuleCommandBuilder) WithUsage(text string) *ModuleCommandBuilder {
	b.cmd.Usage = text
	return b
}

func (b *ModuleCommandBuilder) WithCooldown(cd time.Duration, scope CooldownScope) *ModuleCommandBuilder {
	b.cmd.Cooldown = cd
	b.cmd.CooldownScope = scope
	return b
}

func (b *ModuleCommandBuilder) WithRequiredPerms(perms int64) *ModuleCommandBuilder {
	b.cmd.RequiredPerms = perms
	return b
}

func (b *ModuleCommandBuilder) WithRequiresBotOwner() *ModuleCommandBuilder {
	b.cmd.RequiresUserType = UserTypeBotOwner
	return b
}

func (b *ModuleCommandBuilder) WithCheckBotPerms() *ModuleCommandBuilder {
	b.cmd.CheckBotPerms = true
	return b
}

func (b *ModuleCommandBuilder) WithAllowedTypes(msgType MessageType) *ModuleCommandBuilder {
	b.cmd.AllowedTypes = msgType
	return b
}

func (b *ModuleCommandBuilder) WithAllowDMs() *ModuleCommandBuilder {
	b.cmd.AllowDMs = true
	return b
}

func (b *ModuleCommandBuilder) WithRunFunc(run func(*DiscordMessage)) *ModuleCommandBuilder {
	b.cmd.Run = run
	return b
}

func (b *ModuleCommandBuilder) Build() *ModuleCommand {
	// bunch of if-statements
	return b.cmd
}

type ModulePassiveBuilder struct {
	cmd *ModulePassive
}

func NewModulePassiveBuilder(mod Module, name string) *ModulePassiveBuilder {
	return &ModulePassiveBuilder{
		cmd: &ModulePassive{
			Mod:     mod,
			Name:    name,
			Enabled: true,
		},
	}
}

func (b *ModulePassiveBuilder) WithDescription(text string) *ModulePassiveBuilder {
	b.cmd.Description = text
	return b
}

func (b *ModulePassiveBuilder) WithAllowedTypes(msgType MessageType) *ModulePassiveBuilder {
	b.cmd.AllowedTypes = msgType
	return b
}

func (b *ModulePassiveBuilder) WithRunFunc(run func(*DiscordMessage)) *ModulePassiveBuilder {
	b.cmd.Run = run
	return b
}

func (b *ModulePassiveBuilder) Build() *ModulePassive {
	// bunch of if-statements
	return b.cmd
}
