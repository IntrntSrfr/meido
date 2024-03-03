package builders

import (
	"testing"

	"github.com/intrntsrfr/meido/pkg/utils"
)

func TestNewEmbed(t *testing.T) {
	embed := NewEmbedBuilder()
	if embed == nil {
		t.Error("NewEmbed returned nil")
	}
}

func TestEmbed_AddField(t *testing.T) {
	embed := NewEmbedBuilder()
	embed.AddField("TestName", "TestValue", true)

	if len(embed.Fields) != 1 {
		t.Errorf("Expected 1 field, got %d", len(embed.Fields))
	}

	field := embed.Fields[0]
	if field.Name != "TestName" || field.Value != "TestValue" || field.Inline != true {
		t.Error("AddField did not set the fields correctly")
	}
}

func TestEmbed_WithThumbnail(t *testing.T) {
	embed := NewEmbedBuilder()
	url := "http://example.com/thumbnail.jpg"
	embed.WithThumbnail(url)

	if embed.Thumbnail == nil || embed.Thumbnail.URL != url {
		t.Error("WithThumbnail did not set the thumbnail URL correctly")
	}
}

func TestEmbed_WithFooter(t *testing.T) {
	embed := NewEmbedBuilder()
	text, iconURL := "Test Footer", "http://example.com/icon.jpg"
	embed.WithFooter(text, iconURL)

	if embed.Footer == nil || embed.Footer.Text != text || embed.Footer.IconURL != iconURL {
		t.Error("WithFooter did not set the footer correctly")
	}
}

func TestEmbed_WithAuthor(t *testing.T) {
	embed := NewEmbedBuilder()
	name, url := "Author Name", "http://example.com"
	embed.WithAuthor(name, url)

	if embed.Author == nil || embed.Author.Name != name || embed.Author.URL != url {
		t.Error("WithAuthor did not set the author correctly")
	}
}

func TestEmbed_WithUrl(t *testing.T) {
	embed := NewEmbedBuilder()
	url := "http://example.com"
	embed.WithUrl(url)

	if embed.URL != url {
		t.Error("WithUrl did not set the URL correctly")
	}
}

func TestEmbed_WithImageUrl(t *testing.T) {
	embed := NewEmbedBuilder()
	url := "http://example.com/image.jpg"
	embed.WithImageUrl(url)

	if embed.Image == nil || embed.Image.URL != url {
		t.Error("WithImageUrl did not set the image URL correctly")
	}
}

func TestEmbed_WithTitle(t *testing.T) {
	embed := NewEmbedBuilder()
	title := "Test Title"
	embed.WithTitle(title)

	if embed.Title != title {
		t.Error("WithTitle did not set the title correctly")
	}
}

func TestEmbed_WithDescription(t *testing.T) {
	embed := NewEmbedBuilder()
	description := "Test Description"
	embed.WithDescription(description)

	if embed.Description != description {
		t.Error("WithDescription did not set the description correctly")
	}
}

func TestEmbed_WithTimestamp(t *testing.T) {
	embed := NewEmbedBuilder()
	timestamp := "2024-01-01T12:00:00Z"
	embed.WithTimestamp(timestamp)

	if embed.Timestamp != timestamp {
		t.Error("WithTimestamp did not set the timestamp correctly")
	}
}

func TestEmbed_WithOkColor(t *testing.T) {
	embed := NewEmbedBuilder()
	embed.WithOkColor()

	// Assuming utils.ColorInfo is a predefined constant
	if embed.Color != utils.ColorInfo {
		t.Error("WithOkColor did not set the color correctly")
	}
}

func TestEmbed_WithErrorColor(t *testing.T) {
	embed := NewEmbedBuilder()
	embed.WithErrorColor()

	// Assuming utils.ColorCritical is a predefined constant
	if embed.Color != utils.ColorCritical {
		t.Error("WithErrorColor did not set the color correctly")
	}
}

func TestEmbed_WithColor(t *testing.T) {
	embed := NewEmbedBuilder()
	color := 123456
	embed.WithColor(color)

	if embed.Color != color {
		t.Errorf("WithColor did not set the color correctly, expected %d, got %d", color, embed.Color)
	}
}

func TestEmbed_Build(t *testing.T) {
	embed := NewEmbedBuilder()
	if embed.Build() != embed.MessageEmbed {
		t.Error("Build does not return the correct MessageEmbed instance")
	}
}
