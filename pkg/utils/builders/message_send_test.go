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

func TestMessageSendBuilder_TTS(t *testing.T) {
	builder := NewMessageSendBuilder()
	builder.WithTTS(true)

	if !builder.message.TTS {
		t.Errorf("TTS was not set correctly")
	}
}

func TestMessageSendBuilder_File(t *testing.T) {
	builder := NewMessageSendBuilder()
	testFile := &discordgo.File{Name: "test.txt"}
	builder.WithFile(testFile)

	if !reflect.DeepEqual(builder.message.File, testFile) {
		t.Errorf("File was not set correctly")
	}
}

func TestMessageSendBuilder_Files(t *testing.T) {
	builder := NewMessageSendBuilder()
	testFiles := []*discordgo.File{
		{Name: "test1.txt"},
		{Name: "test2.txt"},
	}
	builder.WithFiles(testFiles)

	if !reflect.DeepEqual(builder.message.Files, testFiles) {
		t.Errorf("Files were not set correctly")
	}
}

func TestMessageSendBuilder_AddTextFile(t *testing.T) {
	builder := NewMessageSendBuilder()
	testName := "test.txt"
	testContent := "Hello, World!"
	builder.AddTextFile(testName, testContent)

	if len(builder.message.Files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(builder.message.Files))
	}

	if builder.message.Files[0].Name != testName {
		t.Errorf("File name was not set correctly")
	}

	if builder.message.Files[0].Reader == nil {
		t.Errorf("File reader was not set correctly")
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
