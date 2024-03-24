package builders

import "github.com/bwmarrin/discordgo"

type SubCommandGroupBuilder struct {
	subCommandGroup *discordgo.ApplicationCommandOption
}

func NewSubCommandGroupBuilder(name, description string) *SubCommandGroupBuilder {
	return &SubCommandGroupBuilder{&discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
		Name:        name,
		Description: description,
	}}
}

func (b *SubCommandGroupBuilder) AddSubCommand(subCommand *discordgo.ApplicationCommandOption) *SubCommandGroupBuilder {
	if b.subCommandGroup.Options == nil {
		b.subCommandGroup.Options = make([]*discordgo.ApplicationCommandOption, 0)
	}
	b.subCommandGroup.Options = append(b.subCommandGroup.Options, subCommand)
	return b
}

func (b *SubCommandGroupBuilder) Build() *discordgo.ApplicationCommandOption {
	return b.subCommandGroup
}

type SubCommandBuilder struct {
	subCommand *discordgo.ApplicationCommandOption
}

func NewSubCommandBuilder(name, description string) *SubCommandBuilder {
	return &SubCommandBuilder{&discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionSubCommand,
		Name:        name,
		Description: description,
	}}
}

func (b *SubCommandBuilder) AddOption(option *discordgo.ApplicationCommandOption) *SubCommandBuilder {
	if b.subCommand.Options == nil {
		b.subCommand.Options = make([]*discordgo.ApplicationCommandOption, 0)
	}
	b.subCommand.Options = append(b.subCommand.Options, option)
	return b
}

func (b *SubCommandBuilder) Build() *discordgo.ApplicationCommandOption {
	return b.subCommand
}
