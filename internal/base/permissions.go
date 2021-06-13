package base

import (
	"errors"
	"sort"
	"sync/atomic"
)

// PermissionType represents what a permission override should look for
type PermissionType int

const (
	PermissionTypeUser = 1 << iota
	PermissionTypeChannel
	PermissionTypeGuild
)

var counter atomic.Value

type PermissionOverride struct {
	UID     int
	GuildID string
	Type    PermissionType

	// TypeID is for the ID for the type this permission belongs to
	TypeID string

	// Allow  is whether this override should allow or disallow
	Allow bool
}

type SortedByUID []*PermissionOverride

func (s SortedByUID) Len() int {
	return len(s)
}

func (s SortedByUID) Less(i, j int) bool {
	return s[i].UID < s[j].UID
}

func (s SortedByUID) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (p *PermissionOverride) Validate() error {
	if p.UID == 0 || p.GuildID == "" || p.Type == 0 || p.TypeID == "" {
		return errors.New("one or more fields are empty")
	}
	return nil
}

type PermissionHandler struct {
	overrides []*PermissionOverride
}

func NewPermissionHandler() *PermissionHandler {
	return &PermissionHandler{
		overrides: []*PermissionOverride{},
	}
}

func (p *PermissionHandler) AddOverride(o *PermissionOverride) error {

	if err := o.Validate(); err != nil {
		return err
	}

	o.UID = (counter.Load()).(int)
	counter.Store(o.UID + 1)

	p.overrides = append(p.overrides, o)
	return nil
}

func (p *PermissionHandler) GetOverrides() []*PermissionOverride {
	return p.overrides
}

func (p *PermissionHandler) GetGuildOverrides(guildID string) []*PermissionOverride {
	var res []*PermissionOverride

	for _, o := range p.overrides {
		if o.GuildID == guildID {
			res = append(res, o)
		}
	}

	sort.Sort(SortedByUID(res))

	return res
}

func (p *PermissionHandler) RemoveOverride(id int) {
	i := p.getOverrideIndex(id)
	if i < 0 {
		return
	}

	p.overrides[i] = p.overrides[len(p.overrides)]
	p.overrides = p.overrides[:len(p.overrides)-1]

}

func (p *PermissionHandler) getOverrideIndex(id int) int {
	for i := range p.overrides {
		if p.overrides[i].UID == id {
			return i
		}
	}

	return -1
}

func (p *PermissionHandler) Calculate() (bool, error) {

	return false, nil
}
