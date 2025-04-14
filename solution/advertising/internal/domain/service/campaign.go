package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"io"
	"nlypage-final/internal/adapters/database/minio"
	"nlypage-final/internal/adapters/database/postgres/ent"
	"nlypage-final/internal/adapters/database/postgres/ent/campaign"
	"nlypage-final/internal/adapters/database/postgres/ent/targeting"
	"nlypage-final/internal/domain/common/errorz"
	"nlypage-final/internal/domain/dto"
	"nlypage-final/pkg/logger"
)

type campaignTimeService interface {
	Now() *dto.CurrentDate
}

type campaignClickhouseRepository interface {
	DeleteStatsByCampaignID(ctx context.Context, campaignID uuid.UUID) error
}

type adImagesRepository interface {
	UploadImage(ctx context.Context, campaignID string, imageData io.Reader) (string, error)
	GetImage(ctx context.Context, campaignID string) (string, error)
	DeleteImage(ctx context.Context, campaignID string) error
}

type CampaignService interface {
	Create(ctx context.Context, campaign *dto.CampaignCreate) (*dto.Campaign, error)
	GetByID(ctx context.Context, campaignID uuid.UUID, advertiserID uuid.UUID) (*dto.Campaign, error)
	Get(ctx context.Context, advertiserID uuid.UUID, size, page int) ([]*dto.Campaign, error)
	Delete(ctx context.Context, campaignID uuid.UUID, advertiserID uuid.UUID) error
	Update(ctx context.Context, campaignUpdate *dto.CampaignUpdate) (*dto.Campaign, error)
	UploadImage(ctx context.Context, uploadImageRequest *dto.CampaignUploadImageRequest, imageData io.Reader) (*dto.CampaignImageURL, error)
	RemoveImage(ctx context.Context, removeImageRequest *dto.CampaignRemoveImageRequest) error
}

type campaignService struct {
	db                   *ent.Client
	timeService          campaignTimeService
	clickhouseRepository campaignClickhouseRepository
	adImagesRepository   adImagesRepository
	moderation           bool
}

func NewCampaignService(
	db *ent.Client,
	timeService campaignTimeService,
	clickhouseRepository campaignClickhouseRepository,
	adImagesRepository adImagesRepository,
	moderation bool,
) CampaignService {
	return &campaignService{
		db:                   db,
		timeService:          timeService,
		clickhouseRepository: clickhouseRepository,
		adImagesRepository:   adImagesRepository,
		moderation:           moderation,
	}
}

func (s *campaignService) Create(ctx context.Context, campaign *dto.CampaignCreate) (*dto.Campaign, error) {
	if campaign.StartDate < s.timeService.Now().CurrentDate {
		return nil, &echo.HTTPError{
			Message: "start date must be gte than current date",
			Code:    echo.ErrBadRequest.Code,
		}
	}

	if campaign.EndDate < campaign.StartDate {
		return nil, &echo.HTTPError{
			Message: "end date must be gte than start date",
			Code:    echo.ErrBadRequest.Code,
		}
	}

	createdCampaign, err := s.db.Campaign.Create().
		SetAdvertiserID(campaign.AdvertiserID).
		SetImpressionsLimit(campaign.ImpressionsLimit).
		SetClicksLimit(campaign.ClicksLimit).
		SetCostPerImpression(campaign.CostPerImpression).
		SetCostPerClick(campaign.CostPerClick).
		SetAdTitle(campaign.AdTitle).
		SetAdText(campaign.AdText).
		SetStartDate(campaign.StartDate).
		SetEndDate(campaign.EndDate).
		SetModerated(!s.moderation).
		Save(ctx)
	if err != nil {
		if ent.IsValidationError(err) {
			return nil, &echo.HTTPError{
				Message: err.Error(),
				Code:    echo.ErrBadRequest.Code,
			}
		}
		logger.Log.Errorf("failed to create campaign: %v", err)
		return nil, errorz.ErrInternal
	}

	request := s.db.Targeting.Create()
	if campaign.Targeting != nil {
		request = request.
			SetCampaign(createdCampaign).
			SetNillableAgeFrom(campaign.Targeting.AgeFrom).
			SetNillableAgeTo(campaign.Targeting.AgeTo)
		if campaign.Targeting.Gender != nil {
			request.SetGender(targeting.Gender(*campaign.Targeting.Gender))
		}
	}

	createdTargeting, err := request.Save(ctx)
	if err != nil {
		if ent.IsValidationError(err) {
			return nil, &echo.HTTPError{
				Message: err.Error(),
				Code:    echo.ErrBadRequest.Code,
			}
		}
		logger.Log.Errorf("failed to create targeting: %v", err)
		return nil, errorz.ErrInternal
	}

	var gender *string
	if createdTargeting.Gender != nil {
		genderStr := createdTargeting.Gender.String()
		gender = &genderStr
	}
	return &dto.Campaign{
		CampaignID:        createdCampaign.ID,
		AdvertiserID:      createdCampaign.AdvertiserID,
		ImpressionsLimit:  createdCampaign.ImpressionsLimit,
		ClicksLimit:       createdCampaign.ClicksLimit,
		CostPerImpression: createdCampaign.CostPerImpression,
		CostPerClick:      createdCampaign.CostPerClick,
		AdTitle:           createdCampaign.AdTitle,
		AdText:            createdCampaign.AdText,
		StartDate:         createdCampaign.StartDate,
		EndDate:           createdCampaign.EndDate,
		Moderated:         createdCampaign.Moderated,
		Targeting: dto.Targeting{
			Gender:   gender,
			AgeFrom:  createdTargeting.AgeFrom,
			AgeTo:    createdTargeting.AgeTo,
			Location: createdTargeting.Location,
		},
	}, nil
}

func (s *campaignService) GetByID(ctx context.Context, campaignID uuid.UUID, advertiserID uuid.UUID) (*dto.Campaign, error) {
	camp, err := s.db.Campaign.Query().
		Where(
			campaign.And(
				campaign.ID(campaignID),
				campaign.AdvertiserID(advertiserID),
			),
		).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errorz.ErrNotFound
		}
		logger.Log.Errorf("failed to get campaign: %v", err)
		return nil, errorz.ErrInternal
	}

	target, err := camp.QueryTargeting().Only(ctx)
	if err != nil {
		logger.Log.Errorf("failed to get targeting: %v", err)
		return nil, errorz.ErrInternal
	}

	var gender *string
	if target.Gender != nil {
		genderStr := target.Gender.String()
		gender = &genderStr
	}

	return &dto.Campaign{
		CampaignID:        camp.ID,
		AdvertiserID:      camp.AdvertiserID,
		ImpressionsLimit:  camp.ImpressionsLimit,
		ClicksLimit:       camp.ClicksLimit,
		CostPerImpression: camp.CostPerImpression,
		CostPerClick:      camp.CostPerClick,
		AdTitle:           camp.AdTitle,
		AdText:            camp.AdText,
		ImageURL:          camp.ImageURL,
		StartDate:         camp.StartDate,
		EndDate:           camp.EndDate,
		Moderated:         camp.Moderated,
		Targeting: dto.Targeting{
			Gender:   gender,
			AgeFrom:  target.AgeFrom,
			AgeTo:    target.AgeTo,
			Location: target.Location,
		},
	}, nil
}

func (s *campaignService) Get(ctx context.Context, advertiserID uuid.UUID, size, page int) ([]*dto.Campaign, error) {
	campaigns, err := s.db.Campaign.Query().
		Where(campaign.AdvertiserID(advertiserID)).
		Offset((page - 1) * size).
		Limit(size).
		Order(ent.Desc(campaign.FieldStartDate)).
		All(ctx)
	if err != nil {
		logger.Log.Errorf("failed to get campaigns: %v", err)
		return nil, errorz.ErrInternal
	}

	var result []*dto.Campaign
	for _, camp := range campaigns {
		target, errQuery := camp.QueryTargeting().Only(ctx)
		if errQuery != nil {
			logger.Log.Errorf("failed to get targeting: %v", errQuery)
			return nil, errorz.ErrInternal
		}

		var gender *string
		if target.Gender != nil {
			genderStr := target.Gender.String()
			gender = &genderStr
		}

		result = append(result, &dto.Campaign{
			CampaignID:        camp.ID,
			AdvertiserID:      camp.AdvertiserID,
			ImpressionsLimit:  camp.ImpressionsLimit,
			ClicksLimit:       camp.ClicksLimit,
			CostPerImpression: camp.CostPerImpression,
			CostPerClick:      camp.CostPerClick,
			AdTitle:           camp.AdTitle,
			AdText:            camp.AdText,
			ImageURL:          camp.ImageURL,
			StartDate:         camp.StartDate,
			EndDate:           camp.EndDate,
			Moderated:         camp.Moderated,
			Targeting: dto.Targeting{
				Gender:   gender,
				AgeFrom:  target.AgeFrom,
				AgeTo:    target.AgeTo,
				Location: target.Location,
			},
		})
	}
	return result, nil
}

func (s *campaignService) Delete(ctx context.Context, campaignID uuid.UUID, advertiserID uuid.UUID) error {
	_, err := s.db.Campaign.Delete().
		Where(
			campaign.And(
				campaign.ID(campaignID),
				campaign.AdvertiserID(advertiserID),
			),
		).Exec(ctx)
	if err != nil {
		logger.Log.Errorf("failed to delete campaign: %v", err)
		return errorz.ErrInternal
	}

	if err := s.clickhouseRepository.DeleteStatsByCampaignID(ctx, campaignID); err != nil {
		logger.Log.Errorf("failed to delete campaign stats: %v", err)
		return errorz.ErrInternal
	}

	if err := s.adImagesRepository.DeleteImage(ctx, campaignID.String()); err != nil {
		logger.Log.Errorf("failed to delete campaign image: %v", err)
		return errorz.ErrInternal
	}

	return nil
}

func (s *campaignService) Update(ctx context.Context, campaignUpdate *dto.CampaignUpdate) (*dto.Campaign, error) {
	camp, err := s.db.Campaign.Query().Where(
		campaign.And(
			campaign.ID(campaignUpdate.CampaignID),
			campaign.AdvertiserID(campaignUpdate.AdvertiserID),
		)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {

			return nil, errorz.ErrNotFound
		}
		logger.Log.Errorf("failed to get campaign: %v", err)
		return nil, errorz.ErrInternal
	}

	currentDate := s.timeService.Now().CurrentDate

	if camp.StartDate <= currentDate {
		switch {
		case campaignUpdate.ImpressionsLimit != camp.ImpressionsLimit:
			return nil, &echo.HTTPError{
				Message: "cannot update impressions limit after campaign start",
				Code:    409,
			}
		case campaignUpdate.ClicksLimit != camp.ClicksLimit:
			return nil, &echo.HTTPError{
				Message: "cannot update clicks limit after campaign start",
				Code:    409,
			}
		case campaignUpdate.StartDate != camp.StartDate:
			return nil, &echo.HTTPError{
				Message: "cannot update start date after campaign start",
				Code:    409,
			}
		case campaignUpdate.EndDate != camp.EndDate:
			return nil, &echo.HTTPError{
				Message: "cannot update end date after campaign start",
				Code:    409,
			}
		}
	}

	if campaignUpdate.StartDate < s.timeService.Now().CurrentDate {
		return nil, &echo.HTTPError{
			Message: fmt.Sprintf("start date must be greater than current date (%d)", currentDate),
			Code:    echo.ErrBadRequest.Code,
		}
	}

	if campaignUpdate.EndDate < campaignUpdate.StartDate {
		return nil, &echo.HTTPError{
			Message: "end date must be greater than start date",
			Code:    echo.ErrBadRequest.Code,
		}
	}

	tx, err := s.db.Tx(ctx)
	if err != nil {
		logger.Log.Errorf("failed to start transaction: %v", err)
		return nil, errorz.ErrInternal
	}

	target, err := camp.QueryTargeting().Only(ctx)
	if err != nil {
		_ = tx.Rollback()
		logger.Log.Errorf("failed to get targeting: %v", err)
		return nil, errorz.ErrInternal
	}

	targetQuery := target.Update()
	if campaignUpdate.Targeting != nil {
		targetQuery = targetQuery.
			SetNillableAgeFrom(campaignUpdate.Targeting.AgeFrom).
			SetNillableAgeTo(campaignUpdate.Targeting.AgeTo).
			SetNillableLocation(campaignUpdate.Targeting.Location)
		if campaignUpdate.Targeting.Gender != nil {
			targetQuery = targetQuery.SetGender(targeting.Gender(*campaignUpdate.Targeting.Gender))
		}
	} else {
		targetQuery = targetQuery.
			SetGender(targeting.GenderALL).
			ClearAgeFrom().
			ClearAgeTo().
			ClearLocation()
	}

	updatedTarget, err := targetQuery.Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		logger.Log.Errorf("failed to update targeting: %v", err)
		return nil, errorz.ErrInternal
	}

	_, err = tx.Campaign.Update().
		Where(
			campaign.And(
				campaign.ID(camp.ID),
				campaign.AdvertiserID(camp.AdvertiserID),
			),
		).
		SetImpressionsLimit(campaignUpdate.ImpressionsLimit).
		SetClicksLimit(campaignUpdate.ClicksLimit).
		SetCostPerImpression(campaignUpdate.CostPerImpression).
		SetCostPerClick(campaignUpdate.CostPerClick).
		SetAdTitle(campaignUpdate.AdTitle).
		SetAdText(campaignUpdate.AdText).
		SetStartDate(campaignUpdate.StartDate).
		SetEndDate(campaignUpdate.EndDate).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		logger.Log.Errorf("failed to update campaign: %v", err)
		return nil, errorz.ErrInternal
	}

	err = tx.Commit()
	if err != nil {
		logger.Log.Errorf("failed to commit transaction: %v", err)
		return nil, errorz.ErrInternal
	}

	var gender *string
	if updatedTarget.Gender != nil {
		genderStr := updatedTarget.Gender.String()
		gender = &genderStr
	}

	return &dto.Campaign{
		CampaignID:        campaignUpdate.CampaignID,
		AdvertiserID:      campaignUpdate.AdvertiserID,
		ImpressionsLimit:  campaignUpdate.ImpressionsLimit,
		ClicksLimit:       campaignUpdate.ClicksLimit,
		CostPerImpression: campaignUpdate.CostPerImpression,
		CostPerClick:      campaignUpdate.CostPerClick,
		AdTitle:           campaignUpdate.AdTitle,
		AdText:            campaignUpdate.AdText,
		ImageURL:          camp.ImageURL,
		StartDate:         campaignUpdate.StartDate,
		EndDate:           campaignUpdate.EndDate,
		Moderated:         camp.Moderated,
		Targeting: dto.Targeting{
			Gender:   gender,
			AgeFrom:  updatedTarget.AgeFrom,
			AgeTo:    updatedTarget.AgeTo,
			Location: updatedTarget.Location,
		},
	}, nil
}

func (s *campaignService) UploadImage(ctx context.Context, uploadImageRequest *dto.CampaignUploadImageRequest, imageData io.Reader) (*dto.CampaignImageURL, error) {
	camp, err := s.db.Campaign.Query().Where(
		campaign.And(
			campaign.ID(uploadImageRequest.CampaignID),
			campaign.AdvertiserID(uploadImageRequest.AdvertiserID),
		)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errorz.ErrNotFound
		}
		logger.Log.Errorf("failed to get campaign: %v", err)
		return nil, errorz.ErrInternal
	}

	imageURL, err := s.adImagesRepository.UploadImage(ctx, camp.ID.String(), imageData)
	if err != nil {
		if errors.Is(minio.ErrFileNotImage, err) {
			return nil, &echo.HTTPError{
				Message: "uploaded file is not image",
				Code:    echo.ErrBadRequest.Code,
			}
		}
		logger.Log.Errorf("failed to upload image: %v", err)
		return nil, errorz.ErrInternal
	}

	_, err = s.db.Campaign.Update().
		Where(campaign.ID(camp.ID)).
		SetImageURL(imageURL).
		Save(ctx)
	if err != nil {
		logger.Log.Errorf("failed to update campaign: %v", err)
		return nil, errorz.ErrInternal
	}

	return &dto.CampaignImageURL{
		AdvertiserID: camp.AdvertiserID,
		CampaignID:   camp.ID,
		ImageURL:     imageURL,
	}, nil
}

func (s *campaignService) RemoveImage(ctx context.Context, removeImageRequest *dto.CampaignRemoveImageRequest) error {
	if err := s.adImagesRepository.DeleteImage(ctx, removeImageRequest.CampaignID.String()); err != nil {
		if errors.Is(minio.ErrImageNotFound, err) {
			return errorz.ErrNotFound
		}
		logger.Log.Errorf("failed to delete image: %v", err)
		return errorz.ErrInternal
	}

	_, err := s.db.Campaign.Update().
		Where(campaign.ID(removeImageRequest.CampaignID)).
		ClearImageURL().
		Save(ctx)
	if err != nil {
		logger.Log.Errorf("failed to update campaign: %v", err)
		return errorz.ErrInternal
	}

	return nil
}
