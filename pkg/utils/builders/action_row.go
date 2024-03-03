package builders

import "github.com/bwmarrin/discordgo"

type ActionRowBuilder struct {
	actionRow *discordgo.ActionsRow
}

func NewActionRowBuilder() *ActionRowBuilder {
	return &ActionRowBuilder{actionRow: &discordgo.ActionsRow{}}
}

func (b *ActionRowBuilder) AddComponent(component discordgo.MessageComponent) *ActionRowBuilder {
	b.actionRow.Components = append(b.actionRow.Components, component)
	return b
}

func (b *ActionRowBuilder) AddButton(label string, style discordgo.ButtonStyle, customID string) *ActionRowBuilder {
	b.actionRow.Components = append(b.actionRow.Components, &discordgo.Button{
		Label:    label,
		Style:    style,
		CustomID: customID,
	})
	return b
}

func (b *ActionRowBuilder) Build() *discordgo.ActionsRow {
	return b.actionRow
}
