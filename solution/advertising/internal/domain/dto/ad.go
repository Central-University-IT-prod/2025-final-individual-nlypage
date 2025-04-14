package dto

import "github.com/google/uuid"

// Ad представляет рекламное объявление
type Ad struct {
	AdID         uuid.UUID `json:"ad_id" validate:"required"`
	AdTitle      string    `json:"ad_title" validate:"required"`
	AdText       string    `json:"ad_text" validate:"required"`
	ImageURL     string    `json:"image_url"`
	AdvertiserID uuid.UUID `json:"advertiser_id" validate:"required"`
}

type ClientAdGet struct {
	ClientID uuid.UUID `query:"client_id" validate:"required"`
}

type ClientAdClick struct {
	AdID     uuid.UUID `param:"adId" validate:"required"`
	ClientID uuid.UUID `json:"client_id" validate:"required"`
}
