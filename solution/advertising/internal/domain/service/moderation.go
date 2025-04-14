package service

import (
	"context"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"nlypage-final/internal/adapters/database/postgres/ent"
	"nlypage-final/internal/adapters/database/postgres/ent/campaign"
	"nlypage-final/internal/domain/dto"
)

type ModerationService interface {
	GetNotModeratedCampaigns(ctx context.Context) ([]*dto.Campaign, error)
	ApproveCampaign(ctx context.Context, campaignID uuid.UUID) error
}

type moderationService struct {
	db *ent.Client
}

func NewModerationService(db *ent.Client) ModerationService {
	return &moderationService{
		db: db,
	}
}

func (s *moderationService) GetNotModeratedCampaigns(ctx context.Context) ([]*dto.Campaign, error) {
	campaigns, err := s.db.Campaign.Query().
		Where(
			campaign.Moderated(false),
		).Order(ent.Asc(campaign.FieldStartDate)).
		All(ctx)
	if err != nil {
		return nil, &echo.HTTPError{
			Message: err.Error(),
			Code:    echo.ErrInternalServerError.Code,
		}
	}

	var result []*dto.Campaign
	for _, camp := range campaigns {
		result = append(result, &dto.Campaign{
			CampaignID:        camp.ID,
			AdvertiserID:      camp.AdvertiserID,
			ImpressionsLimit:  camp.ImpressionsLimit,
			ClicksLimit:       camp.ClicksLimit,
			CostPerImpression: camp.CostPerImpression,
			CostPerClick:      camp.CostPerClick,
			AdTitle:           camp.AdTitle,
			AdText:            camp.AdText,
			StartDate:         camp.StartDate,
			EndDate:           camp.EndDate,
			Moderated:         camp.Moderated,
		})
	}
	return result, nil
}

func (s *moderationService) ApproveCampaign(ctx context.Context, campaignID uuid.UUID) error {
	affected, err := s.db.Campaign.Update().
		Where(campaign.ID(campaignID)).
		SetModerated(true).
		Save(ctx)
	if err != nil {
		return &echo.HTTPError{
			Message: err.Error(),
			Code:    echo.ErrInternalServerError.Code,
		}
	}

	if affected == 0 {
		return &echo.HTTPError{
			Message: "campaign not found",
			Code:    echo.ErrNotFound.Code,
		}
	}

	return nil
}
