package mio

import (
	"context"
	"sync"

	"go.uber.org/zap"
)

type Bot struct {
	sync.Mutex
	Discord *Discord
	Config  Configurable

	*ModuleManager
	MessageProcessor     *EventHandler
	InteractionProcessor *InteractionProcessor
	Callbacks            *CallbackManager
	Cooldowns            *CooldownManager
	*EventEmitter

	Logger *zap.Logger
}

func NewBot(config Configurable, logger *zap.Logger) *Bot {
	logger = logger.Named("Mio")
	bot := &Bot{
		Config: config,
		Logger: logger,
	}

	bot.EventEmitter = NewEventEmitter()
	bot.Callbacks = NewCallbackManager()
	bot.ModuleManager = NewModuleManager(logger)
	bot.Discord = NewDiscord(config.GetString("token"), config.GetInt("shards"), logger)
	bot.MessageProcessor = NewEventHandler(bot, logger)
	bot.InteractionProcessor = NewInteractionProcessor(bot, logger)
	bot.Cooldowns = NewCooldownManager()

	return bot
}

func (b *Bot) UseDefaultHandlers() {
	b.Discord.AddEventHandler(readyHandler(b))
	b.Discord.AddEventHandler(guildJoinHandler(b))
	b.Discord.AddEventHandler(guildLeaveHandler(b))
	b.Discord.AddEventHandler(memberChunkHandler(b))
}

func (b *Bot) Run(ctx context.Context) error {
	b.Logger.Info("Starting up...")
	go b.MessageProcessor.Listen(ctx)
	go b.InteractionProcessor.Listen(ctx)
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
