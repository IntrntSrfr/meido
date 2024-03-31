package bot

import (
	"context"
	"errors"

	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/pkg/mio"
	"github.com/intrntsrfr/meido/pkg/mio/discord"
	"github.com/intrntsrfr/meido/pkg/mio/test"
)

func NewTestBot() *Bot {
	bot := NewBotBuilder(test.NewTestConfig()).
		WithDefaultHandlers().
		WithDiscord(discord.NewTestDiscord(nil, nil, nil)).
		Build()
	return bot
}

func drainBotEvents(ctx context.Context, events chan *BotEventData) {
	for {
		select {
		case <-events:
		case <-ctx.Done():
			return
		}
	}
}

func NewTestModule(bot *Bot, name string, log mio.Logger) *testModule {
	return &testModule{ModuleBase: *NewModule(bot, name, log)}
}

type testModule struct {
	ModuleBase
	hookShouldFail bool
}

func (m *testModule) Hook() error {
	if m.hookShouldFail {
		return errors.New("Something terrible has happened")
	}
	return nil
}

func NewTestCommand(mod Module) *ModuleCommand {
	return NewModuleCommandBuilder(mod, "test").
		Description("testing").
		Triggers(".test").
		Usage(".test").
		Cooldown(0, CooldownScopeChannel).
		AllowedTypes(discord.MessageTypeCreate).
		Execute(testCommandRun).
		Build()
}

func testCommandRun(msg *discord.DiscordMessage) {

}

func NewTestPassive(mod Module) *ModulePassive {
	return NewModulePassiveBuilder(mod, "test").
		Description("testing").
		AllowedTypes(discord.MessageTypeCreate).
		Execute(testPassiveRun).
		Build()
}

func testPassiveRun(msg *discord.DiscordMessage) {

}

func NewTestApplicationCommand(mod Module) *ModuleApplicationCommand {
	return NewModuleApplicationCommandBuilder(mod, "test").
		Type(discordgo.ChatApplicationCommand).
		Description("testing").
		Execute(testApplicationCommandRun).
		Build()
}

func testApplicationCommandRun(msg *discord.DiscordApplicationCommand) {

}

func NewTestMessageComponent(mod Module) *ModuleMessageComponent {
	return &ModuleMessageComponent{
		Mod:           mod,
		Name:          "test",
		Cooldown:      0,
		CooldownScope: CooldownScopeChannel,
		CheckBotPerms: false,
		Enabled:       true,
		Execute:       testMessageComponentRun,
	}
}

func testMessageComponentRun(msg *discord.DiscordMessageComponent) {

}

func NewTestModalSubmit(mod Module) *ModuleModalSubmit {
	return &ModuleModalSubmit{
		Mod:     mod,
		Name:    "test",
		Enabled: true,
		Execute: testModalSubmitRun,
	}
}

func testModalSubmitRun(msg *discord.DiscordModalSubmit) {

}

func NewTestMessage(bot *Bot, guildID string) *discord.DiscordMessage {
	author := &discordgo.User{Username: "jeff"}
	msg := &discord.DiscordMessage{
		Sess:        bot.Discord.Sess,
		Discord:     bot.Discord,
		MessageType: discord.MessageTypeCreate,
		Message: &discordgo.Message{
			Content:   ".test hello",
			GuildID:   guildID,
			Author:    author,
			ChannelID: "1",
			ID:        "1",
		},
	}
	if guildID != "" {
		msg.Message.Member = &discordgo.Member{User: author}
	}
	return msg
}

func NewTestApplicationCommandInteraction(bot *Bot, guildID string) *discord.DiscordInteraction {
	it := newTestInteraction(bot, guildID)
	it.Interaction.Type = discordgo.InteractionApplicationCommand
	it.Interaction.Data = discordgo.ApplicationCommandInteractionData{
		Name:        "test",
		CommandType: discordgo.ChatApplicationCommand,
	}
	return it
}

func NewTestMessageComponentInteraction(bot *Bot, guildID, customID string) *discord.DiscordInteraction {
	it := newTestInteraction(bot, guildID)
	it.Interaction.Type = discordgo.InteractionMessageComponent
	it.Interaction.Data = discordgo.MessageComponentInteractionData{
		CustomID: customID,
	}
	return it
}

func NewTestModalSubmitInteraction(bot *Bot, guildID, customID string) *discord.DiscordInteraction {
	it := newTestInteraction(bot, guildID)
	it.Interaction.Type = discordgo.InteractionModalSubmit
	it.Interaction.Data = discordgo.ModalSubmitInteractionData{
		CustomID: customID,
	}
	return it
}

func newTestInteraction(bot *Bot, guildID string) *discord.DiscordInteraction {
	author := &discordgo.User{Username: "jeff"}
	it := &discord.DiscordInteraction{
		Sess:    bot.Discord.Sess,
		Discord: bot.Discord,
		Interaction: &discordgo.Interaction{
			ChannelID: "1",
			GuildID:   guildID,
			ID:        "1",
		},
	}
	if guildID == "" {
		it.Interaction.User = author
	} else {
		it.Interaction.Member = &discordgo.Member{User: author}
	}
	return it
}
