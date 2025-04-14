package time

import (
	"context"
	"github.com/labstack/echo/v4"
	v1 "nlypage-final/internal/adapters/controller/api/v1"
	"nlypage-final/internal/adapters/controller/api/validator"
	"nlypage-final/internal/domain/dto"
)

type timeService interface {
	Now() *dto.CurrentDate
	Set(date int) (*dto.CurrentDate, error)
}

type timeAdScoringService interface {
	ForceUpdate(ctx context.Context) error
}

type timeHandler struct {
	timeService      timeService
	adScoringService timeAdScoringService
	validator        *validator.Validator
}

func NewTimeHandler(timeService timeService, adScoringService timeAdScoringService, validator *validator.Validator) v1.Handler {
	return &timeHandler{
		timeService:      timeService,
		adScoringService: adScoringService,
		validator:        validator,
	}
}

func (h timeHandler) advance(c echo.Context) error {
	var t dto.CurrentDate
	if err := c.Bind(&t); err != nil {
		return err
	}

	if err := h.validator.ValidateData(t); err != nil {
		return err
	}

	date, err := h.timeService.Set(t.CurrentDate)
	if err != nil {
		return err
	}

	//if err := h.adScoringService.ForceUpdate(c.Request().Context()); err != nil {
	//	return err
	//}

	return c.JSON(200, date)
}

func (h timeHandler) now(c echo.Context) error {
	return c.JSON(200, h.timeService.Now())
}

func (h timeHandler) Setup(group *echo.Group) {
	group.POST("/advance", h.advance)
	group.GET("/now", h.now)
}
