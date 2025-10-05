package meido

import (
	"github.com/intrntsrfr/meido/internal/database"
	"github.com/intrntsrfr/meido/internal/structs"
	"github.com/intrntsrfr/meido/pkg/mio/bot"
)

type dbAliasStore struct {
	db database.DB
}

func newDBAliasStore(db database.DB) bot.CommandAliasStore {
	if db == nil {
		return nil
	}
	return &dbAliasStore{db: db}
}

func (s *dbAliasStore) GuildCommandAliases(guildID string) ([]bot.CommandAlias, error) {
	aliases, err := s.db.GetCommandAliases(guildID)
	if err != nil {
		return nil, err
	}
	return convertAliasStructs(aliases), nil
}

func convertAliasStructs(input []*structs.CommandAlias) []bot.CommandAlias {
	if len(input) == 0 {
		return nil
	}
	out := make([]bot.CommandAlias, 0, len(input))
	for _, alias := range input {
		if alias == nil {
			continue
		}
		out = append(out, bot.CommandAlias{
			GuildID: alias.GuildID,
			Alias:   alias.Alias,
			Command: alias.Command,
		})
	}
	return out
}
