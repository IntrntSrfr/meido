package mio

// Configurable provides a config interface.
// You are expected to implement it yourself
type Configurable interface {
	GetString(key string) string
	GetInt(key string) int
	GetStringSlice(key string) []string

	Set(key string, value interface{})
}
