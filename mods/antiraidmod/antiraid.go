package antiraidmod

import (
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	base2 "github.com/intrntsrfr/meido/base"
	"github.com/intrntsrfr/owo"
	"golang.org/x/time/rate"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// AntiRaidMod represents the antiraid mod
type AntiRaidMod struct {
	sync.Mutex
	name         string
	commands     map[string]*base2.ModCommand
	allowedTypes base2.MessageType
	allowDMs     bool
	owo          *owo.Client

	servers *serverMap
	banChan chan [2]string
}

// New returns a new AntiRaidMod.
func New(n string) base2.Mod {
	return &AntiRaidMod{
		name:         n,
		commands:     make(map[string]*base2.ModCommand),
		allowedTypes: base2.MessageTypeCreate,
		allowDMs:     true,
		servers:      &serverMap{Servers: make(map[string]*server)},
		banChan:      make(chan [2]string, 1024),
	}
}

// Name returns the name of the mod.
func (m *AntiRaidMod) Name() string {
	return m.name
}

// Save saves the mod state to a file.
func (m *AntiRaidMod) Save() error {
	data, err := json.Marshal(m.servers)
	if err != nil {
		return err
	}
	return ioutil.WriteFile("./data/fish", data, os.ModePerm)
}

// Load loads the mod state from a file.
func (m *AntiRaidMod) Load() error {
	if _, err := os.Stat("./data/fish"); err != nil {
		return nil
	}
	data, err := ioutil.ReadFile("./data/fish")
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &m.servers)
}

// Passives returns the mod passives.
func (m *AntiRaidMod) Passives() []*base2.ModPassive {
	return []*base2.ModPassive{}
}

// Commands returns the mod commands.
func (m *AntiRaidMod) Commands() map[string]*base2.ModCommand {
	return m.commands
}

// AllowedTypes returns the allowed MessageTypes.
func (m *AntiRaidMod) AllowedTypes() base2.MessageType {
	return m.allowedTypes
}

// AllowDMs returns whether the mod allows DMs.
func (m *AntiRaidMod) AllowDMs() bool {
	return m.allowDMs
}

// Hook will hook the Mod into the Bot.
func (m *AntiRaidMod) Hook(b *base2.Bot) error {
	m.owo = b.Owo

	err := m.Load()
	if err != nil {
		return err
	}

	b.Discord.Sess.AddHandler(func(s *discordgo.Session, g *discordgo.GuildCreate) {
		m.servers.Add(g.ID)
	})

	b.Discord.Sess.AddHandler(func(s *discordgo.Session, g *discordgo.GuildDelete) {
		m.servers.Remove(g.ID)
	})

	go m.runBanListener(b.Discord.Sess)

	return nil
}

// RegisterCommand registers a ModCommand to the Mod
func (m *AntiRaidMod) RegisterCommand(cmd *base2.ModCommand) {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.commands[cmd.Name]; ok {
		panic(fmt.Sprintf("command '%v' already exists in %v", cmd.Name, m.Name()))
	}
	m.commands[cmd.Name] = cmd
}

func (m *AntiRaidMod) runBanListener(s *discordgo.Session) {
	for {
		select {
		case ban := <-m.banChan:
			//fmt.Println("time to ban: ", ban)
			s.GuildBanCreateWithReason(ban[0], ban[1], "Raid prevention", 7)
		}
	}
}

func (m *AntiRaidMod) GuildMemberAddHandler(s *discordgo.Session, mem *discordgo.GuildMemberAdd) {

	srv, ok := m.servers.Get(mem.GuildID)
	if !ok {
		return
	}

	srv.AddToJoinCache(mem.User.ID)

	if !srv.RaidMode() {
		return
	}

	if isNewAccount(mem.User.ID) {
		srv.lastRaid[mem.User.ID] = struct{}{}
		m.banChan <- [2]string{mem.GuildID, mem.User.ID}
	}
}
func (m *AntiRaidMod) AutoDetectHandler(s *discordgo.Session, mem *discordgo.GuildMemberAdd) {

	srv, ok := m.servers.Get(mem.GuildID)
	if !ok {
		return
	}

	if srv.RaidMode() || !srv.AutoDetect {
		return
	}

	if !srv.joinLimiter.Allow() {
		srv.RaidToggle(m)
		time.AfterFunc(time.Minute*5, func() {
			if srv.raidMode {
				srv.RaidToggle(m)
			}
		})
	}
}

func (m *AntiRaidMod) RaidToggleHandler(s *discordgo.Session, msg *discordgo.MessageCreate) {

	if msg.Author.Bot || len(msg.Content) <= 0 {
		return
	}

	srv, ok := m.servers.Get(msg.GuildID)
	if !ok {
		return
	}

	perms, err := s.State.UserChannelPermissions(msg.Author.ID, msg.ChannelID)
	if err != nil {
		return
	}
	botPerms, err := s.State.UserChannelPermissions(s.State.User.ID, msg.ChannelID)
	if err != nil {
		return
	}

	if perms&discordgo.PermissionBanMembers == 0 && perms&discordgo.PermissionAdministrator == 0 {
		return
	}

	if botPerms&discordgo.PermissionBanMembers == 0 && perms&discordgo.PermissionAdministrator == 0 {
		return
	}

	args := strings.Fields(msg.Content)

	switch strings.ToLower(args[0]) {
	case "m?raidmode":
		srv.RaidToggle(m)
		s.ChannelMessageSend(msg.ChannelID, fmt.Sprintf("raid mode set to %v", srv.RaidMode()))
	case "m?lastraid":
		l := srv.GetLastRaid()
		if len(l) <= 0 {
			s.ChannelMessageSend(msg.ChannelID, "no last raid")
			return
		}
		res, err := m.owo.Upload(strings.Join(l, " "))
		if err != nil {
			s.ChannelMessageSend(msg.ChannelID, "Error getting last raid. try again?")
			return
		}
		s.ChannelMessageSend(msg.ChannelID, res)
	case "m?raidignore":
		if len(args) < 2 {
			return
		}
		srv.IgnoreRole = args[1]
		s.ChannelMessageSend(msg.ChannelID, fmt.Sprintf("raid will ignore users with role id: %v", args[1]))
	case "m?autoraid":
		srv.ToggleAutodetect()
		s.ChannelMessageSend(msg.ChannelID, fmt.Sprintf("raid autodetect set to: %v", srv.AutoDetect))
	default:
		return
	}
}

func (m *AntiRaidMod) MessageCreateHandler(s *discordgo.Session, msg *discordgo.MessageCreate) {
	if msg.Author.Bot {
		return
	}

	srv, ok := m.servers.Get(msg.GuildID)
	if !ok {
		return
	}

	if !srv.RaidMode() {
		return
	}

	if hasRole(msg.Member, srv.IgnoreRole) {
		return
	}

	usr, ok := srv.GetUser(msg.Author.ID)
	if !ok {
		srv.Add(msg.Author.ID)
		return
	}

	if !usr.Allow() || len(msg.Mentions) > 10 {
		// ban the user
		//fmt.Println("bad user", m.GuildID, m.Author.ID)
		srv.lastRaid[msg.Author.ID] = struct{}{}
		m.banChan <- [2]string{msg.GuildID, msg.Author.ID}
		//s.GuildBanCreateWithReason(m.GuildID, m.Author.ID, "Raid measure", 7)
	}
}

func isNewAccount(userID string) bool {

	id, err := strconv.ParseInt(userID, 0, 63)
	if err != nil {
		return false
	}

	id = ((id >> 22) + 1420070400000) / 1000

	// how long time should be acceptable, currently set to 2 days
	threshold := time.Now().Add(-1 * time.Hour * 24 * 2)

	ts := time.Unix(id, 0)

	return ts.Unix() > threshold.Unix()
}

func hasRole(m *discordgo.Member, role string) bool {
	if role == "" {
		return false
	}
	for _, r := range m.Roles {
		if r == role {
			return true
		}
	}
	return false
}

type serverMap struct {
	sync.RWMutex
	Servers map[string]*server
}

func (s *serverMap) Add(id string) {
	s.Lock()
	defer s.Unlock()
	srv, found := s.Servers[id]
	if !found {
		s.Servers[id] = &server{
			id:          id,
			raidMode:    false,
			users:       make(map[string]*rate.Limiter),
			joinedCache: make(map[string]*cacheUser),
			lastRaid:    make(map[string]struct{}),
			joinLimiter: rate.NewLimiter(2, 10),
			AutoDetect:  false,
		}
	} else {
		s.Servers[id] = &server{
			id:          id,
			raidMode:    false,
			users:       make(map[string]*rate.Limiter),
			joinedCache: make(map[string]*cacheUser),
			lastRaid:    make(map[string]struct{}),
			IgnoreRole:  srv.IgnoreRole,
			joinLimiter: rate.NewLimiter(0.25, 5),
			AutoDetect:  srv.AutoDetect,
		}
	}
	//fmt.Println(fmt.Sprintf("added server id: %v", id))
}
func (s *serverMap) Remove(id string) {
	s.Lock()
	defer s.Unlock()
	delete(s.Servers, id)
}
func (s *serverMap) Get(id string) (*server, bool) {
	s.RLock()
	defer s.RUnlock()
	val, ok := s.Servers[id]
	return val, ok
}

type server struct {
	sync.RWMutex
	id          string
	raidMode    bool
	users       map[string]*rate.Limiter
	joinedCache map[string]*cacheUser
	lastRaid    map[string]struct{}
	IgnoreRole  string
	joinLimiter *rate.Limiter
	AutoDetect  bool
}

func (s *server) Add(id string) {
	s.Lock()
	defer s.Unlock()
	s.users[id] = rate.NewLimiter(rate.Every(time.Second*2), 2)
	//fmt.Println(fmt.Sprintf("%v: added user limiter: %v", s.id, id))
}
func (s *server) Remove(id string) {
	s.Lock()
	defer s.Unlock()
	delete(s.users, id)
}
func (s *server) GetUser(id string) (*rate.Limiter, bool) {
	s.RLock()
	defer s.RUnlock()
	val, ok := s.users[id]
	return val, ok
}
func (s *server) Autodetect() bool {
	return s.AutoDetect
}
func (s *server) ToggleAutodetect() {
	s.AutoDetect = !s.AutoDetect
}
func (s *server) RaidMode() bool {
	return s.raidMode
}
func (s *server) RaidToggle(m *AntiRaidMod) {
	s.Lock()
	if s.raidMode {
		// raid mode is being turned off
		s.users = make(map[string]*rate.Limiter)

	} else {
		// raid mode is being turned on

		// new lastraid map
		s.lastRaid = make(map[string]struct{})

		// look through join cache for new bad users and ban said users, and add them to the raidmap
		for _, u := range s.joinedCache {
			if isNewAccount(u.u) {
				s.lastRaid[u.u] = struct{}{}

				delete(s.joinedCache, u.u)

				// ban the user
				m.banChan <- [2]string{s.id, u.u}
				//sess.GuildBanCreateWithReason(s.id, u.u, "Raid measure", 7)
			}
		}

	}
	s.raidMode = !s.raidMode
	s.Unlock()
}
func (s *server) AddToJoinCache(id string) {
	s.Lock()
	defer s.Unlock()
	s.joinedCache[id] = &cacheUser{
		u: id,
		e: time.Now().Add(time.Hour).UnixNano(),
	}
	//fmt.Println(fmt.Sprintf("%v: added user to join cache: %v", s.id, id))
}

func (s *server) GetLastRaid() []string {
	var l []string
	for k := range s.lastRaid {
		l = append(l, k)
	}
	return l
}

func (s *serverMap) removeOld() {
	for _, v := range s.Servers {
		s.Lock()
		for j, v2 := range v.joinedCache {
			if v2.Expired() {
				//fmt.Println(fmt.Sprintf("%v: user expired: %v", i, v2.u))
				delete(v.joinedCache, j)
			}
		}
		s.Unlock()
	}
}

func (s *serverMap) runCleaner() {
	t := time.NewTicker(time.Minute * 5)
	for {
		select {
		case <-t.C:
			s.removeOld()
		}
	}
}

type cacheUser struct {
	u string
	e int64
}

func (c *cacheUser) Expired() bool {
	return time.Now().UnixNano() > c.e
}
