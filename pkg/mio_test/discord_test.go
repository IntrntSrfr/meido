package mio_test

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/pkg/mio"
	"github.com/intrntsrfr/meido/pkg/mio/mocks"
	"github.com/intrntsrfr/meido/pkg/mio/test"
	"github.com/intrntsrfr/meido/pkg/utils"
)

func TestSessionWrapper_ShardID(t *testing.T) {
	shardID := 2
	s := &mio.SessionWrapper{&discordgo.Session{ShardID: shardID}}

	if got := s.ShardID(); got != shardID {
		t.Errorf("SessionWrapper.ShardID() = %v, want %v", got, shardID)
	}
}

func TestSessionWrapper_State(t *testing.T) {
	state := discordgo.NewState()
	s := &mio.SessionWrapper{&discordgo.Session{State: state}}
	s.State()

	if got := s.State(); !reflect.DeepEqual(got, state) {
		t.Errorf("SessionWrapper.State() = %v, want %v", got, state)
	}
}

func TestNewDiscord(t *testing.T) {
	token := "Bot asdf"
	shards := 1
	conf := utils.NewConfig()
	conf.Set("token", token)
	conf.Set("shards", shards)
	mockSess := mocks.NewDiscordSession(token, shards)

	logger := mio.NewDiscardLogger()
	d := test.NewTestDiscord(conf, mockSess, nil)

	if got := d.token; d.token != token {
		t.Errorf("SessionWrapper.token = %v, want %v", got, token)
	}
	if got := d.shards; d.shards != shards {
		t.Errorf("SessionWrapper.shards = %v, want %v", got, shards)
	}
	if got := d.logger; d.logger == nil {
		t.Errorf("SessionWrapper.shards = %v, want %v", got, logger)
	}

	if got := len(d.Sessions); got != shards {
		t.Errorf("len(Discord.Sessions) = %v, want %v", got, 1)
	}
}

func TestDiscord_Run(t *testing.T) {
	d := NewTestDiscord(nil, nil, nil)
	if got := d.Run(); got != nil {
		t.Errorf("Discord.Run() error = %v, wantErr %v", got, false)
	}

	// should fail second time
	if got := d.Run(); got == nil {
		t.Errorf("Discord.Run() error = %v, wantErr %v", got, true)
	}
}

func TestDiscord_Close(t *testing.T) {
	sess := mocks.NewDiscordSession("123", 1)
	sess.CloseShouldFail = true

	logBuf := bytes.Buffer{}
	d := NewTestDiscord(nil, sess, mio.NewLogger(&logBuf))
	d.Close()

	expected := "Failed to close session"
	if got := logBuf.String(); !strings.Contains(got, expected) {
		t.Errorf("Discord.Close() log output = %v, want %v", got, expected)
	}
}

func expectMessage(ch chan *mio.DiscordMessage) error {
	select {
	case <-ch:
	case <-time.After(time.Millisecond * 25):
		return fmt.Errorf("message was not received; timed out")
	}
	return nil
}

func expectNoMessage(ch chan *mio.DiscordMessage) error {
	select {
	case <-ch:
		return fmt.Errorf("message was received; expected none")
	case <-time.After(time.Millisecond * 25):
	}
	return nil
}

func TestDiscord_onMessageCreate(t *testing.T) {
	d := mio.NewDiscord("asdf", 1, mio.NewLogger(io.Discard))

	t.Run("empty message does not go through", func(t *testing.T) {
		d.onMessageCreate(&discordgo.Session{}, &discordgo.MessageCreate{})
		if err := expectNoMessage(d.Messages()); err != nil {
			t.Errorf("%v", err.Error())
		}
	})

	t.Run("DM goes through", func(t *testing.T) {
		d.onMessageCreate(&discordgo.Session{}, &discordgo.MessageCreate{
			Message: &discordgo.Message{
				Author: &discordgo.User{},
			},
		})
		if err := expectMessage(d.Messages()); err != nil {
			t.Errorf("%v", err.Error())
		}
	})

	t.Run("message with guild, but no member does not go through", func(t *testing.T) {
		d.onMessageCreate(&discordgo.Session{}, &discordgo.MessageCreate{
			Message: &discordgo.Message{
				GuildID: "1234",
				Author:  &discordgo.User{},
			},
		})
		if err := expectNoMessage(d.Messages()); err != nil {
			t.Errorf("%v", err.Error())
		}
	})

	t.Run("message with guild and member goes through", func(t *testing.T) {
		d.onMessageCreate(&discordgo.Session{}, &discordgo.MessageCreate{
			Message: &discordgo.Message{
				GuildID: "1234",
				Author:  &discordgo.User{},
				Member:  &discordgo.Member{},
			},
		})
		if err := expectMessage(d.Messages()); err != nil {
			t.Errorf("%v", err.Error())
		}
	})
}

func TestDiscord_onMessageUpdate(t *testing.T) {
	d := mio.NewDiscord("asdf", 1, mio.NewDefaultLogger())

	t.Run("empty message does not go through", func(t *testing.T) {
		d.onMessageUpdate(&discordgo.Session{}, &discordgo.MessageUpdate{})
		if err := expectNoMessage(d.Messages()); err != nil {
			t.Errorf("%v", err.Error())
		}
	})

	t.Run("DM goes through", func(t *testing.T) {
		d.onMessageUpdate(&discordgo.Session{}, &discordgo.MessageUpdate{
			Message: &discordgo.Message{
				Author: &discordgo.User{},
			},
		})
		if err := expectMessage(d.Messages()); err != nil {
			t.Errorf("%v", err.Error())
		}
	})

	t.Run("message with guild, but no member does not go through", func(t *testing.T) {
		d.onMessageUpdate(&discordgo.Session{}, &discordgo.MessageUpdate{
			Message: &discordgo.Message{
				GuildID: "1234",
				Author:  &discordgo.User{},
			},
		})
		if err := expectNoMessage(d.Messages()); err != nil {
			t.Errorf("%v", err.Error())
		}
	})

	t.Run("message with guild and member goes through", func(t *testing.T) {
		d.onMessageUpdate(&discordgo.Session{}, &discordgo.MessageUpdate{
			Message: &discordgo.Message{
				GuildID: "1234",
				Author:  &discordgo.User{},
				Member:  &discordgo.Member{},
			},
		})
		if err := expectMessage(d.Messages()); err != nil {
			t.Errorf("%v", err.Error())
		}
	})
}

func TestDiscord_onMessageDelete(t *testing.T) {
	d := mio.NewDiscord("asdf", 1, mio.NewLogger(io.Discard))

	t.Run("empty message goes through", func(t *testing.T) {
		d.onMessageDelete(&discordgo.Session{}, &discordgo.MessageDelete{})
		if err := expectMessage(d.Messages()); err != nil {
			t.Errorf("%v", err.Error())
		}
	})
}

func TestDiscord_BotUser(t *testing.T) {
	tests := []struct {
		name string
		d    *mio.Discord
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
		d       *mio.Discord
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
		d       *mio.Discord
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
		d       *mio.Discord
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
		d    *mio.Discord
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
		d    *mio.Discord
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
		d    *mio.Discord
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
		a    mio.RoleByPos
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
		a    mio.RoleByPos
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
		a    mio.RoleByPos
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

func NewTestDiscord_AddEventHandler(t *testing.T) {
	type args struct {
		h interface{}
	}
	tests := []struct {
		name string
		d    *mio.Discord
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

func NewTestDiscord_AddEventHandlerOnce(t *testing.T) {
	type args struct {
		h interface{}
	}
	tests := []struct {
		name string
		d    *mio.Discord
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

func NewTestDiscord_Guilds(t *testing.T) {
	tests := []struct {
		name string
		d    *mio.Discord
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

func NewTestDiscord_GuildCount(t *testing.T) {
	tests := []struct {
		name string
		d    *mio.Discord
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

func NewTestDiscord_Guild(t *testing.T) {
	type args struct {
		guildID string
	}
	tests := []struct {
		name    string
		d       *mio.Discord
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

func NewTestDiscord_Channel(t *testing.T) {
	type args struct {
		channelID string
	}
	tests := []struct {
		name    string
		d       *mio.Discord
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

func NewTestDiscord_Member(t *testing.T) {
	type args struct {
		guildID string
		userID  string
	}
	tests := []struct {
		name    string
		d       *mio.Discord
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

func NewTestDiscord_Role(t *testing.T) {
	type args struct {
		guildID string
		roleID  string
	}
	tests := []struct {
		name    string
		d       *mio.Discord
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

func NewTestDiscord_GuildRoleByNameOrID(t *testing.T) {
	type args struct {
		guildID string
		name    string
		id      string
	}
	tests := []struct {
		name    string
		d       *mio.Discord
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

func NewTestDiscord_StartTyping(t *testing.T) {
	type args struct {
		channelID string
	}
	tests := []struct {
		name    string
		d       *mio.Discord
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

func NewTestDiscord_SendMessage(t *testing.T) {
	type args struct {
		channelID string
		content   string
	}
	tests := []struct {
		name    string
		d       *mio.Discord
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

func NewTestDiscord_UpdateStatus(t *testing.T) {
	type args struct {
		status       string
		activityType discordgo.ActivityType
	}
	tests := []struct {
		name string
		d    *mio.Discord
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
