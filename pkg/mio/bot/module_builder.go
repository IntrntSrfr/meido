package bot

import (
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/pkg/mio/discord"
)

type ModuleCommandBuilder struct {
	cmd *ModuleCommand
}

func NewModuleCommandBuilder(mod Module, name string) *ModuleCommandBuilder {
	return &ModuleCommandBuilder{
		cmd: &ModuleCommand{
			Mod:              mod,
			Name:             name,
			Triggers:         make([]string, 0),
			CooldownScope:    CooldownScopeNone,
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

func (b *ModuleCommandBuilder) WithAllowedTypes(msgType discord.MessageType) *ModuleCommandBuilder {
	b.cmd.AllowedTypes = msgType
	return b
}

func (b *ModuleCommandBuilder) WithAllowDMs() *ModuleCommandBuilder {
	b.cmd.AllowDMs = true
	return b
}

func (b *ModuleCommandBuilder) WithRunFunc(run func(*discord.DiscordMessage)) *ModuleCommandBuilder {
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

func (b *ModulePassiveBuilder) WithAllowedTypes(msgType discord.MessageType) *ModulePassiveBuilder {
	b.cmd.AllowedTypes = msgType
	return b
}

func (b *ModulePassiveBuilder) WithRunFunc(run func(*discord.DiscordMessage)) *ModulePassiveBuilder {
	b.cmd.Run = run
	return b
}

func (b *ModulePassiveBuilder) Build() *ModulePassive {
	// bunch of if-statements
	return b.cmd
}

type ModuleApplicationCommandBuilder struct {
	command *ModuleApplicationCommand
}

func NewModuleApplicationCommandBuilder(mod Module, name string) *ModuleApplicationCommandBuilder {
	return &ModuleApplicationCommandBuilder{
		command: &ModuleApplicationCommand{
			Mod: mod,
			ApplicationCommand: &discordgo.ApplicationCommand{
				Name: name,
			},
			Enabled: true,
		},
	}
}

func (b *ModuleApplicationCommandBuilder) Description(description string) *ModuleApplicationCommandBuilder {
	b.command.Description = description
	return b
}

func (b *ModuleApplicationCommandBuilder) Type(commandType discordgo.ApplicationCommandType) *ModuleApplicationCommandBuilder {
	b.command.Type = commandType
	return b
}

func (b *ModuleApplicationCommandBuilder) AddOption(option *discordgo.ApplicationCommandOption) *ModuleApplicationCommandBuilder {
	if b.command.Options == nil {
		b.command.Options = make([]*discordgo.ApplicationCommandOption, 0)
	}
	b.command.Options = append(b.command.Options, option)
	return b
}

func (b *ModuleApplicationCommandBuilder) Cooldown(cooldown time.Duration, scope CooldownScope) *ModuleApplicationCommandBuilder {
	b.command.Cooldown = cooldown
	b.command.CooldownScope = scope
	return b
}

func (b *ModuleApplicationCommandBuilder) NoDM() *ModuleApplicationCommandBuilder {
	dmPerms := false
	b.command.DMPermission = &dmPerms
	return b
}

func (b *ModuleApplicationCommandBuilder) Permissions(perms int64) *ModuleApplicationCommandBuilder {
	b.command.DefaultMemberPermissions = &perms
	return b
}

func (b *ModuleApplicationCommandBuilder) CheckBotPerms() *ModuleApplicationCommandBuilder {
	b.command.CheckBotPerms = true
	return b
}

func (b *ModuleApplicationCommandBuilder) Run(f func(*discord.DiscordApplicationCommand)) *ModuleApplicationCommandBuilder {
	b.command.Run = f
	return b
}

func (b *ModuleApplicationCommandBuilder) Build() *ModuleApplicationCommand {
	if b.command.Type == 0 {
		panic("command type cannot be 0")
	}
	if b.command.Run == nil {
		panic("missing run function")
	}
	return b.command
}
