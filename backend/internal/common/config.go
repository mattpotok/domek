package common

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Telegram *TelegramConfig
}

type TelegramConfig struct {
	BotToken string
}

func getEnvRequired(key string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		log.Fatalf("Environment variable '%s' is not set", key)
	}

	return value
}

func LoadConfig() *Config {
	godotenv.Load()

	cfg := &Config{
		Telegram: &TelegramConfig{
			BotToken: getEnvRequired("TELEGRAM_BOT_TOKEN"),
		},
	}

	return cfg
}
