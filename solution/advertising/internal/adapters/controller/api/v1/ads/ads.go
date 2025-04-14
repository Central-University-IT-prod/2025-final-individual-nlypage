package ads

import (
	"context"
	"github.com/labstack/echo/v4"
	v1 "nlypage-final/internal/adapters/controller/api/v1"
	"nlypage-final/internal/adapters/controller/api/validator"
	"nlypage-final/internal/domain/dto"
)

type adService interface {
	SelectAd(ctx context.Context, clientID dto.ClientAdGet) (*dto.Ad, error)
	RecordClick(ctx context.Context, click dto.ClientAdClick) error
}

type adsHandler struct {
	adService adService
	validator *validator.Validator
}

func NewAdsHandler(adService adService, validator *validator.Validator) v1.Handler {
	return &adsHandler{
		adService: adService,
		validator: validator,
	}
}

func (a adsHandler) getAd(c echo.Context) error {
	var request dto.ClientAdGet
	if err := c.Bind(&request); err != nil {
		return err
	}
	if err := a.validator.ValidateData(request); err != nil {
		return err
	}

	ad, err := a.adService.SelectAd(c.Request().Context(), request)
	if err != nil {
		return err
	}

	return c.JSON(200, ad)
}

func (a adsHandler) clickAd(c echo.Context) error {
	var request dto.ClientAdClick
	if err := c.Bind(&request); err != nil {
		return err
	}
	if err := a.validator.ValidateData(request); err != nil {
		return err
	}

	if err := a.adService.RecordClick(c.Request().Context(), request); err != nil {
		return err
	}

	return c.NoContent(204)
}

func (a adsHandler) Setup(group *echo.Group) {
	group.GET("", a.getAd)
	group.POST("/:adID/click", a.clickAd)
}
