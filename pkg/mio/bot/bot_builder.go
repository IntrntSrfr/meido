package bot

import (
	"github.com/intrntsrfr/meido/pkg/mio"
	"github.com/intrntsrfr/meido/pkg/mio/discord"
	mutils "github.com/intrntsrfr/meido/pkg/mio/utils"
	"github.com/intrntsrfr/meido/pkg/utils"
)

type BotBuilder struct {
	discord      *discord.Discord
	modules      *ModuleManager
	callbacks    *mutils.CallbackManager
	cooldowns    *mutils.CooldownManager
	eventHandler *EventHandler
	eventBus     *mio.EventBus

	config *utils.Config
	logger mio.Logger

	useDefaultHandlers bool
}

func NewBotBuilder(config *utils.Config) *BotBuilder {
	return &BotBuilder{
		config: config,
		logger: mio.NewDefaultLogger().Named("Mio"),
	}
}

func (b *BotBuilder) WithDiscord(d *discord.Discord) *BotBuilder {
	b.discord = d
	return b
}

func (b *BotBuilder) WithLogger(log mio.Logger) *BotBuilder {
	b.logger = log
	return b
}

func (b *BotBuilder) WithDefaultHandlers() *BotBuilder {
	b.useDefaultHandlers = true
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
	if b.eventBus == nil {
		b.eventBus = mio.NewEventBus()
	}
	if b.eventHandler == nil {
		b.eventHandler = NewEventHandler(b.discord, b.modules, b.callbacks, b.eventBus, b.logger)
	}
	if b.useDefaultHandlers {
		b.discord.AddEventHandler(readyHandler(b.logger))
		b.discord.AddEventHandler(guildJoinHandler(b.logger))
		b.discord.AddEventHandler(guildLeaveHandler(b.logger))
		b.discord.AddEventHandler(memberChunkHandler(b.logger))
	}

	return &Bot{
		Discord:       b.discord,
		ModuleManager: b.modules,
		Callbacks:     b.callbacks,
		Cooldowns:     b.cooldowns,
		EventHandler:  b.eventHandler,
		EventBus:      b.eventBus,
		Config:        b.config,
		Logger:        b.logger,
	}
}
