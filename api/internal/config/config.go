package config

import (
	"log/slog"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env        string         `yaml:"env" env:"ENV" env-default:"local" env-required:"true"`
	Db         DatabaseConfig `yaml:"db"`
	HTTPServer HTTPServer     `yaml:"http_server"`
	JWTSecret  string         `yaml:"jwt_secret" env:"JWT_SECRET" env-required:"true"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env:"HTTP_SERVER_ADDRESS" env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"timeout" env:"HTTP_SERVER_TIMEOUT" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env:"HTTP_SERVER_IDLE_TIMEOUT" env-default:"60s"`
}

type DatabaseConfig struct {
	ConnectionString string `yaml:"connection_string" env:"DB_CONNECTION_STRING" env-required:"true"`
	Driver           string `yaml:"driver" env:"DB_DRIVER" env-required:"true"`
}

func MustLoad() *Config {
	if err := godotenv.Load(); err != nil {
		slog.Warn("No .env file found, using environment variables", "err", err)
	}

	var cfg Config

	configPath := os.Getenv("CONFIG_PATH")
	if configPath != "" {
		if _, err := os.Stat(configPath); err == nil {
			if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
				slog.Error("cannot read config from file", "err", err)
				os.Exit(1)
			}
		} else {
			slog.Info("Config file does not exist, proceeding with environment variables", "configPath", configPath)
		}
	}

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		slog.Warn("error reading environment variables", "err", err)
	}

	if cfg.JWTSecret == "" {
		slog.Error("JWT_SECRET is required but not set")
		os.Exit(1)
	}
	if cfg.Db.ConnectionString == "" || cfg.Db.Driver == "" {
		slog.Error("Database configuration is required but not set")
		os.Exit(1)
	}
	if cfg.HTTPServer.Address == "" {
		slog.Error("HTTPServer address is required but not set")
		os.Exit(1)
	}

	return &cfg
}
