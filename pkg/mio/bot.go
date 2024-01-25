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
	MessageProcessor *MessageProcessor
	Callbacks        *CallbackManager
	*EventManager

	Log *zap.Logger
}

func NewBot(config Configurable, log *zap.Logger) *Bot {
	bot := &Bot{
		Config: config,
		Log:    log,
	}

	bot.EventManager = NewEventManager()
	bot.Callbacks = NewCallbackManager()
	bot.ModuleManager = NewModuleManager(log)
	bot.Discord = NewDiscord(config.GetString("token"), config.GetInt("shards"), log)
	bot.MessageProcessor = NewMessageProcessor(bot, log)

	return bot
}

func (b *Bot) UseDefaultHandlers() {
	b.Discord.AddEventHandler(readyHandler(b))
	b.Discord.AddEventHandler(guildJoinHandler(b))
	b.Discord.AddEventHandler(guildLeaveHandler(b))
	b.Discord.AddEventHandler(memberChunkHandler(b))
}

func (b *Bot) Run(ctx context.Context) error {
	b.Log.Info("starting bot")
	go b.MessageProcessor.ListenMessages(ctx)
	return b.Discord.Run()
}

func (b *Bot) Close() {
	b.Log.Info("stopping bot")
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
