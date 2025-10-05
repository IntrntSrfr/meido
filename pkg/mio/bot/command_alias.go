package bot

import (
	"strings"
	"sync"
)

// CommandAliasStore defines retrieval for per-guild command aliases.
// Implementations can back this with any storage engine.
type CommandAliasStore interface {
	GuildCommandAliases(guildID string) ([]CommandAlias, error)
}

// CommandAlias represents an alias mapping for a guild.
type CommandAlias struct {
	GuildID string
	Alias   string
	Command string
}

type aliasInfo struct {
	aliasTokens []string
	commandName string
}

type aliasCacheEntry struct {
	aliases []*aliasInfo
}

type commandAliasManager struct {
	store CommandAliasStore
	lock  sync.RWMutex
	cache map[string]*aliasCacheEntry
}

func newCommandAliasManager(store CommandAliasStore) *commandAliasManager {
	if store == nil {
		return nil
	}
	return &commandAliasManager{
		store: store,
		cache: make(map[string]*aliasCacheEntry),
	}
}

func (m *commandAliasManager) Resolve(guildID string, tokens []string) (string, bool) {
	if m == nil || guildID == "" || len(tokens) == 0 {
		return "", false
	}
	entry, err := m.getOrLoadGuildAliases(guildID)
	if err != nil || entry == nil || len(entry.aliases) == 0 {
		return "", false
	}
	lowered := make([]string, len(tokens))
	for i, tok := range tokens {
		lowered[i] = strings.ToLower(tok)
	}
	for _, alias := range entry.aliases {
		if len(lowered) < len(alias.aliasTokens) {
			continue
		}
		if equalTokens(lowered[:len(alias.aliasTokens)], alias.aliasTokens) {
			return alias.commandName, true
		}
	}
	return "", false
}

func (m *commandAliasManager) Invalidate(guildID string) {
	if m == nil || guildID == "" {
		return
	}
	m.lock.Lock()
	defer m.lock.Unlock()
	delete(m.cache, guildID)
}

func (m *commandAliasManager) getOrLoadGuildAliases(guildID string) (*aliasCacheEntry, error) {
	m.lock.RLock()
	if entry, ok := m.cache[guildID]; ok {
		m.lock.RUnlock()
		return entry, nil
	}
	m.lock.RUnlock()

	aliases, err := m.store.GuildCommandAliases(guildID)
	if err != nil {
		return nil, err
	}
	entry := &aliasCacheEntry{aliases: convertAliases(aliases)}

	m.lock.Lock()
	defer m.lock.Unlock()
	m.cache[guildID] = entry
	return entry, nil
}

func convertAliases(in []CommandAlias) []*aliasInfo {
	result := make([]*aliasInfo, 0, len(in))
	for _, a := range in {
		alias := strings.TrimSpace(a.Alias)
		if alias == "" || a.Command == "" {
			continue
		}
		info := &aliasInfo{
			aliasTokens: splitToLower(alias),
			commandName: a.Command,
		}
		result = append(result, info)
	}
	return result
}

func splitToLower(s string) []string {
	parts := strings.Fields(s)
	if len(parts) == 0 {
		return nil
	}
	out := make([]string, len(parts))
	for i, p := range parts {
		out[i] = strings.ToLower(p)
	}
	return out
}

func equalTokens(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
