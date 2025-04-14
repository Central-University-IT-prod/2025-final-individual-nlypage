package dto

import "github.com/google/uuid"

// Advertiser представляет рекламодателя
type Advertiser struct {
	AdvertiserID uuid.UUID `json:"advertiser_id" validate:"required"`
	Name         string    `json:"name" validate:"required"`
}

type AdvertiserGet struct {
	AdvertiserID uuid.UUID `param:"advertiserId" validate:"required"`
}

// AdvertiserUpsert представляет DTO для создания/обновления рекламодателя
type AdvertiserUpsert struct {
	AdvertiserID uuid.UUID `json:"advertiser_id" validate:"required"`
	Name         string    `json:"name" validate:"required"`
}

// MlScoreUpsert представляет DTO для создания/обновления ML-score для пары клиент-рекламодатель
type MlScoreUpsert struct {
	ClientID     uuid.UUID `json:"client_id" validate:"required"`
	AdvertiserID uuid.UUID `json:"advertiser_id" validate:"required"`
	Score        int64     `json:"score" validate:"gte=0"`
}
