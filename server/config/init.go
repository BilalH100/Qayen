package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DbUrl                 string
	Port                  string
	ApproximateTestAlerts bool
}

func Load() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}
	return &Config{
		DbUrl:                 os.Getenv("DB_URL"),
		Port:                  os.Getenv("PORT"),
		ApproximateTestAlerts: os.Getenv("QAYEN_ENABLE_APPROXIMATE_ALERTS") == "true",
	}, nil
}
