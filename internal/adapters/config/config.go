package config

import (
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type Config struct {
	DB_HOST    string `validate:"required"`
	DB_PORT    string `validate:"required"`
	DB_USER    string `validate:"required"`
	DB_PASS    string `validate:"required"`
	DB_NAME    string `validate:"required"`
	HTTP_PORT  string `validate:"required"`
	QUERY_PASS string `validate:"required"`
}

func Load() (*Config, error) {
	viper.AutomaticEnv()

	viper.SetConfigFile(".env")
	_ = viper.ReadInConfig()

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	cfg := &Config{
		DB_HOST:    viper.GetString("DB_HOST"),
		DB_PORT:    viper.GetString("DB_PORT"),
		DB_USER:    viper.GetString("DB_USER"),
		DB_PASS:    viper.GetString("DB_PASS"),
		DB_NAME:    viper.GetString("DB_NAME"),
		HTTP_PORT:  viper.GetString("HTTP_PORT"),
		QUERY_PASS: viper.GetString("QUERYPASS"),
	}

	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
