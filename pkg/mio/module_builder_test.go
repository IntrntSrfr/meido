package mio

import (
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/pkg/mio"
	"github.com/intrntsrfr/meido/pkg/mio/discord"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModuleCommandBuilder(t *testing.T) {
	mod := NewTestModule(nil, "test", mio.NewDiscardLogger())

	t.Run("NewModuleCommandBuilder", func(t *testing.T) {
		name := "testCommand"
		builder := NewModuleCommandBuilder(mod, name)

		require.NotNil(t, builder, "Builder should not be nil")
		assert.Equal(t, name, builder.cmd.Name, "Command name mismatch")
		assert.True(t, builder.cmd.Enabled, "Command should be enabled by default")
	})

	t.Run("Description", func(t *testing.T) {
		description := "A test command"
		builder := NewModuleCommandBuilder(mod, "testCommand").Description(description)

		assert.Equal(t, description, builder.cmd.Description, "Description mismatch")
	})

	t.Run("RequiredPerms", func(t *testing.T) {
		permissions := int64(12345)
		builder := NewModuleCommandBuilder(mod, "testCommand").RequiredPerms(permissions)

		assert.Equal(t, permissions, builder.cmd.RequiredPerms, "RequiredPerms mismatch")
	})

	t.Run("RequiresBotOwner", func(t *testing.T) {
		builder := NewModuleCommandBuilder(mod, "testCommand").RequiresBotOwner()

		assert.Equal(t, UserTypeBotOwner, builder.cmd.RequiresUserType, "RequiresUserType mismatch")
	})

	t.Run("CheckBotPerms", func(t *testing.T) {
		builder := NewModuleCommandBuilder(mod, "testCommand").CheckBotPerms()

		assert.True(t, builder.cmd.CheckBotPerms, "CheckBotPerms should be true")
	})

	t.Run("AllowedTypes", func(t *testing.T) {
		msgType := discord.MessageTypeCreate
		builder := NewModuleCommandBuilder(mod, "testCommand").AllowedTypes(msgType)

		assert.Equal(t, msgType, builder.cmd.AllowedTypes, "AllowedTypes mismatch")
	})

	t.Run("AllowDMs", func(t *testing.T) {
		builder := NewModuleCommandBuilder(mod, "testCommand").AllowDMs()

		assert.True(t, builder.cmd.AllowDMs, "AllowDMs should be true")
	})

	t.Run("Execute", func(t *testing.T) {
		exec := func(msg *discord.DiscordMessage) {}
		builder := NewModuleCommandBuilder(mod, "testCommand").Execute(exec)

		require.NotNil(t, builder.cmd.Execute, "Execute should be set")
	})

	t.Run("Build", func(t *testing.T) {
		t.Run("Panic on AllowedTypes zero", func(t *testing.T) {
			assert.Panics(t, func() {
				NewModuleCommandBuilder(mod, "testCommand").Execute(func(msg *discord.DiscordMessage) {}).Build()
			}, "Build should panic if allowed types is 0")
		})

		t.Run("Panic on missing execute", func(t *testing.T) {
			assert.Panics(t, func() {
				NewModuleCommandBuilder(mod, "testCommand").AllowedTypes(discord.MessageTypeCreate).Build()
			}, "Build should panic if execute is missing")
		})

		t.Run("Successful", func(t *testing.T) {
			exec := func(msg *discord.DiscordMessage) {}
			command := NewModuleCommandBuilder(mod, "testCommand").
				AllowedTypes(discord.MessageTypeCreate).
				Execute(exec).
				Build()

			require.NotNil(t, command, "Command should not be nil after build")
			assert.Equal(t, discord.MessageTypeCreate, command.AllowedTypes, "AllowedTypes mismatch after build")
			require.NotNil(t, command.Execute, "Execute should be set after build")
		})
	})
}

func TestModulePassiveBuilder(t *testing.T) {
	mod := NewTestModule(nil, "test", mio.NewDiscardLogger())

	t.Run("NewModulePassiveBuilder", func(t *testing.T) {
		name := "testPassive"
		builder := NewModulePassiveBuilder(mod, name)

		require.NotNil(t, builder, "Builder should not be nil")
		assert.Equal(t, name, builder.pas.Name, "Passive command name mismatch")
		assert.True(t, builder.pas.Enabled, "Passive command should be enabled by default")
	})

	t.Run("Description", func(t *testing.T) {
		description := "A test passive command"
		builder := NewModulePassiveBuilder(mod, "testPassive").Description(description)

		assert.Equal(t, description, builder.pas.Description, "Description mismatch")
	})

	t.Run("AllowedTypes", func(t *testing.T) {
		msgType := discord.MessageTypeCreate
		builder := NewModulePassiveBuilder(mod, "testPassive").AllowedTypes(msgType)

		assert.Equal(t, msgType, builder.pas.AllowedTypes, "AllowedTypes mismatch")
	})

	t.Run("Execute", func(t *testing.T) {
		exec := func(msg *discord.DiscordMessage) {}
		builder := NewModulePassiveBuilder(mod, "testPassive").Execute(exec)

		require.NotNil(t, builder.pas.Execute, "Execute should be set")
	})

	t.Run("Build", func(t *testing.T) {
		t.Run("Panic on AllowedTypes zero", func(t *testing.T) {
			assert.Panics(t, func() {
				NewModulePassiveBuilder(mod, "testPassive").Execute(func(msg *discord.DiscordMessage) {}).Build()
			}, "Build should panic if allowed types is 0")
		})

		t.Run("Panic on missing execute", func(t *testing.T) {
			assert.Panics(t, func() {
				NewModulePassiveBuilder(mod, "testPassive").AllowedTypes(discord.MessageTypeCreate).Build()
			}, "Build should panic if execute is missing")
		})

		t.Run("Successful", func(t *testing.T) {
			exec := func(msg *discord.DiscordMessage) {}
			passive := NewModulePassiveBuilder(mod, "testPassive").
				AllowedTypes(discord.MessageTypeCreate).
				Execute(exec).
				Build()

			require.NotNil(t, passive, "Passive command should not be nil after build")
			assert.Equal(t, discord.MessageTypeCreate, passive.AllowedTypes, "AllowedTypes mismatch after build")
			require.NotNil(t, passive.Execute, "Execute should be set after build")
		})
	})
}

func TestModuleApplicationCommandBuilder(t *testing.T) {
	mod := NewTestModule(nil, "test", mio.NewDiscardLogger())

	t.Run("NewModuleApplicationCommandBuilder", func(t *testing.T) {
		name := "testCommand"
		builder := NewModuleApplicationCommandBuilder(mod, name)

		require.NotNil(t, builder, "Expected builder to not be nil")
		assert.Equal(t, name, builder.command.Name, "Command name mismatch")
		assert.True(t, builder.command.Enabled, "Command should be enabled by default")
	})

	t.Run("Description", func(t *testing.T) {
		expectedDescription := "A test command"
		builder := NewModuleApplicationCommandBuilder(mod, "testCommand").Description(expectedDescription)

		assert.Equal(t, expectedDescription, builder.command.Description, "Description mismatch")
	})

	t.Run("Type", func(t *testing.T) {
		builder := NewModuleApplicationCommandBuilder(mod, "testCommand").Type(discordgo.ChatApplicationCommand)

		assert.Equal(t, discordgo.ChatApplicationCommand, builder.command.Type, "Command type mismatch")
	})

	t.Run("AddOption", func(t *testing.T) {
		option := &discordgo.ApplicationCommandOption{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "option",
			Description: "An option",
			Required:    true,
		}
		builder := NewModuleApplicationCommandBuilder(mod, "testCommand").AddOption(option)

		require.Len(t, builder.command.Options, 1, "Options slice should have one option")
		assert.Equal(t, option, builder.command.Options[0], "Option mismatch")
	})

	t.Run("Cooldown", func(t *testing.T) {
		expectedDuration := 10 * time.Second
		expectedScope := CooldownScopeChannel
		builder := NewModuleApplicationCommandBuilder(mod, "testCommand").Cooldown(expectedDuration, expectedScope)

		assert.Equal(t, expectedDuration, builder.command.Cooldown, "Cooldown duration mismatch")
		assert.Equal(t, expectedScope, builder.command.CooldownScope, "Cooldown scope mismatch")
	})

	t.Run("NoDM", func(t *testing.T) {
		builder := NewModuleApplicationCommandBuilder(mod, "testCommand").NoDM()

		require.NotNil(t, builder.command.DMPermission, "DMPermission should not be nil")
		assert.False(t, *builder.command.DMPermission, "DMPermission should be false")
	})

	t.Run("Permissions", func(t *testing.T) {
		expectedPerms := int64(12345)
		builder := NewModuleApplicationCommandBuilder(mod, "testCommand").Permissions(expectedPerms)

		require.NotNil(t, builder.command.DefaultMemberPermissions, "DefaultMemberPermissions should not be nil")
		assert.Equal(t, expectedPerms, *builder.command.DefaultMemberPermissions, "Permissions mismatch")
	})

	t.Run("CheckBotPerms", func(t *testing.T) {
		builder := NewModuleApplicationCommandBuilder(mod, "testCommand").CheckBotPerms()

		assert.True(t, builder.command.CheckBotPerms, "CheckBotPerms should be true")
	})

	t.Run("Execute", func(t *testing.T) {
		exec := func(cmd *discord.DiscordApplicationCommand) {}
		builder := NewModuleApplicationCommandBuilder(mod, "testCommand").Execute(exec)

		require.NotNil(t, builder.command.Execute, "Execute should be set")
	})

	t.Run("Build", func(t *testing.T) {
		t.Run("Panic on Type zero", func(t *testing.T) {
			assert.Panics(t, func() {
				NewModuleApplicationCommandBuilder(mod, "testCommand").Execute(func(cmd *discord.DiscordApplicationCommand) {}).Build()
			}, "Build should panic if command type is 0")
		})

		t.Run("Panic on missing execute", func(t *testing.T) {
			assert.Panics(t, func() {
				NewModuleApplicationCommandBuilder(mod, "testCommand").Type(discordgo.ChatApplicationCommand).Build()
			}, "Build should panic if execute is missing")
		})
		t.Run("Successful", func(t *testing.T) {
			exec := func(cmd *discord.DiscordApplicationCommand) {}
			command := NewModuleApplicationCommandBuilder(mod, "testCommand").
				Type(discordgo.ChatApplicationCommand).
				Execute(exec).
				Build()

			require.NotNil(t, command, "Command should not be nil after build")
			assert.Equal(t, discordgo.ChatApplicationCommand, command.Type, "Command type mismatch after build")
			require.NotNil(t, command.Execute, "Execute should be set after build")
		})
	})
}
