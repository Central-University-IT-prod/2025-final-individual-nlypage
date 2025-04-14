package config

import "github.com/spf13/viper"

type ClickHouseConfig interface {
	Host() string
	Port() string
	Database() string
	Username() string
	Password() string
	Debug() bool
}

type clickHouseConfig struct {
	host     string
	port     string
	database string
	username string
	password string
	debug    bool
}

func NewClickHouseConfig(v *viper.Viper) ClickHouseConfig {
	return &clickHouseConfig{
		host:     v.GetString("service.clickhouse.host"),
		port:     v.GetString("service.clickhouse.port"),
		database: v.GetString("service.clickhouse.database"),
		username: v.GetString("service.clickhouse.username"),
		password: v.GetString("service.clickhouse.password"),
		debug:    v.GetBool("service.debug"),
	}
}

func (c *clickHouseConfig) Host() string {
	return c.host
}

func (c *clickHouseConfig) Port() string {
	return c.port
}

func (c *clickHouseConfig) Database() string {
	return c.database
}

func (c *clickHouseConfig) Username() string {
	return c.username
}

func (c *clickHouseConfig) Password() string {
	return c.password
}

func (c *clickHouseConfig) Debug() bool {
	return c.debug
}
