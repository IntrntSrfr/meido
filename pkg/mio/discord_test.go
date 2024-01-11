package mio

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/pkg/mio/mocks"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestSessionWrapper_ShardID(t *testing.T) {
	shardID := 2
	s := &SessionWrapper{&discordgo.Session{ShardID: shardID}}

	if got := s.ShardID(); got != shardID {
		t.Errorf("SessionWrapper.ShardID() = %v, want %v", got, shardID)
	}
}

func TestSessionWrapper_State(t *testing.T) {
	state := discordgo.NewState()
	s := &SessionWrapper{&discordgo.Session{State: state}}
	s.State()

	if got := s.State(); !reflect.DeepEqual(got, state) {
		t.Errorf("SessionWrapper.State() = %v, want %v", got, state)
	}
}

func setupLogger() *zap.Logger {
	loggerConfig := zap.NewDevelopmentConfig()
	loggerConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	loggerConfig.OutputPaths = []string{}
	loggerConfig.ErrorOutputPaths = []string{}
	logger, _ := loggerConfig.Build()
	return logger.Named("test")
}

func TestNewDiscord(t *testing.T) {
	token := "Bot asdf"
	shards := 1
	logger := setupLogger()
	d := NewDiscord(token, shards, logger)

	if got := d.token; d.token != token {
		t.Errorf("SessionWrapper.token = %v, want %v", got, token)
	}
	if got := d.shards; d.shards != shards {
		t.Errorf("SessionWrapper.shards = %v, want %v", got, shards)
	}
	if got := d.logger; d.logger == nil {
		t.Errorf("SessionWrapper.shards = %v, want %v", got, logger)
	}
}

func TestDiscord_Open(t *testing.T) {
	d := NewDiscord("asfd", 1, setupLogger())
	if err := d.Open(); err != nil {
		t.Errorf("Discord.Open() error = %v, wantErr %v", err, false)
	}
}

func setupDiscord() *Discord {
	d := NewDiscord("Bot asdf", 1, setupLogger())
	d.Sess = &mocks.DiscordSessionMock{}
	d.Sessions = []DiscordSession{d.Sess}
	return d
}

func TestDiscord_Run(t *testing.T) {
	d := setupDiscord()
	fmt.Println(d.Sess)

	if got := d.Run(); got != nil {
		t.Errorf("Discord.Run() error = %v, wantErr %v", got, false)
	}

	// should fail second time
	if got := d.Run(); got == nil {
		t.Errorf("Discord.Run() error = %v, wantErr %v", got, true)
	}
}

func TestDiscord_Close(t *testing.T) {
	tests := []struct {
		name string
		d    *Discord
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.d.Close()
		})
	}
}

func TestDiscord_botRecover(t *testing.T) {
	type args struct {
		i interface{}
	}
	tests := []struct {
		name string
		d    *Discord
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.d.botRecover(tt.args.i)
		})
	}
}

func TestDiscord_onMessageCreate(t *testing.T) {
	type args struct {
		s *discordgo.Session
		m *discordgo.MessageCreate
	}
	tests := []struct {
		name string
		d    *Discord
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.d.onMessageCreate(tt.args.s, tt.args.m)
		})
	}
}

func TestDiscord_onMessageUpdate(t *testing.T) {
	type args struct {
		s *discordgo.Session
		m *discordgo.MessageUpdate
	}
	tests := []struct {
		name string
		d    *Discord
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.d.onMessageUpdate(tt.args.s, tt.args.m)
		})
	}
}

func TestDiscord_onMessageDelete(t *testing.T) {
	type args struct {
		s *discordgo.Session
		m *discordgo.MessageDelete
	}
	tests := []struct {
		name string
		d    *Discord
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.d.onMessageDelete(tt.args.s, tt.args.m)
		})
	}
}

func TestDiscord_BotUser(t *testing.T) {
	tests := []struct {
		name string
		d    *Discord
		want *discordgo.User
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.d.BotUser(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Discord.BotUser() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDiscord_UserChannelPermissions(t *testing.T) {
	type args struct {
		userID    string
		channelID string
	}
	tests := []struct {
		name    string
		d       *Discord
		args    args
		want    int64
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.d.UserChannelPermissions(tt.args.userID, tt.args.channelID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Discord.UserChannelPermissions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Discord.UserChannelPermissions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDiscord_BotHasPermissions(t *testing.T) {
	type args struct {
		channelID string
		perm      int64
	}
	tests := []struct {
		name    string
		d       *Discord
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.d.BotHasPermissions(tt.args.channelID, tt.args.perm)
			if (err != nil) != tt.wantErr {
				t.Errorf("Discord.BotHasPermissions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Discord.BotHasPermissions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDiscord_HasPermissions(t *testing.T) {
	type args struct {
		channelID string
		userID    string
		perm      int64
	}
	tests := []struct {
		name    string
		d       *Discord
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.d.HasPermissions(tt.args.channelID, tt.args.userID, tt.args.perm)
			if (err != nil) != tt.wantErr {
				t.Errorf("Discord.HasPermissions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Discord.HasPermissions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDiscord_HighestRole(t *testing.T) {
	type args struct {
		gid string
		uid string
	}
	tests := []struct {
		name string
		d    *Discord
		args args
		want *discordgo.Role
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.d.HighestRole(tt.args.gid, tt.args.uid); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Discord.HighestRole() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDiscord_HighestRolePosition(t *testing.T) {
	type args struct {
		gid string
		uid string
	}
	tests := []struct {
		name string
		d    *Discord
		args args
		want int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.d.HighestRolePosition(tt.args.gid, tt.args.uid); got != tt.want {
				t.Errorf("Discord.HighestRolePosition() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDiscord_HighestColor(t *testing.T) {
	type args struct {
		gid string
		uid string
	}
	tests := []struct {
		name string
		d    *Discord
		args args
		want int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.d.HighestColor(tt.args.gid, tt.args.uid); got != tt.want {
				t.Errorf("Discord.HighestColor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoleByPos_Len(t *testing.T) {
	tests := []struct {
		name string
		a    RoleByPos
		want int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.Len(); got != tt.want {
				t.Errorf("RoleByPos.Len() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoleByPos_Swap(t *testing.T) {
	type args struct {
		i int
		j int
	}
	tests := []struct {
		name string
		a    RoleByPos
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.a.Swap(tt.args.i, tt.args.j)
		})
	}
}

func TestRoleByPos_Less(t *testing.T) {
	type args struct {
		i int
		j int
	}
	tests := []struct {
		name string
		a    RoleByPos
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.Less(tt.args.i, tt.args.j); got != tt.want {
				t.Errorf("RoleByPos.Less() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDiscord_AddEventHandler(t *testing.T) {
	type args struct {
		h interface{}
	}
	tests := []struct {
		name string
		d    *Discord
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.d.AddEventHandler(tt.args.h)
		})
	}
}

func TestDiscord_AddEventHandlerOnce(t *testing.T) {
	type args struct {
		h interface{}
	}
	tests := []struct {
		name string
		d    *Discord
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.d.AddEventHandlerOnce(tt.args.h)
		})
	}
}

func TestDiscord_Guilds(t *testing.T) {
	tests := []struct {
		name string
		d    *Discord
		want []*discordgo.Guild
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.d.Guilds(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Discord.Guilds() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDiscord_GuildCount(t *testing.T) {
	tests := []struct {
		name string
		d    *Discord
		want int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.d.GuildCount(); got != tt.want {
				t.Errorf("Discord.GuildCount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDiscord_Guild(t *testing.T) {
	type args struct {
		guildID string
	}
	tests := []struct {
		name    string
		d       *Discord
		args    args
		want    *discordgo.Guild
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.d.Guild(tt.args.guildID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Discord.Guild() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Discord.Guild() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDiscord_Channel(t *testing.T) {
	type args struct {
		channelID string
	}
	tests := []struct {
		name    string
		d       *Discord
		args    args
		want    *discordgo.Channel
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.d.Channel(tt.args.channelID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Discord.Channel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Discord.Channel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDiscord_Member(t *testing.T) {
	type args struct {
		guildID string
		userID  string
	}
	tests := []struct {
		name    string
		d       *Discord
		args    args
		want    *discordgo.Member
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.d.Member(tt.args.guildID, tt.args.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Discord.Member() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Discord.Member() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDiscord_Role(t *testing.T) {
	type args struct {
		guildID string
		roleID  string
	}
	tests := []struct {
		name    string
		d       *Discord
		args    args
		want    *discordgo.Role
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.d.Role(tt.args.guildID, tt.args.roleID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Discord.Role() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Discord.Role() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDiscord_GuildRoleByNameOrID(t *testing.T) {
	type args struct {
		guildID string
		name    string
		id      string
	}
	tests := []struct {
		name    string
		d       *Discord
		args    args
		want    *discordgo.Role
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.d.GuildRoleByNameOrID(tt.args.guildID, tt.args.name, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Discord.GuildRoleByNameOrID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Discord.GuildRoleByNameOrID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDiscord_IsBotOwner(t *testing.T) {
	type args struct {
		msg *DiscordMessage
	}
	tests := []struct {
		name string
		d    *Discord
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.d.IsBotOwner(tt.args.msg); got != tt.want {
				t.Errorf("Discord.IsBotOwner() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDiscord_StartTyping(t *testing.T) {
	type args struct {
		channelID string
	}
	tests := []struct {
		name    string
		d       *Discord
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.d.StartTyping(tt.args.channelID); (err != nil) != tt.wantErr {
				t.Errorf("Discord.StartTyping() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDiscord_SendMessage(t *testing.T) {
	type args struct {
		channelID string
		content   string
	}
	tests := []struct {
		name    string
		d       *Discord
		args    args
		want    *discordgo.Message
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.d.SendMessage(tt.args.channelID, tt.args.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("Discord.SendMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Discord.SendMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDiscord_UpdateStatus(t *testing.T) {
	type args struct {
		status       string
		activityType discordgo.ActivityType
	}
	tests := []struct {
		name string
		d    *Discord
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.d.UpdateStatus(tt.args.status, tt.args.activityType)
		})
	}
}

func TestDiscord_IsOwner(t *testing.T) {
	type args struct {
		userID string
	}
	tests := []struct {
		name string
		d    *Discord
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.d.IsOwner(tt.args.userID); got != tt.want {
				t.Errorf("Discord.IsOwner() = %v, want %v", got, tt.want)
			}
		})
	}
}
