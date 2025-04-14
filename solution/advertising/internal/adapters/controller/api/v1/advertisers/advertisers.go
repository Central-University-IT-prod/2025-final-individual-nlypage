package advertisers

import (
	"context"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	v1 "nlypage-final/internal/adapters/controller/api/v1"
	"nlypage-final/internal/adapters/controller/api/validator"
	"nlypage-final/internal/domain/dto"
)

type advertiserService interface {
	GetByID(ctx context.Context, advertiserID uuid.UUID) (*dto.Advertiser, error)
	UpsertBulk(ctx context.Context, upsertAdvertisers []dto.AdvertiserUpsert) error
}

type advertisersHandler struct {
	advertiserService advertiserService
	validator         *validator.Validator
}

func NewAdvertisersHandler(advertiserService advertiserService, validator *validator.Validator) v1.Handler {
	return &advertisersHandler{
		advertiserService: advertiserService,
		validator:         validator,
	}
}

func (h advertisersHandler) upsertBulk(c echo.Context) error {
	var advertisers []dto.AdvertiserUpsert
	if err := c.Bind(&advertisers); err != nil {
		return err
	}

	if err := h.validator.ValidateData(advertisers); err != nil {
		return err
	}

	err := h.advertiserService.UpsertBulk(c.Request().Context(), advertisers)
	if err != nil {
		return &echo.HTTPError{
			Message: err.Error(),
			Code:    echo.ErrInternalServerError.Code,
		}
	}

	return c.JSON(201, advertisers)
}

func (h advertisersHandler) GetByID(c echo.Context) error {
	var advertiserID dto.AdvertiserGet
	if err := c.Bind(&advertiserID); err != nil {
		return err
	}

	if err := h.validator.ValidateData(advertiserID); err != nil {
		return err
	}

	client, err := h.advertiserService.GetByID(c.Request().Context(), advertiserID.AdvertiserID)
	if err != nil {
		return err
	}

	return c.JSON(200, client)
}

func (h advertisersHandler) Setup(group *echo.Group) {
	group.GET("/:advertiserId", h.GetByID)
	group.POST("/bulk", h.upsertBulk)
}
