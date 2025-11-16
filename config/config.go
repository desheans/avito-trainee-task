package config

import "github.com/ilyakaznacheev/cleanenv"

type Env string

const (
	Development Env = "dev"
	Production  Env = "prod"
)

type Config struct {
	Postgres Postgres
	Server   Server

	Env Env `env:"ENV" env-default:"dev"`
}

type Postgres struct {
	PGURL string `env:"PG_URL" env-required:"true"`
}

type Server struct {
	Port string `env:"PORT" env-required:"true"`
}

func LoadConfig() (*Config, error) {
	var cfg Config
	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
