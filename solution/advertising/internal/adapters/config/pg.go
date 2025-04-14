package config

import (
	"fmt"
	"github.com/spf13/viper"
)

type PGConfig interface {
	DSN() string
}

type pgConfig struct {
	host     string
	user     string
	password string
	port     int
	dbName   string
	sslMode  string
	timeZone string
}

func NewPGConfig(v *viper.Viper) PGConfig {
	return &pgConfig{
		host:     v.GetString("service.database.host"),
		user:     v.GetString("service.database.user"),
		password: v.GetString("service.database.password"),
		port:     v.GetInt("service.database.port"),
		dbName:   v.GetString("service.database.name"),
		sslMode:  v.GetString("service.database.ssl-mode"),
		timeZone: v.GetString("settings.timezone"),
	}
}

func (c *pgConfig) DSN() string {
	return fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%d sslmode=%s TimeZone=%s",
		c.user,
		c.password,
		c.dbName,
		c.host,
		c.port,
		c.sslMode,
		c.timeZone,
	)
}
