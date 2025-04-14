package ai

import (
	"context"
	"github.com/labstack/echo/v4"
	v1 "nlypage-final/internal/adapters/controller/api/v1"
	"nlypage-final/internal/adapters/controller/api/validator"
	"nlypage-final/internal/domain/dto"
)

type generateService interface {
	GenerateAdText(ctx context.Context, generateAdText *dto.GenerateAdTextRequest) (*dto.GenerateAdTextResponse, error)
}

type aiHandler struct {
	generateService generateService
	validator       *validator.Validator
}

func NewAiHandler(generateService generateService, validator *validator.Validator) v1.Handler {
	return &aiHandler{
		generateService: generateService,
		validator:       validator,
	}
}

func (h aiHandler) generateAdText(c echo.Context) error {
	var generateAdTextRequest dto.GenerateAdTextRequest
	if err := c.Bind(&generateAdTextRequest); err != nil {
		return err
	}

	if err := h.validator.ValidateData(&generateAdTextRequest); err != nil {
		return err
	}

	generateAdTextResponse, err := h.generateService.GenerateAdText(c.Request().Context(), &generateAdTextRequest)
	if err != nil {
		return err
	}

	return c.JSON(200, generateAdTextResponse)
}

func (h aiHandler) Setup(group *echo.Group) {
	group.POST("/advertisers/:advertiserId/generate/ad-text", h.generateAdText)
}
