package config

import (
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type AppConfig struct {
	Port        string `envconfig:"PORT" required:"true"`
	DatabaseURL string `envconfig:"DATABASE_URL" required:"true"`
}

func NewAppConfig() *AppConfig {
	godotenv.Load()
	appCfg := AppConfig{}
	envconfig.MustProcess("", &appCfg)
	return &appCfg
}
