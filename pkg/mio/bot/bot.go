package bot

import (
	"context"
	"sync"

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
	b.Logger.Info("Running")
	return nil

}

func (b *Bot) Close() {
	b.Logger.Info("Shutting down")
	b.Discord.Close()
}

func (b *Bot) IsOwner(userID string) bool {
	for _, id := range b.Config.GetStringSlice("owner_ids") {
		if id == userID {
			return true
		}
	}
	return false
}
