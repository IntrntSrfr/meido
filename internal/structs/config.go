package structs

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/intrntsrfr/meido/pkg/utils"
)

// Config is the config struct for the bot
type Config struct {
	*utils.Config
}

type jsonConfig struct {
	Token            string   `json:"token"`
	Shards           int      `json:"shards"`
	ConnectionString string   `json:"connection_string"`
	OwnerIds         []string `json:"owner_ids"`
	DmLogChannels    []string `json:"dm_log_channels"`
	OwoToken         string   `json:"owo_token"`
	YouTubeToken     string   `json:"youtube_key"`
	OpenWeatherKey   string   `json:"open_weather_api_key"`
}

func LoadConfig(cfg *utils.Config) error {
	if err := loadJson(cfg); err != nil {
		return err
	}

	loadEnvs(cfg)
	return nil
}

func loadJson(cfg *utils.Config) error {
	file, err := os.ReadFile("./config.json")
	if err != nil {
		return err
	}

	var jsonCfg jsonConfig
	err = json.Unmarshal(file, &jsonCfg)
	if err != nil {
		return err
	}

	cfg.Set("shards", jsonCfg.Shards)
	cfg.Set("token", jsonCfg.Token)
	cfg.Set("connection_string", jsonCfg.ConnectionString)
	cfg.Set("owner_ids", jsonCfg.OwnerIds)
	cfg.Set("dm_log_channels", jsonCfg.DmLogChannels)
	cfg.Set("owo_token", jsonCfg.OwoToken)
	cfg.Set("youtube_token", jsonCfg.YouTubeToken)
	cfg.Set("open_weather_key", jsonCfg.OpenWeatherKey)
	return nil
}

func loadEnvs(cfg *utils.Config) {
	if e := os.Getenv("DISCORD_TOKEN"); e != "" {
		cfg.Set("token", os.Getenv("DISCORD_TOKEN"))
	}

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	connStr := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable",
		dbHost,
		dbPort,
		dbName,
		dbUser,
		dbPassword,
	)
	if dbHost != "" && dbPort != "" && dbName != "" && dbUser != "" && dbPassword != "" {
		cfg.Set("connection_string", connStr)
	}
}
