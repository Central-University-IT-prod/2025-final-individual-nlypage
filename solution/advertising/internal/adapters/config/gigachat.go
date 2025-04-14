package config

import "github.com/spf13/viper"

type GigachatConfig interface {
	AuthKey() string
}

type gigachatConfig struct {
	authKey string
}

func NewGigachatConfig(v *viper.Viper) GigachatConfig {
	return &gigachatConfig{
		authKey: v.GetString("service.gigachat.auth-key"),
	}
}

func (c *gigachatConfig) AuthKey() string {
	return c.authKey
}
