package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DbUrl string
	Port  string
}

func Load() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}
	return &Config{
		DbUrl: os.Getenv("DB_URL"),
		Port:  os.Getenv("PORT"),
	}, nil
}
