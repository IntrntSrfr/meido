package bot

import (
	"reflect"
	"sync"
	"testing"
)

type fakeAliasStore struct {
	lock  sync.Mutex
	data  map[string][]CommandAlias
	calls map[string]int
}

func newFakeAliasStore(data map[string][]CommandAlias) *fakeAliasStore {
	return &fakeAliasStore{data: data, calls: make(map[string]int)}
}

func (f *fakeAliasStore) GuildCommandAliases(guildID string) ([]CommandAlias, error) {
	f.lock.Lock()
	f.calls[guildID]++
	f.lock.Unlock()
	aliases := f.data[guildID]
	if aliases == nil {
		return nil, nil
	}
	out := make([]CommandAlias, len(aliases))
	copy(out, aliases)
	return out, nil
}

func (f *fakeAliasStore) callCount(guildID string) int {
	f.lock.Lock()
	defer f.lock.Unlock()
	return f.calls[guildID]
}

func TestCommandAliasManagerResolve(t *testing.T) {
	store := newFakeAliasStore(map[string][]CommandAlias{
		"guild": {
			{GuildID: "guild", Alias: "!ping", Command: "ping"},
			{GuildID: "guild", Alias: "m?Stats", Command: "stats"},
		},
	})
	manager := newCommandAliasManager(store)

	tests := []struct {
		name   string
		guild  string
		tokens []string
		want   string
		wantOK bool
	}{
		{
			name:   "single token alias",
			guild:  "guild",
			tokens: []string{"!ping"},
			want:   "ping",
			wantOK: true,
		},
		{
			name:   "case insensitive match",
			guild:  "guild",
			tokens: []string{"M?stats"},
			want:   "stats",
			wantOK: true,
		},
		{
			name:   "missing alias",
			guild:  "guild",
			tokens: []string{"!unknown"},
			want:   "",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := manager.Resolve(tt.guild, tt.tokens)
			if ok != tt.wantOK {
				t.Fatalf("Resolve() ok = %v, want %v", ok, tt.wantOK)
			}
			if got != tt.want {
				t.Fatalf("Resolve() = %q, want %q", got, tt.want)
			}
		})
	}

	if calls := store.callCount("guild"); calls != 1 {
		t.Fatalf("expected aliases to be loaded once, got %d", calls)
	}
}

func TestCommandAliasManagerResolveMultiWord(t *testing.T) {
	store := newFakeAliasStore(map[string][]CommandAlias{
		"guild": {
			{GuildID: "guild", Alias: "m?mod settings", Command: "moderationsettings"},
		},
	})
	manager := newCommandAliasManager(store)

	if got, ok := manager.Resolve("guild", []string{"m?mod", "settings", "warns"}); !ok || got != "moderationsettings" {
		t.Fatalf("Resolve() = %q, ok=%v; want moderationsettings,true", got, ok)
	}
}

func TestCommandAliasManagerInvalidate(t *testing.T) {
	store := newFakeAliasStore(map[string][]CommandAlias{
		"guild": {
			{GuildID: "guild", Alias: "!ping", Command: "ping"},
		},
	})
	manager := newCommandAliasManager(store)

	if _, ok := manager.Resolve("guild", []string{"!ping"}); !ok {
		t.Fatalf("initial resolve failed")
	}

	store.lock.Lock()
	store.data["guild"] = []CommandAlias{
		{GuildID: "guild", Alias: "!pong", Command: "ping"},
	}
	store.lock.Unlock()

	manager.Invalidate("guild")

	if cmd, ok := manager.Resolve("guild", []string{"!ping"}); ok || cmd != "" {
		t.Fatalf("expected !ping alias to be gone, got %q", cmd)
	}
	if cmd, ok := manager.Resolve("guild", []string{"!pong"}); !ok || cmd != "ping" {
		t.Fatalf("expected new alias after invalidation, got %q ok=%v", cmd, ok)
	}

	if calls := store.callCount("guild"); calls != 2 {
		t.Fatalf("expected store to be called twice after invalidation, got %d", calls)
	}
}

func TestConvertAliasesIgnoresInvalid(t *testing.T) {
	in := []CommandAlias{
		{GuildID: "guild", Alias: "", Command: "ping"},
		{GuildID: "guild", Alias: "!pong", Command: ""},
		{GuildID: "guild", Alias: "!ping", Command: "ping"},
	}
	got := convertAliases(in)
	if len(got) != 1 {
		t.Fatalf("convertAliases() length = %d, want 1", len(got))
	}
	if !reflect.DeepEqual(got[0].aliasTokens, []string{"!ping"}) {
		t.Fatalf("unexpected alias tokens: %v", got[0].aliasTokens)
	}
}
