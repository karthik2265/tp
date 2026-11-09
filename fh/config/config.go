package config

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	APIKey    string
	APISecret string
	Feed      string
}

func LoadConfig() Config {
	// Load .env into process environment.
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	viper.AutomaticEnv()

	cfg := Config{
		APIKey:    viper.GetString("APCA_API_KEY_ID"),
		APISecret: viper.GetString("APCA_API_SECRET_KEY"),
		Feed:      viper.GetString("ALPACA_FEED"),
	}

	if cfg.APIKey == "" || cfg.APISecret == "" {
		log.Fatal("Missing Alpaca API credentials")
	}

	if cfg.Feed == "" {
		cfg.Feed = "iex"
	}

	return cfg
}
