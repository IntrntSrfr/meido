package searchmod

import (
	"fmt"
	"github.com/intrntsrfr/meido/base"
	"github.com/intrntsrfr/meido/internal/services"
	"strings"
	"sync"
)

type SearchMod struct {
	sync.Mutex
	name         string
	commands     map[string]*base.ModCommand
	allowedTypes base.MessageType
	allowDMs     bool
	search       *services.SearchService
}

func New(s *services.SearchService) base.Mod {
	return &SearchMod{
		name:         "Search",
		commands:     make(map[string]*base.ModCommand),
		allowedTypes: base.MessageTypeCreate,
		allowDMs:     true,
		search:       s,
	}
}

func (m *SearchMod) Name() string {
	return m.name
}
func (m *SearchMod) Passives() []*base.ModPassive {
	return []*base.ModPassive{}
}
func (m *SearchMod) Commands() map[string]*base.ModCommand {
	return m.commands
}
func (m *SearchMod) AllowedTypes() base.MessageType {
	return m.allowedTypes
}
func (m *SearchMod) AllowDMs() bool {
	return m.allowDMs
}
func (m *SearchMod) Hook() error {
	m.RegisterCommand(NewYouTubeCommand(m))
	return nil
}
func (m *SearchMod) RegisterCommand(cmd *base.ModCommand) {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.commands[cmd.Name]; ok {
		panic(fmt.Sprintf("command '%v' already exists in %v", cmd.Name, m.Name()))
	}
	m.commands[cmd.Name] = cmd
}

func NewYouTubeCommand(m *SearchMod) *base.ModCommand {
	return &base.ModCommand{
		Mod:           m,
		Name:          "youtube",
		Description:   "Search for a YouTube video",
		Triggers:      []string{"m?youtube", "m?yt"},
		Usage:         "m?yt deez nuts",
		Cooldown:      2,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  base.MessageTypeCreate,
		AllowDMs:      true,
		Enabled:       true,
		Run:           m.youtubeCommand,
	}
}
func (m *SearchMod) youtubeCommand(msg *base.DiscordMessage) {
	if msg.LenArgs() < 2 {
		return
	}

	query := strings.Join(msg.Args()[1:], " ")
	ids, err := m.search.SearchGoogleImages(query)
	if err != nil {
		msg.Reply("Could not fetch images :(")
	}

	if len(ids) < 1 {
		msg.Reply("I got no results for that :(")
		return
	}

	msg.Reply("https://youtube.com/watch?v=" + ids[0])
}
