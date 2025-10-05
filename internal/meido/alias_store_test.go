package meido

import (
	"testing"

	"github.com/intrntsrfr/meido/internal/structs"
	"github.com/intrntsrfr/meido/pkg/mio/bot"
)

func TestConvertAliasStructs(t *testing.T) {
	input := []*structs.CommandAlias{
		{GuildID: "guild", Alias: "!ping", Command: "ping"},
		nil,
		{GuildID: "guild", Alias: "m?stats", Command: "stats"},
	}
	got := convertAliasStructs(input)
	want := []bot.CommandAlias{
		{GuildID: "guild", Alias: "!ping", Command: "ping"},
		{GuildID: "guild", Alias: "m?stats", Command: "stats"},
	}
	if len(got) != len(want) {
		t.Fatalf("convertAliasStructs length = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("convertAliasStructs[%d] = %#v, want %#v", i, got[i], want[i])
		}
	}
}

func TestConvertAliasStructs_Empty(t *testing.T) {
	if got := convertAliasStructs(nil); got != nil {
		t.Fatalf("convertAliasStructs(nil) = %#v, want nil", got)
	}
}
