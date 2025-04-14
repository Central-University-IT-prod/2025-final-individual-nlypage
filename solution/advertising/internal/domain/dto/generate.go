package dto

import "github.com/google/uuid"

type GenerateAdTextRequest struct {
	AdvertiserID   uuid.UUID `param:"advertiserId" validate:"required"`
	AdTitle        string    `json:"ad_title" validate:"required"`
	AdditionalInfo string    `json:"additional_info"`
}

type GenerateAdTextResponse struct {
	AdText string `json:"ad_text" validate:"required"`
}
