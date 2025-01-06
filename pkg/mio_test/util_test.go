package mio_test

import (
	"errors"
	"fmt"
	"image"
	"io"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/pkg/mio"
	"github.com/intrntsrfr/meido/pkg/mio/test"
	"github.com/intrntsrfr/meido/pkg/utils"
)

func NewTestConfig() *utils.Config {
	conf := utils.NewConfig()
	conf.Set("shards", 1)
	conf.Set("token", "asdf")
	return conf
}

func NewTestDiscord(conf *utils.Config, sess mio.DiscordSession, logger mio.Logger) *mio.Discord {
	if conf == nil {
		conf = NewTestConfig()
	}
	if sess == nil {
		sess = NewDiscordSession(conf.GetString("token"), conf.GetInt("shards"))
	}
	if logger == nil {
		logger = mio.NewLogger(io.Discard)
	}
	d := mio.NewDiscord(conf.GetString("token"), conf.GetInt("shards"), logger)
	d.Sess = sess
	d.Sessions = []mio.DiscordSession{d.Sess}
	return d
}

func NewTestBot() *mio.Bot {
	bot := mio.NewBotBuilder(test.NewTestConfig()).
		WithDefaultHandlers().
		WithDiscord(NewTestDiscord(nil, nil, nil)).
		WithLogger(mio.NewDiscardLogger()).
		Build()
	return bot
}

func NewTestModule(bot *mio.Bot, name string, log mio.Logger) *testModule {
	return &testModule{ModuleBase: *mio.NewModule(bot, name, log)}
}

type testModule struct {
	mio.ModuleBase
	hookShouldFail bool
}

func (m *testModule) Hook() error {
	if m.hookShouldFail {
		return errors.New("Something terrible has happened")
	}
	return nil
}

func NewTestCommand(mod mio.Module) *mio.ModuleCommand {
	return mio.NewModuleCommandBuilder(mod, "test").
		Description("testing").
		Triggers(".test").
		Usage(".test").
		Cooldown(0, mio.CooldownScopeChannel).
		AllowedTypes(mio.MessageTypeCreate).
		Execute(testCommandRun).
		Build()
}

func testCommandRun(msg *mio.DiscordMessage) {

}

func NewTestPassive(mod mio.Module) *mio.ModulePassive {
	return mio.NewModulePassiveBuilder(mod, "test").
		Description("testing").
		AllowedTypes(mio.MessageTypeCreate).
		Execute(testPassiveRun).
		Build()
}

func testPassiveRun(msg *mio.DiscordMessage) {

}

func NewTestApplicationCommand(mod mio.Module) *mio.ModuleApplicationCommand {
	return mio.NewModuleApplicationCommandBuilder(mod, "test").
		Type(discordgo.ChatApplicationCommand).
		Description("testing").
		Execute(testApplicationCommandRun).
		Build()
}

func testApplicationCommandRun(msg *mio.DiscordApplicationCommand) {

}

func NewTestMessageComponent(mod mio.Module) *mio.ModuleMessageComponent {
	return &mio.ModuleMessageComponent{
		Mod:           mod,
		Name:          "test",
		Cooldown:      0,
		CooldownScope: mio.CooldownScopeChannel,
		CheckBotPerms: false,
		Enabled:       true,
		Execute:       testMessageComponentRun,
	}
}

func testMessageComponentRun(msg *mio.DiscordMessageComponent) {

}

func NewTestModalSubmit(mod mio.Module) *mio.ModuleModalSubmit {
	return &mio.ModuleModalSubmit{
		Mod:     mod,
		Name:    "test",
		Enabled: true,
		Execute: testModalSubmitRun,
	}
}

func testModalSubmitRun(msg *mio.DiscordModalSubmit) {

}

func NewTestMessage(bot *mio.Bot, guildID string) *mio.DiscordMessage {
	author := &discordgo.User{Username: "jeff"}
	msg := &mio.DiscordMessage{
		Sess:        bot.Discord.Sess,
		Discord:     bot.Discord,
		MessageType: mio.MessageTypeCreate,
		Message: &discordgo.Message{
			Content:   ".test hello",
			GuildID:   guildID,
			Author:    author,
			ChannelID: "1",
			ID:        "1",
		},
	}
	if guildID != "" {
		msg.Message.Member = &discordgo.Member{User: author}
	}
	return msg
}

func NewTestApplicationCommandInteraction(bot *mio.Bot, guildID string) *mio.DiscordInteraction {
	it := newTestInteraction(bot, guildID)
	it.Interaction.Type = discordgo.InteractionApplicationCommand
	it.Interaction.Data = discordgo.ApplicationCommandInteractionData{
		Name:        "test",
		CommandType: discordgo.ChatApplicationCommand,
	}
	return it
}

func NewTestMessageComponentInteraction(bot *mio.Bot, guildID, customID string) *mio.DiscordInteraction {
	it := newTestInteraction(bot, guildID)
	it.Interaction.Type = discordgo.InteractionMessageComponent
	it.Interaction.Data = discordgo.MessageComponentInteractionData{
		CustomID: customID,
	}
	return it
}

func NewTestModalSubmitInteraction(bot *mio.Bot, guildID, customID string) *mio.DiscordInteraction {
	it := newTestInteraction(bot, guildID)
	it.Interaction.Type = discordgo.InteractionModalSubmit
	it.Interaction.Data = discordgo.ModalSubmitInteractionData{
		CustomID: customID,
	}
	return it
}

func newTestInteraction(bot *mio.Bot, guildID string) *mio.DiscordInteraction {
	author := &discordgo.User{Username: "jeff"}
	it := &mio.DiscordInteraction{
		Sess:    bot.Discord.Sess,
		Discord: bot.Discord,
		Interaction: &discordgo.Interaction{
			ChannelID: "1",
			GuildID:   guildID,
			ID:        "1",
		},
	}
	if guildID == "" {
		it.Interaction.User = author
	} else {
		it.Interaction.Member = &discordgo.Member{User: author}
	}
	return it
}

type DiscordSessionMock struct {
	token      string
	shardID    int
	shardCount int
	identify   discordgo.Identify
	state      *discordgo.State

	handlersMu   sync.RWMutex
	handlers     map[string][]interface{}
	onceHandlers map[string][]interface{}

	IsOpen          bool
	CloseShouldFail bool
}

func NewDiscordSession(token string, shards int) *DiscordSessionMock {
	s := &DiscordSessionMock{
		token:        token,
		shardID:      0,
		shardCount:   shards,
		IsOpen:       false,
		state:        discordgo.NewState(),
		handlers:     make(map[string][]interface{}, 0),
		onceHandlers: make(map[string][]interface{}, 0),
	}
	return s
}

func (s *DiscordSessionMock) Open() error {
	if s.IsOpen {
		return errors.New("session is already open")
	}
	s.IsOpen = true
	s.state.User = &discordgo.User{
		ID:       "1",
		Username: "Mio",
		Bot:      true,
	}
	return nil
}

func (s *DiscordSessionMock) Close() error {
	if s.CloseShouldFail {
		return errors.New("failed to close session")
	}
	return nil
}

func (s *DiscordSessionMock) ShardID() int {
	return s.shardID
}

func (s *DiscordSessionMock) State() *discordgo.State {
	return s.state
}

func (s *DiscordSessionMock) Real() *discordgo.Session {
	return nil
}

func handlerForInterface(ifc interface{}) string {
	switch ifc.(type) {
	case func(s *discordgo.Session, r *discordgo.Ready):
		return "ready"
	case func(s *discordgo.Session, g *discordgo.GuildCreate):
		return "guildCreate"
	case func(s *discordgo.Session, g *discordgo.GuildDelete):
		return "guildDelete"
	case func(s *discordgo.Session, g *discordgo.GuildMembersChunk):
		return "guildMembersChunk"
	}
	return ""
}

func (s *DiscordSessionMock) AddHandler(handler interface{}) func() {
	typeStr := handlerForInterface(handler)
	s.handlersMu.Lock()
	defer s.handlersMu.Unlock()

	s.handlers[typeStr] = append(s.handlers[typeStr], handler)
	return func() {}
}

func (s *DiscordSessionMock) AddHandlerOnce(handler interface{}) func() {
	typeStr := handlerForInterface(handler)
	s.handlersMu.Lock()
	defer s.handlersMu.Unlock()

	s.onceHandlers[typeStr] = append(s.onceHandlers[typeStr], handler)
	return func() {}
}

func (s *DiscordSessionMock) ApplicationCommandBulkOverwrite(appID string, guildID string, commands []*discordgo.ApplicationCommand, options ...discordgo.RequestOption) (createdCommands []*discordgo.ApplicationCommand, err error) {
	for _, c := range commands {
		if strings.ToLower(c.Name) != c.Name {
			return nil, errors.New("lower case name")
		}
		// can check options too, but im lazy
	}
	return commands, nil
}

func (s *DiscordSessionMock) Channel(channelID string, options ...discordgo.RequestOption) (st *discordgo.Channel, err error) {
	return s.state.Channel(channelID)
}

func (s *DiscordSessionMock) ChannelFileSend(channelID string, name string, r io.Reader, options ...discordgo.RequestOption) (*discordgo.Message, error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) ChannelMessageDelete(channelID string, messageID string, options ...discordgo.RequestOption) (err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) ChannelMessageEdit(channelID string, messageID string, content string, options ...discordgo.RequestOption) (*discordgo.Message, error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) ChannelMessageEditComplex(m *discordgo.MessageEdit, options ...discordgo.RequestOption) (st *discordgo.Message, err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) ChannelMessageEditEmbed(channelID string, messageID string, embed *discordgo.MessageEmbed, options ...discordgo.RequestOption) (*discordgo.Message, error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) ChannelMessageEditEmbeds(channelID string, messageID string, embeds []*discordgo.MessageEmbed, options ...discordgo.RequestOption) (*discordgo.Message, error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) ChannelMessagePin(channelID string, messageID string, options ...discordgo.RequestOption) (err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) ChannelMessageSend(channelID string, content string, options ...discordgo.RequestOption) (*discordgo.Message, error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) ChannelMessageSendComplex(channelID string, data *discordgo.MessageSend, options ...discordgo.RequestOption) (st *discordgo.Message, err error) {
	if !s.IsOpen {
		return nil, errors.New("session is closed")
	}

	channel, err := s.Channel(channelID)
	if err != nil {
		return nil, errors.New("channel not found")
	}

	guildID := ""
	if guild, err := s.Guild(channel.GuildID); err == nil {
		guildID = guild.ID
	}

	message := &discordgo.Message{
		ID:        fmt.Sprintf("%d", rand.Int()), // Generate a random ID or use a counter for simplicity
		ChannelID: channelID,
		GuildID:   guildID,
		Content:   data.Content,
		Author:    &discordgo.User{ID: s.State().User.ID}, // Assuming the author is the user represented by this session
		Timestamp: time.Now(),
	}

	s.state.MessageAdd(message)
	return message, nil
}

func (s *DiscordSessionMock) ChannelMessageSendEmbed(channelID string, embed *discordgo.MessageEmbed, options ...discordgo.RequestOption) (*discordgo.Message, error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) ChannelMessageSendEmbedReply(channelID string, embed *discordgo.MessageEmbed, reference *discordgo.MessageReference, options ...discordgo.RequestOption) (*discordgo.Message, error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) ChannelMessageSendEmbeds(channelID string, embeds []*discordgo.MessageEmbed, options ...discordgo.RequestOption) (*discordgo.Message, error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) ChannelMessageSendEmbedsReply(channelID string, embeds []*discordgo.MessageEmbed, reference *discordgo.MessageReference, options ...discordgo.RequestOption) (*discordgo.Message, error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) ChannelMessageSendReply(channelID string, content string, reference *discordgo.MessageReference, options ...discordgo.RequestOption) (*discordgo.Message, error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) ChannelMessages(channelID string, limit int, beforeID string, afterID string, aroundID string, options ...discordgo.RequestOption) (st []*discordgo.Message, err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) ChannelMessagesBulkDelete(channelID string, messages []string, options ...discordgo.RequestOption) (err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) ChannelPermissionSet(channelID, targetID string, targetType discordgo.PermissionOverwriteType, allow, deny int64, options ...discordgo.RequestOption) (err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) ChannelTyping(channelID string, options ...discordgo.RequestOption) (err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) Guild(guildID string, options ...discordgo.RequestOption) (st *discordgo.Guild, err error) {
	return s.state.Guild(guildID)
}

func (s *DiscordSessionMock) GuildBanCreate(guildID string, userID string, days int, options ...discordgo.RequestOption) (err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) GuildBanCreateWithReason(guildID string, userID string, reason string, days int, options ...discordgo.RequestOption) (err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) GuildBanDelete(guildID string, userID string, options ...discordgo.RequestOption) (err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) GuildBans(guildID string, limit int, beforeID string, afterID string, options ...discordgo.RequestOption) (st []*discordgo.GuildBan, err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) GuildChannels(guildID string, options ...discordgo.RequestOption) (st []*discordgo.Channel, err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) GuildIcon(guildID string, options ...discordgo.RequestOption) (img image.Image, err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) GuildMember(guildID string, userID string, options ...discordgo.RequestOption) (st *discordgo.Member, err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) GuildMemberAdd(guildID string, userID string, data *discordgo.GuildMemberAddParams, options ...discordgo.RequestOption) (err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) GuildMemberDelete(guildID string, userID string, options ...discordgo.RequestOption) (err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) GuildMemberDeleteWithReason(guildID string, userID string, reason string, options ...discordgo.RequestOption) (err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) GuildMemberRoleAdd(guildID string, userID string, roleID string, options ...discordgo.RequestOption) (err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) GuildMemberRoleRemove(guildID string, userID string, roleID string, options ...discordgo.RequestOption) (err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) GuildMemberTimeout(guildID string, userID string, until *time.Time, options ...discordgo.RequestOption) (err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) GuildMembers(guildID string, after string, limit int, options ...discordgo.RequestOption) (st []*discordgo.Member, err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) GuildRoleCreate(guildID string, data *discordgo.RoleParams, options ...discordgo.RequestOption) (st *discordgo.Role, err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) GuildRoleDelete(guildID string, roleID string, options ...discordgo.RequestOption) (err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) GuildRoleEdit(guildID string, roleID string, data *discordgo.RoleParams, options ...discordgo.RequestOption) (st *discordgo.Role, err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) GuildRoles(guildID string, options ...discordgo.RequestOption) (st []*discordgo.Role, err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) GuildSplash(guildID string, options ...discordgo.RequestOption) (img image.Image, err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) RequestGuildMembers(guildID string, query string, limit int, nonce string, presences bool) error {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) RequestGuildMembersBatch(guildIDs []string, query string, limit int, nonce string, presences bool) (err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) RequestGuildMembersBatchList(guildIDs []string, userIDs []string, limit int, nonce string, presences bool) (err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) RequestGuildMembersList(guildID string, userIDs []string, limit int, nonce string, presences bool) error {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) User(userID string, options ...discordgo.RequestOption) (st *discordgo.User, err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) UserChannelCreate(recipientID string, options ...discordgo.RequestOption) (st *discordgo.Channel, err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) UpdateStatusComplex(usd discordgo.UpdateStatusData) (err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) InteractionRespond(interaction *discordgo.Interaction, resp *discordgo.InteractionResponse, options ...discordgo.RequestOption) error {
	panic("not implemented")
}
