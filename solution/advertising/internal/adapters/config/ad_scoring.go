package config

import (
	"github.com/spf13/viper"
	"time"
)

type AdScoringConfig interface {
	PlatformProfitWeight() float64
	RelevanceWeight() float64
	PerformanceWeight() float64
	UpdateInterval() time.Duration
}

type adScoringConfig struct {
	platformProfitWeight float64
	relevanceWeight      float64
	performanceWeight    float64
	updateInterval       time.Duration
}

func NewAdScoringConfig(v *viper.Viper) AdScoringConfig {
	return &adScoringConfig{
		platformProfitWeight: v.GetFloat64("service.backend.settings.ad-scoring.weights.profit"),
		relevanceWeight:      v.GetFloat64("service.backend.settings.ad-scoring.weights.relevance"),
		performanceWeight:    v.GetFloat64("service.backend.settings.ad-scoring.weights.performance"),
		updateInterval:       v.GetDuration("service.backend.settings.ad-scoring.interval"),
	}
}

func (c *adScoringConfig) PlatformProfitWeight() float64 {
	return c.platformProfitWeight
}

func (c *adScoringConfig) RelevanceWeight() float64 {
	return c.relevanceWeight
}

func (c *adScoringConfig) PerformanceWeight() float64 {
	return c.performanceWeight
}

func (c *adScoringConfig) UpdateInterval() time.Duration {
	return c.updateInterval
}
