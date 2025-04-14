package stats

import (
	"context"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	v1 "nlypage-final/internal/adapters/controller/api/v1"
	"nlypage-final/internal/adapters/controller/api/validator"
	"nlypage-final/internal/domain/dto"
)

type statsService interface {
	Campaign(ctx context.Context, campaignID uuid.UUID) (*dto.Stats, error)
	CampaignDaily(ctx context.Context, campaignID uuid.UUID) ([]*dto.StatsDaily, error)
	Advertiser(ctx context.Context, advertiserID uuid.UUID) (*dto.Stats, error)
	AdvertiserDaily(ctx context.Context, advertiserID uuid.UUID) ([]*dto.StatsDaily, error)
}

type statsHandler struct {
	statsService statsService
	validator    *validator.Validator
}

func NewStatsHandler(statsService statsService, validator *validator.Validator) v1.Handler {
	return &statsHandler{
		statsService: statsService,
		validator:    validator,
	}
}

func (h statsHandler) campaign(c echo.Context) error {
	var campaignStatsGet dto.CampaignStatsGet
	if err := c.Bind(&campaignStatsGet); err != nil {
		return err
	}
	if err := h.validator.ValidateData(campaignStatsGet); err != nil {
		return err
	}

	stats, err := h.statsService.Campaign(c.Request().Context(), campaignStatsGet.CampaignID)
	if err != nil {
		return err
	}

	return c.JSON(200, stats)
}

func (h statsHandler) campaignDaily(c echo.Context) error {
	var campaignDailyStatsGet dto.CampaignDailyStatsGet
	if err := c.Bind(&campaignDailyStatsGet); err != nil {
		return err
	}
	if err := h.validator.ValidateData(campaignDailyStatsGet); err != nil {
		return err
	}

	stats, err := h.statsService.CampaignDaily(c.Request().Context(), campaignDailyStatsGet.CampaignID)
	if err != nil {
		return err
	}

	if len(stats) == 0 {
		return c.JSON(200, []dto.StatsDaily{})
	}

	return c.JSON(200, stats)
}

func (h statsHandler) advertiser(c echo.Context) error {
	var advertiserStatsGet dto.AdvertiserStatsGet
	if err := c.Bind(&advertiserStatsGet); err != nil {
		return err
	}
	if err := h.validator.ValidateData(advertiserStatsGet); err != nil {
		return err
	}

	stats, err := h.statsService.Advertiser(c.Request().Context(), advertiserStatsGet.AdvertiserID)
	if err != nil {
		return err
	}

	return c.JSON(200, stats)
}

func (h statsHandler) advertiserDaily(c echo.Context) error {
	var advertiserDailyStatsGet dto.AdvertiserDailyStatsGet
	if err := c.Bind(&advertiserDailyStatsGet); err != nil {
		return err
	}
	if err := h.validator.ValidateData(advertiserDailyStatsGet); err != nil {
		return err
	}

	stats, err := h.statsService.AdvertiserDaily(c.Request().Context(), advertiserDailyStatsGet.AdvertiserID)
	if err != nil {
		return err
	}

	if len(stats) == 0 {
		return c.JSON(200, []dto.StatsDaily{})
	}

	return c.JSON(200, stats)
}

func (h statsHandler) Setup(group *echo.Group) {
	group.GET("/campaigns/:campaignId", h.campaign)
	group.GET("/campaigns/:campaignId/daily", h.campaignDaily)
	group.GET("/advertisers/:advertiserId/campaigns", h.advertiser)
	group.GET("/advertisers/:advertiserId/campaigns/daily", h.advertiserDaily)
}
