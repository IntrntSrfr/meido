package base

type PermissionType int

// PermissionType codes
const (
	PermissionTypeUser PermissionType = 1 << iota
	PermissionTypeChannel
	PermissionTypeGuild
)

type PermissionEntry struct {
	UID     string
	GuildID string
	ID      string
	Type    PermissionType
	Allow   bool
}
