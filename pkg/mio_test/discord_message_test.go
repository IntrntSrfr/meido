package mio_test

import (
	"reflect"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/pkg/mio"
)

func TestDiscordMessage_Type(t *testing.T) {
	msg := &mio.DiscordMessage{MessageType: mio.MessageTypeCreate}
	if got := msg.Type(); got != mio.MessageTypeCreate {
		t.Errorf("DiscordMessage.Type() = %v, want %v", got, mio.MessageTypeCreate)
	}
}

func TestDiscordMessage_Args(t *testing.T) {
	msg := &mio.DiscordMessage{Message: &discordgo.Message{Content: "I am a message"}}
	expected := []string{"i", "am", "a", "message"}
	if got := msg.Args(); !reflect.DeepEqual(got, expected) {
		t.Errorf("DiscordMessage.Args() = %v, want %v", got, expected)
	}
}

func TestDiscordMessage_RawArgs(t *testing.T) {
	msg := &mio.DiscordMessage{Message: &discordgo.Message{Content: "I am a message"}}
	expected := []string{"I", "am", "a", "message"}
	if got := msg.RawArgs(); !reflect.DeepEqual(got, expected) {
		t.Errorf("DiscordMessage.RawArgs() = %v, want %v", got, expected)
	}
}

func TestDiscordMessage_RawContent(t *testing.T) {
	msg := &mio.DiscordMessage{Message: &discordgo.Message{Content: "I am a message"}}
	expected := "I am a message"
	if got := msg.RawContent(); got != expected {
		t.Errorf("DiscordMessage.RawContent() = %v, want %v", got, expected)
	}
}

func TestDiscordMessage_IsDM(t *testing.T) {
	msg := &mio.DiscordMessage{Message: &discordgo.Message{GuildID: ""}}
	if ok := msg.IsDM(); !ok {
		t.Errorf("DiscordMessage.IsDM() = %v, want %v", ok, true)
	}

	msg = &mio.DiscordMessage{Message: &discordgo.Message{GuildID: "12341234"}}
	if ok := msg.IsDM(); ok {
		t.Errorf("DiscordMessage.IsDM() = %v, want %v", ok, false)
	}
}

func TestDiscordMessage_IsBot(t *testing.T) {
	msg := &mio.DiscordMessage{Message: &discordgo.Message{Author: &discordgo.User{Bot: false}}}
	if ok := msg.IsBot(); ok {
		t.Errorf("DiscordMessage.IsBot() = %v, want %v", ok, false)
	}
	msg = &mio.DiscordMessage{Message: &discordgo.Message{Author: &discordgo.User{Bot: true}}}
	if ok := msg.IsBot(); !ok {
		t.Errorf("DiscordMessage.IsBot() = %v, want %v", ok, true)
	}
}

func TestDiscordMessage_Author(t *testing.T) {
	author := &discordgo.User{}
	msg := &mio.DiscordMessage{Message: &discordgo.Message{Author: author}}
	if got := msg.Author(); got != author {
		t.Errorf("DiscordMessage.Author() = %v, want %v", got, author)
	}
}

func TestDiscordMessage_Member(t *testing.T) {
	member := &discordgo.Member{}
	msg := &mio.DiscordMessage{Message: &discordgo.Message{Member: member}}
	if got := msg.Member(); got != member {
		t.Errorf("DiscordMessage.Member() = %v, want %v", got, member)
	}
}

func TestDiscordMessage_ID(t *testing.T) {
	id := "1234"
	msg := &mio.DiscordMessage{Message: &discordgo.Message{ID: id}}
	if got := msg.ID(); got != id {
		t.Errorf("DiscordMessage.ID() = %v, want %v", got, id)
	}
}

func TestDiscordMessage_AuthorID(t *testing.T) {
	id := "1234"
	msg := &mio.DiscordMessage{Message: &discordgo.Message{Author: &discordgo.User{ID: id}}}
	if got := msg.AuthorID(); got != id {
		t.Errorf("DiscordMessage.AuthorID() = %v, want %v", got, id)
	}

	id = ""
	msg = &mio.DiscordMessage{Message: &discordgo.Message{}}
	if got := msg.AuthorID(); got != id {
		t.Errorf("DiscordMessage.AuthorID() = %v, want %v", got, id)
	}
}

func TestDiscordMessage_GuildID(t *testing.T) {
	id := "1234"
	msg := &mio.DiscordMessage{Message: &discordgo.Message{GuildID: id}}
	if got := msg.GuildID(); got != id {
		t.Errorf("DiscordMessage.GuildID() = %v, want %v", got, id)
	}
}

func TestDiscordMessage_ChannelID(t *testing.T) {
	id := "1234"
	msg := &mio.DiscordMessage{Message: &discordgo.Message{ChannelID: id}}
	if got := msg.ChannelID(); got != id {
		t.Errorf("DiscordMessage.ChannelID() = %v, want %v", got, id)
	}
}

func TestDiscordMessage_Mentions(t *testing.T) {
	mentions := []*discordgo.User{{ID: "123"}, {ID: "456"}}
	msg := &mio.DiscordMessage{Message: &discordgo.Message{Mentions: mentions}}
	if got := msg.Mentions(); !reflect.DeepEqual(got, mentions) {
		t.Errorf("DiscordMessage.Mentions() = %v, want %v", got, mentions)
	}
}

func TestDiscordMessage_MentionRoles(t *testing.T) {
	mentions := []string{"123", "456"}
	msg := &mio.DiscordMessage{Message: &discordgo.Message{MentionRoles: mentions}}
	if got := msg.MentionRoles(); !reflect.DeepEqual(got, mentions) {
		t.Errorf("DiscordMessage.MentionRoles() = %v, want %v", got, mentions)
	}
}

func TestDiscordMessage_Attachments(t *testing.T) {
	attachments := []*discordgo.MessageAttachment{{ID: "123"}, {ID: "456"}}
	msg := &mio.DiscordMessage{Message: &discordgo.Message{Attachments: attachments}}
	if got := msg.Attachments(); !reflect.DeepEqual(got, attachments) {
		t.Errorf("DiscordMessage.Attachments() = %v, want %v", got, attachments)
	}
}

func TestDiscordMessage_CallbackKey(t *testing.T) {
	msg := &mio.DiscordMessage{Message: &discordgo.Message{ChannelID: "1", Author: &discordgo.User{ID: "2"}}}
	if got := msg.CallbackKey(); got != "1:2" {
		t.Errorf("DiscordMessage.CallbackKey() = %v, want %v", got, "1:2")
	}
}
