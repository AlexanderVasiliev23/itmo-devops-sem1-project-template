package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"os"
)

type Config struct {
	ServerPort int `env:"SERVER_PORT" envDefault:"8080"`
	PGConfig   PGConfig
}

type PGConfig struct {
	Host string `env:"POSTGRES_HOST" env-required:"true"`
	Port int    `env:"POSTGRES_PORT" env-required:"true"`
	User string `env:"POSTGRES_USER" env-required:"true"`
	Pass string `env:"POSTGRES_PASSWORD" env-required:"true"`
	Name string `env:"POSTGRES_DB" env-required:"true"`
}

func Load() (*Config, error) {
	cfg := &Config{}

	if _, err := os.Stat("../.env"); err == nil {
		if err := godotenv.Load("../.env"); err != nil {
			return nil, err
		}
	}

	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(".env"); err != nil {
			return nil, err
		}
	}

	if _, err := os.Stat(".env.dist"); err == nil {
		if err := godotenv.Load(".env.dist"); err != nil {
			return nil, err
		}
	}

	if err := cleanenv.ReadEnv(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
