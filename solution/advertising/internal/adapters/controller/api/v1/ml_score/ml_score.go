package ml_score

import (
	"context"
	"github.com/labstack/echo/v4"
	v1 "nlypage-final/internal/adapters/controller/api/v1"
	"nlypage-final/internal/adapters/controller/api/validator"
	"nlypage-final/internal/domain/dto"
)

type mlScoreService interface {
	Upsert(ctx context.Context, upsertMlScore dto.MlScoreUpsert) error
}

type mlScoreHandler struct {
	mlScoreService mlScoreService
	validator      *validator.Validator
}

func NewMlScoreHandler(mlScoreService mlScoreService, validator *validator.Validator) v1.Handler {
	return &mlScoreHandler{
		mlScoreService: mlScoreService,
		validator:      validator,
	}
}

func (h mlScoreHandler) upsertBulk(c echo.Context) error {
	var upsertMlScores dto.MlScoreUpsert
	if err := c.Bind(&upsertMlScores); err != nil {
		return err
	}
	if err := h.validator.ValidateData(upsertMlScores); err != nil {
		return err
	}
	if err := h.mlScoreService.Upsert(c.Request().Context(), upsertMlScores); err != nil {
		return err
	}
	return c.JSON(200, upsertMlScores)
}

func (h mlScoreHandler) Setup(group *echo.Group) {
	group.POST("", h.upsertBulk)
}
