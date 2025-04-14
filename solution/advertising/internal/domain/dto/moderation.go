package dto

import "github.com/google/uuid"

type CampaignApprove struct {
	CampaignID uuid.UUID `param:"campaignId" validate:"required"`
}
