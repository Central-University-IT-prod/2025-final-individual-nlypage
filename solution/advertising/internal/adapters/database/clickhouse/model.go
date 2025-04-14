package clickhouse

import (
	"github.com/google/uuid"
)

type AdImpression struct {
	CampaignID   uuid.UUID
	AdvertiserID uuid.UUID
	ClientID     uuid.UUID
	Income       float64
	Day          int
	ViewCount    uint64
}

type AdClick struct {
	CampaignID   uuid.UUID
	AdvertiserID uuid.UUID
	ClientID     uuid.UUID
	Income       float64
	Day          int
}

type Stats struct {
	ImpressionsCount uint64
	ClicksCount      uint64
	Conversion       float64
	SpentImpressions float64
	SpentClicks      float64
	SpentTotal       float64
}

type StatsDaily struct {
	Stats
	Date int32
}

type UserCampaignStats struct {
	CampaignID       uuid.UUID
	ImpressionsCount uint64
	ClicksCount      uint64
	IsViewedByUser   bool
	IsClickedByUser  bool
}

// CampaignViews содержит информацию о просмотрах кампании
type CampaignViews struct {
	CampaignID uuid.UUID
	ViewCount  uint64
}

// ViewsGroup группирует кампании по количеству просмотров
type ViewsGroup struct {
	ViewCount uint64
	Campaigns []uuid.UUID
}
