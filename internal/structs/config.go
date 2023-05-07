package structs

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config is the config struct for the bot
type Config struct {
	data map[string]interface{}
}

type jsonConfig struct {
	Token            string   `json:"token"`
	ConnectionString string   `json:"connection_string"`
	OwnerIds         []string `json:"owner_ids"`
	DmLogChannels    []string `json:"dm_log_channels"`
	OwoToken         string   `json:"owo_token"`
	YouTubeToken     string   `json:"youtube_key"`
	OpenWeatherKey   string   `json:"open_weather_api_key"`
}

func NewConfig() *Config {
	return &Config{
		data: make(map[string]interface{}),
	}
}

func LoadConfig() (*Config, error) {
	config := &Config{
		data: make(map[string]interface{}),
	}

	if err := config.loadJson(); err != nil {
		return nil, err
	}

	if os.Getenv("USE_ENV_CONFIG") != "" {
		config.loadEnvs()
	}

	fmt.Println(config)
	return config, nil
}

func (c *Config) loadEnvs() {
	c.Set("token", os.Getenv("DISCORD_TOKEN"))

	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbPort := os.Getenv("DB_PORT")
	dbHost := os.Getenv("DB_HOST")
	connStr := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable",
		dbHost,
		dbPort,
		dbName,
		dbUser,
		dbPassword,
	)
	c.Set("connection_string", connStr)
}

func (c *Config) loadJson() error {
	file, err := os.ReadFile("./config.json")
	if err != nil {
		return err
	}

	var jsonCfg jsonConfig
	err = json.Unmarshal(file, &jsonCfg)
	if err != nil {
		return err
	}

	c.Set("owner_ids", jsonCfg.OwnerIds)
	c.Set("dm_log_channels", jsonCfg.DmLogChannels)
	c.Set("owo_token", jsonCfg.OwoToken)
	c.Set("youtube_token", jsonCfg.YouTubeToken)
	c.Set("open_weather_key", jsonCfg.OpenWeatherKey)
	return nil
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

func (c *Config) Set(key string, value interface{}) {
	c.data[key] = value
}