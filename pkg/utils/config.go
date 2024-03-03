package utils

type Config struct {
	data map[string]interface{}
}

func NewConfig() *Config {
	return &Config{
		data: make(map[string]interface{}),
	}
}

func (c *Config) Set(key string, value interface{}) {
	c.data[key] = value
}

func (c *Config) GetString(key string) string {
	if v, found := c.data[key]; found {
		if vt, ok := v.(string); ok {
			return vt
		}
	}
	return ""
}

func (c *Config) GetInt(key string) int {
	if v, found := c.data[key]; found {
		if vt, ok := v.(int); ok {
			return vt
		}
	}
	return -1
}

func (c *Config) GetStringSlice(key string) []string {
	if v, found := c.data[key]; found {
		if vt, ok := v.([]string); ok {
			return vt
		}
	}
	return []string{}
}
