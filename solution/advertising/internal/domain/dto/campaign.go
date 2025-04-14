package dto

import "github.com/google/uuid"

// Campaign представляет рекламную кампанию
type Campaign struct {
	CampaignID        uuid.UUID `json:"campaign_id" validate:"required"`
	AdvertiserID      uuid.UUID `json:"advertiser_id" validate:"required"`
	ImpressionsLimit  int       `json:"impressions_limit" validate:"required,gt=0,gtfield=ClicksLimit"`
	ClicksLimit       int       `json:"clicks_limit" validate:"required,gt=0"`
	CostPerImpression float64   `json:"cost_per_impression" validate:"gte=0"`
	CostPerClick      float64   `json:"cost_per_click" validate:"gte=0"`
	AdTitle           string    `json:"ad_title" validate:"required"`
	AdText            string    `json:"ad_text" validate:"required"`
	ImageURL          string    `json:"image_url"`
	StartDate         int       `json:"start_date" validate:"gte=0"`
	EndDate           int       `json:"end_date" validate:"gte=0,gtefield=StartDate"`
	Moderated         bool      `json:"moderated"`
	Targeting         Targeting `json:"targeting" validate:"required"`
}

// CampaignCreate представляет DTO для создания новой рекламной кампании
type CampaignCreate struct {
	AdvertiserID      uuid.UUID  `param:"advertiserId" validate:"required"`
	ImpressionsLimit  int        `json:"impressions_limit" validate:"required,gt=0,gtefield=ClicksLimit"`
	ClicksLimit       int        `json:"clicks_limit" validate:"required,gt=0"`
	CostPerImpression float64    `json:"cost_per_impression" validate:"gte=0"`
	CostPerClick      float64    `json:"cost_per_click" validate:"gte=0"`
	AdTitle           string     `json:"ad_title" validate:"required"`
	AdText            string     `json:"ad_text" validate:"required"`
	StartDate         int        `json:"start_date" validate:"gte=0"`
	EndDate           int        `json:"end_date" validate:"gte=0,gtefield=StartDate"`
	Targeting         *Targeting `json:"targeting,omitempty"`
}

// Targeting описывает настройки таргетирования для рекламной кампании
type Targeting struct {
	Gender   *string `json:"gender" validate:"omitempty,oneof=MALE FEMALE ALL"`
	AgeFrom  *int    `json:"age_from" validate:"omitempty,gte=0"`
	AgeTo    *int    `json:"age_to" validate:"omitempty,gte=0,gtefield=AgeFrom"`
	Location *string `json:"location" validate:"omitempty"`
}

type CampaignGet struct {
	AdvertiserID uuid.UUID `param:"advertiserId" validate:"required"`
	CampaignID   uuid.UUID `param:"campaignId" validate:"required"`
}

type CampaignsGet struct {
	AdvertiserID uuid.UUID `param:"advertiserId" validate:"required"`
	Size         int       `query:"size"`
	Page         int       `query:"page"`
}

// CampaignUpdate представляет DTO для обновления параметров кампании
type CampaignUpdate struct {
	AdvertiserID      uuid.UUID  `param:"advertiserId" validate:"required"`
	CampaignID        uuid.UUID  `param:"campaignId" validate:"required"`
	ImpressionsLimit  int        `json:"impressions_limit" validate:"required,gt=0,gtfield=ClicksLimit"`
	ClicksLimit       int        `json:"clicks_limit" validate:"required,gt=0"`
	CostPerImpression float64    `json:"cost_per_impression" validate:"gt=0"`
	CostPerClick      float64    `json:"cost_per_click" validate:"gt=0"`
	AdTitle           string     `json:"ad_title" validate:"required"`
	AdText            string     `json:"ad_text" validate:"required"`
	StartDate         int        `json:"start_date" validate:"gte=0"`
	EndDate           int        `json:"end_date" validate:"gte=0,gtefield=StartDate"`
	Targeting         *Targeting `json:"targeting"`
}

type CampaignDelete struct {
	AdvertiserID uuid.UUID `param:"advertiserId" validate:"required"`
	CampaignID   uuid.UUID `param:"campaignId" validate:"required"`
}

type CampaignUploadImageRequest struct {
	AdvertiserID uuid.UUID `param:"advertiserId" validate:"required"`
	CampaignID   uuid.UUID `param:"campaignId" validate:"required"`
}

type CampaignRemoveImageRequest struct {
	AdvertiserID uuid.UUID `param:"advertiserId" validate:"required"`
	CampaignID   uuid.UUID `param:"campaignId" validate:"required"`
}

type CampaignImageURL struct {
	AdvertiserID uuid.UUID `json:"advertiser_id" validate:"required"`
	CampaignID   uuid.UUID `json:"campaign_id" validate:"required"`
	ImageURL     string    `json:"image_url" validate:"required"`
}
