package bot

import (
	"context"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/pkg/mio/discord"
	mutils "github.com/intrntsrfr/meido/pkg/mio/utils"
	"github.com/intrntsrfr/meido/pkg/utils"
	"go.uber.org/zap"
)

type Bot struct {
	sync.Mutex
	Discord *discord.Discord
	Config  *utils.Config

	*ModuleManager
	EventHandler *EventHandler
	Callbacks    *mutils.CallbackManager
	Cooldowns    *mutils.CooldownManager
	*EventEmitter

	Logger *zap.Logger
}

func (b *Bot) UseDefaultHandlers() {
	b.Discord.AddEventHandler(readyHandler(b))
	b.Discord.AddEventHandler(guildJoinHandler(b))
	b.Discord.AddEventHandler(guildLeaveHandler(b))
	b.Discord.AddEventHandler(memberChunkHandler(b))
}

func (b *Bot) Run(ctx context.Context) error {
	b.Logger.Info("Starting up...")

	go b.EventHandler.Listen(ctx)
	if err := b.Discord.Run(); err != nil {
		return err
	}
	if err := b.setApplicationCommands(); err != nil {
		return err
	}
	b.Logger.Info("Running")
	return nil
}

func (b *Bot) Close() {
	b.Logger.Info("Shutting down")
	b.Discord.Close()
}

func (b *Bot) setApplicationCommands() error {
	var allCommands []*discordgo.ApplicationCommand
	for _, m := range b.Modules {
		allCommands = append(allCommands, m.ApplicationCommandStructs()...)
	}

	created, err := b.Discord.Sess.ApplicationCommandBulkOverwrite(b.Discord.Sess.State().User.ID, "", allCommands)
	if err != nil {
		b.logger.Error("could not overwrite commands", zap.Error(err))
		return err
	}
	for _, c := range created {
		b.logger.Info("created/updated command", zap.String("name", c.Name))
	}
	return nil
}

func (b *Bot) IsOwner(userID string) bool {
	for _, id := range b.Config.GetStringSlice("owner_ids") {
		if id == userID {
			return true
		}
	}
	return false
}
