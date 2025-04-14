package config

import (
	"github.com/spf13/viper"
	"time"
)

type LoggerConfig interface {
	Debug() bool
	LogToFile() bool
	LogsDir() string
	TimeLocation() *time.Location
}

type loggerConfig struct {
	debug        bool
	logToFile    bool
	logsDir      string
	timeLocation *time.Location
}

func NewLoggerConfig(v *viper.Viper, location *time.Location) LoggerConfig {
	return &loggerConfig{
		debug:        v.GetBool("settings.debug"),
		logToFile:    v.GetBool("settings.logger.log-to-file"),
		logsDir:      v.GetString("settings.logger.logs-dir"),
		timeLocation: location,
	}
}

func (l *loggerConfig) Debug() bool {
	return l.debug
}

func (l *loggerConfig) LogToFile() bool {
	return l.logToFile
}

func (l *loggerConfig) LogsDir() string {
	return l.logsDir
}

func (l *loggerConfig) TimeLocation() *time.Location {
	return l.timeLocation
}
