package builders

import (
	"reflect"
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestNewActionRowBuilder(t *testing.T) {
	builder := NewActionRowBuilder()
	if builder == nil {
		t.Errorf("NewActionRowBuilder() returned nil")
		return // to get rid of staticcheck warning
	}
	if builder.actionRow == nil {
		t.Errorf("NewActionRowBuilder().actionRow is nil")
	}
}

func TestActionRowBuilder_AddComponent(t *testing.T) {
	builder := NewActionRowBuilder()
	component := &discordgo.Button{Label: "Test Button", Style: discordgo.PrimaryButton, CustomID: "test_button"}
	builder.AddComponent(component)

	if len(builder.actionRow.Components) != 1 {
		t.Errorf("Expected 1 component, got %d", len(builder.actionRow.Components))
	}

	if !reflect.DeepEqual(builder.actionRow.Components[0], component) {
		t.Errorf("Component was not added correctly")
	}
}

func TestActionRowBuilder_AddButton(t *testing.T) {
	builder := NewActionRowBuilder()
	label := "Test Button"
	style := discordgo.PrimaryButton
	customID := "test_button"

	builder.AddButton(label, style, customID)
	if len(builder.actionRow.Components) != 1 {
		t.Errorf("Expected 1 button, got %d", len(builder.actionRow.Components))
	}

	button, ok := builder.actionRow.Components[0].(*discordgo.Button)
	if !ok {
		t.Fatalf("Expected component to be of type *discordgo.Button")
	}

	if button.Label != label || button.Style != style || button.CustomID != customID {
		t.Errorf("Button properties do not match")
	}
}

func TestActionRowBuilder_Build(t *testing.T) {
	builder := NewActionRowBuilder()
	component := &discordgo.Button{Label: "Test Button", Style: discordgo.PrimaryButton, CustomID: "test_button"}
	builder.AddComponent(component)

	actionRow := builder.Build()
	if !reflect.DeepEqual(actionRow, builder.actionRow) {
		t.Errorf("Build() did not return the correct ActionsRow")
	}
}
