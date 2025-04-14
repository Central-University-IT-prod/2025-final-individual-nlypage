package dto

import "github.com/google/uuid"

// Stats содержит агрегированную статистику для рекламной кампании или рекламодателя
type Stats struct {
	ImpressionsCount int     `json:"impressions_count" validate:"required,gte=0"`
	ClicksCount      int     `json:"clicks_count" validate:"required,gte=0"`
	Conversion       float64 `json:"conversion" validate:"required,gte=0"`
	SpentImpressions float64 `json:"spent_impressions" validate:"required,gte=0"`
	SpentClicks      float64 `json:"spent_clicks" validate:"required,gte=0"`
	SpentTotal       float64 `json:"spent_total" validate:"required,gte=0"`
}

type CampaignStatsGet struct {
	CampaignID uuid.UUID `param:"campaignId" validate:"required"`
}

type CampaignDailyStatsGet struct {
	CampaignID uuid.UUID `param:"campaignId" validate:"required"`
}

type AdvertiserStatsGet struct {
	AdvertiserID uuid.UUID `param:"advertiserId" validate:"required"`
}

type AdvertiserDailyStatsGet struct {
	AdvertiserID uuid.UUID `param:"advertiserId" validate:"required"`
}

// StatsDaily представляет ежедневную статистику с указанием дня
type StatsDaily struct {
	Stats
	Date int `json:"date" validate:"required,gte=0"`
}
