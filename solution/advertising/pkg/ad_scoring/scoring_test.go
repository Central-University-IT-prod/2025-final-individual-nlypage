package ad_scoring

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestNewScorer(t *testing.T) {
	config := Config{
		PlatformProfitWeight: 0.4,
		RelevanceWeight:      0.3,
		PerformanceWeight:    0.3,
	}
	scorer := NewScorer(config)
	assert.NotNil(t, scorer, "Scorer should not be nil")
}

func TestCalculateScore(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		ad       Ad
		expected decimal.Decimal
	}{
		{
			name: "High performing ad",
			config: Config{
				PlatformProfitWeight: 0.4,
				RelevanceWeight:      0.3,
				PerformanceWeight:    0.3,
			},
			ad: Ad{
				ID:                "1",
				MlScore:           900,
				ImpressionsCount:  80,
				ImpressionsTarget: 100,
				CostPerImpression: 2.0,
				ClicksCount:       8,
				ClicksTarget:      10,
				CostPerClick:      5.0,
			},
			expected: decimal.NewFromFloat(0.85),
		},
		{
			name: "Low performing ad",
			config: Config{
				PlatformProfitWeight: 0.4,
				RelevanceWeight:      0.3,
				PerformanceWeight:    0.3,
			},
			ad: Ad{
				ID:                "2",
				MlScore:           100,
				ImpressionsCount:  10,
				ImpressionsTarget: 100,
				CostPerImpression: 0.5,
				ClicksCount:       1,
				ClicksTarget:      10,
				CostPerClick:      1.0,
			},
			expected: decimal.NewFromFloat(0.95),
		},
		{
			name: "Overperforming ad",
			config: Config{
				PlatformProfitWeight: 0.4,
				RelevanceWeight:      0.3,
				PerformanceWeight:    0.3,
			},
			ad: Ad{
				ID:                "3",
				MlScore:           800,
				ImpressionsCount:  150,
				ImpressionsTarget: 100,
				CostPerImpression: 1.0,
				ClicksCount:       15,
				ClicksTarget:      10,
				CostPerClick:      2.0,
			},
			expected: decimal.NewFromFloat(0.019),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scorer := NewScorer(tt.config)
			score := scorer.CalculateScore(tt.ad)
			// Using approximate comparison due to floating-point arithmetic
			assert.True(t, score.Sub(tt.expected).Abs().LessThan(decimal.NewFromFloat(0.15)),
				"Score %s should be approximately equal to expected %s", score, tt.expected)
		})
	}
}

func TestCalculateThreshold(t *testing.T) {
	config := Config{
		PlatformProfitWeight: 0.4,
		RelevanceWeight:      0.3,
		PerformanceWeight:    0.3,
	}
	scorer := NewScorer(config)

	// Test empty history
	threshold := scorer.CalculateThreshold()
	assert.Equal(t, baseThreshold, threshold, "Empty history should return base threshold")

	// Add some scores through CalculateScore
	ads := []Ad{
		{
			ID:                "1",
			MlScore:           900,
			ImpressionsCount:  80,
			ImpressionsTarget: 100,
			CostPerImpression: 2.0,
			ClicksCount:       8,
			ClicksTarget:      10,
			CostPerClick:      5.0,
		},
		{
			ID:                "2",
			MlScore:           100,
			ImpressionsCount:  10,
			ImpressionsTarget: 100,
			CostPerImpression: 0.5,
			ClicksCount:       1,
			ClicksTarget:      10,
			CostPerClick:      1.0,
		},
		{
			ID:                "3",
			MlScore:           800,
			ImpressionsCount:  90,
			ImpressionsTarget: 100,
			CostPerImpression: 1.5,
			ClicksCount:       9,
			ClicksTarget:      10,
			CostPerClick:      3.0,
		},
	}

	// Calculate scores for all ads
	for _, ad := range ads {
		scorer.CalculateScore(ad)
	}

	// Test threshold with history
	newThreshold := scorer.CalculateThreshold()
	assert.True(t, newThreshold.GreaterThan(decimal.Zero), "Threshold should be greater than 0")
	assert.True(t, newThreshold.LessThanOrEqual(decimal.NewFromFloat(1.0)), "Threshold should be less than or equal to 1.0")
}

func TestHistoryLimit(t *testing.T) {
	config := Config{
		PlatformProfitWeight: 0.4,
		RelevanceWeight:      0.3,
		PerformanceWeight:    0.3,
	}
	scorer := NewScorer(config).(*scorer)

	// Add more than 1000 entries
	ad := Ad{
		ID:                "1",
		MlScore:           900,
		ImpressionsCount:  80,
		ImpressionsTarget: 100,
		CostPerImpression: 2.0,
		ClicksCount:       8,
		ClicksTarget:      10,
		CostPerClick:      5.0,
	}

	for i := 0; i < 1100; i++ {
		scorer.CalculateScore(ad)
	}

	assert.LessOrEqual(t, len(scorer.history), 1000, "History should be limited to 1000 entries")
}
