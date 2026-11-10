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

// TP is the connection to the kdb+ ticker plant.
type TP struct {
	Host string // TP_HOST, e.g. localhost
	Port int    // TP_PORT, e.g. 5020
	Auth string // TP_AUTH, "user:password"; empty if the tp has no -u/-U
}

// LoadTP reads the ticker plant settings from the environment, loading .env
// first if there is one. TP_HOST and TP_PORT are required.
func LoadTP() TP {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	viper.AutomaticEnv()

	cfg := TP{
		Host: viper.GetString("TP_HOST"),
		Port: viper.GetInt("TP_PORT"),
		Auth: viper.GetString("TP_AUTH"),
	}

	if cfg.Host == "" || cfg.Port <= 0 {
		log.Fatal("Missing ticker plant address: set TP_HOST and TP_PORT")
	}

	return cfg
}
