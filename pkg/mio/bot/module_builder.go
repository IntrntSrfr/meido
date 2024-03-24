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

func (b *ModuleCommandBuilder) Description(text string) *ModuleCommandBuilder {
	b.cmd.Description = text
	return b
}

func (b *ModuleCommandBuilder) Triggers(trigs ...string) *ModuleCommandBuilder {
	b.cmd.Triggers = trigs
	return b
}

func (b *ModuleCommandBuilder) Usage(text string) *ModuleCommandBuilder {
	b.cmd.Usage = text
	return b
}

func (b *ModuleCommandBuilder) Cooldown(duration time.Duration, scope CooldownScope) *ModuleCommandBuilder {
	b.cmd.Cooldown = duration
	b.cmd.CooldownScope = scope
	return b
}

func (b *ModuleCommandBuilder) RequiredPerms(perms int64) *ModuleCommandBuilder {
	b.cmd.RequiredPerms = perms
	return b
}

func (b *ModuleCommandBuilder) RequiresBotOwner() *ModuleCommandBuilder {
	b.cmd.RequiresUserType = UserTypeBotOwner
	return b
}

func (b *ModuleCommandBuilder) CheckBotPerms() *ModuleCommandBuilder {
	b.cmd.CheckBotPerms = true
	return b
}

func (b *ModuleCommandBuilder) AllowedTypes(msgType discord.MessageType) *ModuleCommandBuilder {
	b.cmd.AllowedTypes = msgType
	return b
}

func (b *ModuleCommandBuilder) AllowDMs() *ModuleCommandBuilder {
	b.cmd.AllowDMs = true
	return b
}

func (b *ModuleCommandBuilder) Execute(exec func(*discord.DiscordMessage)) *ModuleCommandBuilder {
	b.cmd.Execute = exec
	return b
}

func (b *ModuleCommandBuilder) Build() *ModuleCommand {
	if b.cmd.AllowedTypes == 0 {
		panic("allowed types cannot be 0")
	}
	if b.cmd.Execute == nil {
		panic("missing execute")
	}
	return b.cmd
}

type ModulePassiveBuilder struct {
	pas *ModulePassive
}

func NewModulePassiveBuilder(mod Module, name string) *ModulePassiveBuilder {
	return &ModulePassiveBuilder{
		pas: &ModulePassive{
			Mod:     mod,
			Name:    name,
			Enabled: true,
		},
	}
}

func (b *ModulePassiveBuilder) Description(text string) *ModulePassiveBuilder {
	b.pas.Description = text
	return b
}

func (b *ModulePassiveBuilder) AllowedTypes(msgType discord.MessageType) *ModulePassiveBuilder {
	b.pas.AllowedTypes = msgType
	return b
}

func (b *ModulePassiveBuilder) Execute(exec func(*discord.DiscordMessage)) *ModulePassiveBuilder {
	b.pas.Execute = exec
	return b
}

func (b *ModulePassiveBuilder) Build() *ModulePassive {
	if b.pas.AllowedTypes == 0 {
		panic("allowed types cannot be 0")
	}
	if b.pas.Execute == nil {
		panic("missing execute")
	}
	return b.pas
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

func (b *ModuleApplicationCommandBuilder) AddSubcommand(subcommand *discordgo.ApplicationCommandOption) *ModuleApplicationCommandBuilder {
	subcommand.Type = discordgo.ApplicationCommandOptionSubCommand
	return b.AddOption(subcommand)
}

func (b *ModuleApplicationCommandBuilder) AddSubcommandGroup(group *discordgo.ApplicationCommandOption) *ModuleApplicationCommandBuilder {
	group.Type = discordgo.ApplicationCommandOptionSubCommandGroup
	return b.AddOption(group)
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

func (b *ModuleApplicationCommandBuilder) Execute(f func(*discord.DiscordApplicationCommand)) *ModuleApplicationCommandBuilder {
	b.command.Execute = f
	return b
}

func (b *ModuleApplicationCommandBuilder) Build() *ModuleApplicationCommand {
	if b.command.Type == 0 {
		panic("command type cannot be 0")
	}
	if b.command.Execute == nil {
		panic("missing execute")
	}
	return b.command
}
