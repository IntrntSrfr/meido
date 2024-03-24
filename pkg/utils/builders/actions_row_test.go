package builders

import (
	"reflect"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
)

func TestNewActionRowBuilder(t *testing.T) {
	builder := NewActionRowBuilder()
	assert.NotNil(t, builder, "NewActionRowBuilder() returned nil")
	assert.NotNil(t, builder.actionRow, "NewActionRowBuilder().actionRow is nil")
}

func TestActionRowBuilder_AddComponent(t *testing.T) {
	builder := NewActionRowBuilder()
	component := &discordgo.Button{Label: "Test Button", Style: discordgo.PrimaryButton, CustomID: "test_button"}
	builder.AddComponent(component)

	assert.Len(t, builder.actionRow.Components, 1, "Expected 1 component")
	assert.True(t, reflect.DeepEqual(builder.actionRow.Components[0], component), "Component was not added correctly")
}

func TestActionRowBuilder_AddButton(t *testing.T) {
	builder := NewActionRowBuilder()
	label := "Test Button"
	style := discordgo.PrimaryButton
	customID := "test_button"

	builder.AddButton(label, style, customID)

	assert.Len(t, builder.actionRow.Components, 1, "Expected 1 button")

	button, ok := builder.actionRow.Components[0].(*discordgo.Button)
	assert.True(t, ok, "Expected component to be of type *discordgo.Button")

	assert.Equal(t, label, button.Label, "Button label does not match")
	assert.Equal(t, style, button.Style, "Button style does not match")
	assert.Equal(t, customID, button.CustomID, "Button custom ID does not match")
}

func TestActionRowBuilder_AddSelectMenu(t *testing.T) {
	builder := NewActionRowBuilder()
	placeholder := "Test Select Menu"
	options := []discordgo.SelectMenuOption{
		{Label: "Option 1", Value: "option_1"},
		{Label: "Option 2", Value: "option_2"},
	}
	customID := "test_select_menu"

	builder.AddSelectMenu(placeholder, options, customID)

	assert.Len(t, builder.actionRow.Components, 1, "Expected 1 select menu")

	selectMenu, ok := builder.actionRow.Components[0].(*discordgo.SelectMenu)
	assert.True(t, ok, "Expected component to be of type *discordgo.SelectMenu")

	assert.Equal(t, placeholder, selectMenu.Placeholder, "Select menu placeholder does not match")
	assert.Equal(t, options, selectMenu.Options, "Select menu options do not match")
	assert.Equal(t, customID, selectMenu.CustomID, "Select menu custom ID does not match")
}

func TestActionRowBuilder_Build(t *testing.T) {
	builder := NewActionRowBuilder()
	component := &discordgo.Button{Label: "Test Button", Style: discordgo.PrimaryButton, CustomID: "test_button"}
	builder.AddComponent(component)

	actionRow := builder.Build()
	assert.True(t, reflect.DeepEqual(actionRow, builder.actionRow), "Build() did not return the correct ActionsRow")
}
