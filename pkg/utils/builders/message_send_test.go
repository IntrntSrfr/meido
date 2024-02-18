package builders

import (
	"reflect"
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestNewMessageSendBuilder(t *testing.T) {
	builder := NewMessageSendBuilder()
	if builder == nil {
		t.Errorf("NewMessageSendBuilder() returned nil")
		return // to get rid of staticcheck warning
	}
	if builder.message == nil {
		t.Errorf("NewMessageSendBuilder().message is nil")
	}
}

func TestMessageSendBuilder_Content(t *testing.T) {
	builder := NewMessageSendBuilder()
	testContent := "Hello, World!"
	builder.Content(testContent)

	if builder.message.Content != testContent {
		t.Errorf("Content was not set correctly, got: %s, want: %s", builder.message.Content, testContent)
	}
}

func TestMessageSendBuilder_Embed(t *testing.T) {
	builder := NewMessageSendBuilder()
	testEmbed := &discordgo.MessageEmbed{Title: "Test Embed"}
	builder.Embed(testEmbed)

	if !reflect.DeepEqual(builder.message.Embed, testEmbed) {
		t.Errorf("Embed was not added correctly")
	}
}

func TestMessageSendBuilder_AddActionRow(t *testing.T) {
	builder := NewMessageSendBuilder()
	actionRow := &discordgo.ActionsRow{}
	builder.AddActionRow(actionRow)

	if len(builder.message.Components) != 1 {
		t.Errorf("Expected 1 action row, got %d", len(builder.message.Components))
	}

	if !reflect.DeepEqual(builder.message.Components[0], actionRow) {
		t.Errorf("Action row was not added correctly")
	}
}

func TestMessageSendBuilder_Build(t *testing.T) {
	builder := NewMessageSendBuilder()
	testContent := "Hello, World!"
	builder.Content(testContent)

	messageSend := builder.Build()
	if !reflect.DeepEqual(messageSend, builder.message) {
		t.Errorf("Build() did not return the correct MessageSend")
	}
}
