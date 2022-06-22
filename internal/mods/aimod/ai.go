package aimod

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/intrntsrfr/meido/base"
	gogpt "github.com/sashabaranov/go-gpt3"
)

// AIMod represents the ping mod
type AIMod struct {
	sync.Mutex
	name         string
	commands     map[string]*base.ModCommand
	allowedTypes base.MessageType
	allowDMs     bool
	gptClient    *gogpt.Client
	engine       string
}

// New returns a new TestMod.
func New(gptClient *gogpt.Client, engine string) base.Mod {
	return &AIMod{
		name:         "AI",
		commands:     make(map[string]*base.ModCommand),
		allowedTypes: base.MessageTypeCreate,
		allowDMs:     true,
		gptClient:    gptClient,
		engine:       engine,
	}
}

// Name returns the name of the mod.
func (m *AIMod) Name() string {
	return m.name
}

// Passives returns the mod passives.
func (m *AIMod) Passives() []*base.ModPassive {
	return []*base.ModPassive{}
}

// Commands returns the mod commands.
func (m *AIMod) Commands() map[string]*base.ModCommand {
	return m.commands
}

// AllowedTypes returns the allowed MessageTypes.
func (m *AIMod) AllowedTypes() base.MessageType {
	return m.allowedTypes
}

// AllowDMs returns whether the mod allows DMs.
func (m *AIMod) AllowDMs() bool {
	return m.allowDMs
}

// Hook will hook the Mod into the Bot.
func (m *AIMod) Hook() error {
	m.RegisterCommand(NewPromptCommand(m))
	//m.RegisterCommand(NewMonkeyCommand(m))

	return nil
}

// RegisterCommand registers a ModCommand to the Mod
func (m *AIMod) RegisterCommand(cmd *base.ModCommand) {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.commands[cmd.Name]; ok {
		panic(fmt.Sprintf("command '%v' already exists in %v", cmd.Name, m.Name()))
	}
	m.commands[cmd.Name] = cmd
}

// NewPromptCommand returns a new ping command.
func NewPromptCommand(m *AIMod) *base.ModCommand {
	return &base.ModCommand{
		Mod:           m,
		Name:          "prompt",
		Description:   "Generate completions for a prompt using GPT-3.",
		Triggers:      []string{"m?prompt"},
		Usage:         "m?prompt tell me about maids.",
		Cooldown:      15,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  base.MessageTypeCreate,
		AllowDMs:      true,
		Enabled:       true,
		Run:           m.promptCommand,
	}
}

func (m *AIMod) promptCommand(msg *base.DiscordMessage) {
	if len(msg.Args()) < 2 {
		return
	}

	prompt := strings.Join(msg.Args()[1:], " ")

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

	msg.Reply(resp.Choices[0].Text)
}
