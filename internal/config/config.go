package config

import (
	"fmt"

	"github.com/BurntSushi/toml"
	"github.com/go-playground/validator/v10"
)

var Validator = validator.New()

type ServerConfig struct {
	Addr string `toml:"addr" validate:"required,hostname_port"`
}

type PostgresConfig struct {
	Host     string `toml:"host" validate:"required,hostname_port"`
	UserName string `toml:"user_name" validate:"required"`
	DBName   string `toml:"db_name" validate:"required"`
	Password string `toml:"password" validate:"required"`
}

type Config struct {
	ServerConf *ServerConfig   `toml:"server"`
	PgConfig   *PostgresConfig `toml:"database"`
}

func (cfg *PostgresConfig) DSN() string {
	return "postgres://" + cfg.UserName + ":" + cfg.Password + "@" +
		cfg.Host + "/" + cfg.DBName + "?sslmode=disable"
}

func ParseAndValidate(filename string) (Config, error) {
	config := Config{}
	if _, err := toml.DecodeFile(filename, &config); err != nil {
		return Config{}, fmt.Errorf("decoding config file error: %v", err)
	}

	err := Validator.Struct(config)
	if err != nil {
		return Config{}, fmt.Errorf("config validation error: %v", err)
	}
	return config, nil
}
