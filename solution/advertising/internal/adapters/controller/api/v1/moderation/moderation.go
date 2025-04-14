package moderation

import (
	"context"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	v1 "nlypage-final/internal/adapters/controller/api/v1"
	"nlypage-final/internal/adapters/controller/api/validator"
	"nlypage-final/internal/domain/dto"
)

type moderationService interface {
	GetNotModeratedCampaigns(ctx context.Context) ([]*dto.Campaign, error)
	ApproveCampaign(ctx context.Context, campaignID uuid.UUID) error
}

type moderationHandler struct {
	service   moderationService
	validator *validator.Validator
}

func NewModerationHandler(service moderationService, validator *validator.Validator) v1.Handler {
	return &moderationHandler{
		service:   service,
		validator: validator,
	}
}

func (h moderationHandler) list(c echo.Context) error {
	campaigns, err := h.service.GetNotModeratedCampaigns(c.Request().Context())
	if err != nil {
		return err
	}

	if len(campaigns) == 0 {
		return c.JSON(200, []dto.Campaign{})
	}
	return c.JSON(200, campaigns)
}

func (h moderationHandler) approve(c echo.Context) error {
	var campaignApprove dto.CampaignApprove
	if err := c.Bind(&campaignApprove); err != nil {
		return err
	}
	if err := h.validator.ValidateData(campaignApprove); err != nil {
		return err
	}

	if err := h.service.ApproveCampaign(c.Request().Context(), campaignApprove.CampaignID); err != nil {
		return err
	}

	return c.NoContent(204)
}

func (h moderationHandler) Setup(group *echo.Group) {
	group.GET("/campaigns", h.list)
	group.POST("/campaigns/:campaignId/approve", h.approve)
}
