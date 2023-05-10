package ai

import (
	"context"
	"github.com/intrntsrfr/meido/pkg/mio"
	gogpt "github.com/sashabaranov/go-gpt3"
	"go.uber.org/zap"
	"strings"
)

type Module struct {
	*mio.ModuleBase
	gptClient *gogpt.Client
	engine    string
}

func New(bot *mio.Bot, logger *zap.Logger, gptClient *gogpt.Client, engine string) mio.Module {
	return &Module{
		ModuleBase: mio.NewModule(bot, "AI", logger),
		gptClient:  gptClient,
		engine:     engine,
	}
}

// Hook will hook the Module into the Bot.
func (m *Module) Hook() error {
	return m.RegisterCommand(newPromptCommand(m))
}

func newPromptCommand(m *Module) *mio.ModuleCommand {
	return &mio.ModuleCommand{
		Mod:           m,
		Name:          "prompt",
		Description:   "Generate completions for a prompt using GPT-3.",
		Triggers:      []string{"m?prompt"},
		Usage:         "m?prompt tell me about maids.",
		Cooldown:      15,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  mio.MessageTypeCreate,
		AllowDMs:      true,
		Enabled:       true,
		Run: func(msg *mio.DiscordMessage) {
			if len(msg.Args()) < 2 {
				return
			}

			prompt := strings.Join(msg.Args()[1:], " ")
			if strings.TrimSpace(prompt) == "" {
				return
			}

			req := gogpt.CompletionRequest{
				MaxTokens:   128,
				Prompt:      prompt,
				Temperature: 0.9,
			}
			resp, err := m.gptClient.CreateCompletion(context.Background(), m.engine, req)
			if err != nil {
				_, _ = msg.Reply("Could not create completion.")
			}

			if len(resp.Choices) < 1 {
				_, _ = msg.Reply("Could not create completion.")
				return
			}

			_, _ = msg.Reply(resp.Choices[0].Text)
		},
	}
}
