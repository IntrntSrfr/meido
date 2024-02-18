package bot

import (
	"github.com/intrntsrfr/meido/pkg/mio/discord"
	mutils "github.com/intrntsrfr/meido/pkg/mio/utils"
	"github.com/intrntsrfr/meido/pkg/utils"
	"go.uber.org/zap"
)

type BotBuilder struct {
	discord      *discord.Discord
	modules      *ModuleManager
	callbacks    *mutils.CallbackManager
	cooldowns    *mutils.CooldownManager
	eventHandler *EventHandler
	eventEmitter *EventEmitter

	config *utils.Config
	logger *zap.Logger
}

func NewBotBuilder(config *utils.Config, logger *zap.Logger) *BotBuilder {
	logger = logger.Named("Mio")

	return &BotBuilder{
		config: config,
		logger: logger,
	}
}

func (b *BotBuilder) WithDiscord(d *discord.Discord) *BotBuilder {
	b.discord = d
	return b
}

func (b *BotBuilder) Build() *Bot {
	if b.discord == nil {
		b.discord = discord.NewDiscord(b.config.GetString("token"), b.config.GetInt("shards"), b.logger)
	}
	if b.modules == nil {
		b.modules = NewModuleManager(b.logger)
	}
	if b.callbacks == nil {
		b.callbacks = mutils.NewCallbackManager()
	}
	if b.cooldowns == nil {
		b.cooldowns = mutils.NewCooldownManager()
	}
	if b.eventHandler == nil {
		b.eventHandler = NewEventHandler(b.discord, b.modules, b.callbacks, b.logger)
	}
	if b.eventEmitter == nil {
		b.eventEmitter = NewEventEmitter()
	}

	return &Bot{
		Discord:       b.discord,
		ModuleManager: b.modules,
		Callbacks:     b.callbacks,
		Cooldowns:     b.cooldowns,
		EventHandler:  b.eventHandler,
		EventEmitter:  b.eventEmitter,
		Config:        b.config,
		Logger:        b.logger,
	}
}
