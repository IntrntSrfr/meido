package mio

// Configurable provides a config interface.
// You are expected to implement it yourself
type Configurable interface {
	GetString(key string) string
	GetInt(key string) int
	GetStringSlice(key string) []string

	Set(key string, value interface{})
}

type ConfigBase struct {
	data map[string]interface{}
}

func NewConfig() *ConfigBase {
	return &ConfigBase{
		data: make(map[string]interface{}),
	}
}

func (c *ConfigBase) GetString(key string) string {
	if v, found := c.data[key]; found {
		if vt, ok := v.(string); ok {
			return vt
		}
	}
	return ""
}

func (c *ConfigBase) GetInt(key string) int {
	if v, found := c.data[key]; found {
		if vt, ok := v.(int); ok {
			return vt
		}
	}
	return -1
}

func (c *ConfigBase) GetStringSlice(key string) []string {
	if v, found := c.data[key]; found {
		if vt, ok := v.([]string); ok {
			return vt
		}
	}
	return []string{}
}

func (c *ConfigBase) Set(key string, value interface{}) {
	c.data[key] = value
}
