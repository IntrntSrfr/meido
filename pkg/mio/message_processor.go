package mio

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"go.uber.org/zap"
)

type MessageProcessor struct {
	Bot       *Bot
	Cooldowns *CooldownManager
	Callbacks *CallbackManager
	logger    *zap.Logger
}

func NewMessageProcessor(bot *Bot, logger *zap.Logger) *MessageProcessor {
	return &MessageProcessor{
		Bot:       bot,
		Cooldowns: NewCooldownManager(),
		Callbacks: bot.Callbacks,
		logger:    logger.Named("MessageProcessor"),
	}
}

func (mp *MessageProcessor) ListenMessages(ctx context.Context) {
	for {
		select {
		case msg := <-mp.Bot.Discord.messageChan:
			go mp.DeliverCallbacks(msg)
			go mp.ProcessMessage(msg)
		case <-ctx.Done():
			return
		}
	}
}

func (mp *MessageProcessor) ProcessMessage(msg *DiscordMessage) {
	for _, mod := range mp.Bot.Modules {
		if !mod.AllowsMessage(msg) {
			return
		}

		for _, pas := range mod.Passives() {
			if !pas.Enabled || msg.Type()&pas.AllowedTypes == 0 {
				continue
			}
			go pas.Run(msg)
		}

		if len(msg.Args()) <= 0 {
			continue
		}

		if cmd, err := mod.FindCommandByTriggers(msg.RawContent()); err == nil {
			mp.ProcessCommand(cmd, msg)
		}
	}
}

func (mp *MessageProcessor) ProcessCommand(cmd *ModuleCommand, msg *DiscordMessage) {
	if !cmd.IsEnabled || !cmd.AllowsMessage(msg) {
		return
	}

	if cmd.RequiresUserType == UserTypeBotOwner && !mp.Bot.IsOwner(msg.AuthorID()) {
		_, _ = msg.Reply("This command is owner only")
		return
	}

	if cdKey := cmd.CooldownKey(msg); cdKey != "" {
		if t, ok := mp.Cooldowns.Check(cdKey); ok {
			_, _ = msg.ReplyAndDelete(fmt.Sprintf("This command is on cooldown for another %v", t), time.Second*2)
			return
		}
		mp.Cooldowns.Set(cdKey, time.Duration(cmd.Cooldown))
	}
	mp.RunCommand(cmd, msg)
}

// if a command causes panic, this will surely keep everything from crashing
func (mp *MessageProcessor) RunCommand(cmd *ModuleCommand, msg *DiscordMessage) {
	defer func() {
		if r := recover(); r != nil {
			mp.logger.Warn("Recovery needed", zap.Any("error", r))
			mp.Bot.Emit(BotEventCommandPanicked, &CommandPanicked{cmd, msg, string(debug.Stack())})
			_, _ = msg.Reply("Something terrible happened. Please try again. If that does not work, send a DM to bot dev(s)")
		}
	}()

	cmd.Run(msg)
	mp.Bot.Emit(BotEventCommandRan, &CommandRan{cmd, msg})
	mp.logger.Info("Command",
		zap.String("id", msg.ID()),
		zap.String("content", msg.RawContent()),
		zap.String("userID", msg.AuthorID()),
	)
}

func (mp *MessageProcessor) DeliverCallbacks(msg *DiscordMessage) {
	if msg.Type() != MessageTypeCreate {
		return
	}

	ch, err := mp.Callbacks.Get(msg.CallbackKey())
	if err != nil {
		return
	}
	ch <- msg
}
