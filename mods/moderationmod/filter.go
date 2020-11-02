package moderationmod

import "github.com/intrntsrfr/meidov2"

func (m *ModerationMod) FilterWord(msg *meidov2.DiscordMessage) {
	if msg.LenArgs() < 2 || msg.Args()[0] != "m?fw" {
		return
	}

	/*	phrase := msg.Args()[1:]

		fe := &FilterEntry{}

	*/
}
