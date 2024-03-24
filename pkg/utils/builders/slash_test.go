package builders

import (
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
)

func TestAddSubCommand(t *testing.T) {
	builder := NewSubCommandGroupBuilder("testGroup", "A test command group")
	subCommand := &discordgo.ApplicationCommandOption{Name: "subTest", Description: "A sub test command", Type: discordgo.ApplicationCommandOptionSubCommand}
	builder.AddSubCommand(subCommand)

	assert.Equal(t, 1, len(builder.subCommandGroup.Options))
	assert.Equal(t, subCommand, builder.subCommandGroup.Options[0])
}

func TestSubCommandGroupBuilder_Build(t *testing.T) {
	builder := NewSubCommandGroupBuilder("testGroup", "A test command group")
	result := builder.Build()

	assert.Equal(t, builder.subCommandGroup, result)
}

func TestNewSubCommandBuilder(t *testing.T) {
	builder := NewSubCommandBuilder("testCommand", "A test command")
	assert.NotNil(t, builder)
	assert.Equal(t, discordgo.ApplicationCommandOptionSubCommand, builder.subCommand.Type)
	assert.Equal(t, "testCommand", builder.subCommand.Name)
	assert.Equal(t, "A test command", builder.subCommand.Description)
}

func TestAddOption(t *testing.T) {
	builder := NewSubCommandBuilder("testCommand", "A test command")
	option := &discordgo.ApplicationCommandOption{Name: "optionTest", Description: "An option test", Type: discordgo.ApplicationCommandOptionString}
	builder.AddOption(option)

	assert.Equal(t, 1, len(builder.subCommand.Options))
	assert.Equal(t, option, builder.subCommand.Options[0])
}

func TestSubCommandBuilder_Build(t *testing.T) {
	builder := NewSubCommandBuilder("testCommand", "A test command")
	result := builder.Build()

	assert.Equal(t, builder.subCommand, result)
}
