package clients

import (
	"context"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	v1 "nlypage-final/internal/adapters/controller/api/v1"
	"nlypage-final/internal/adapters/controller/api/validator"
	"nlypage-final/internal/domain/dto"
)

type clientService interface {
	UpsertBulk(ctx context.Context, upsertClients []dto.ClientUpsert) error
	GetByID(ctx context.Context, clientID uuid.UUID) (*dto.Client, error)
}

type clientsHandler struct {
	clientService clientService
	validator     *validator.Validator
}

func NewClientsHandler(clientService clientService, validator *validator.Validator) v1.Handler {
	return &clientsHandler{
		clientService: clientService,
		validator:     validator,
	}
}

func (h clientsHandler) upsertBulk(c echo.Context) error {
	var clients []dto.ClientUpsert
	if err := c.Bind(&clients); err != nil {
		return err
	}

	if err := h.validator.ValidateData(clients); err != nil {
		return err
	}

	err := h.clientService.UpsertBulk(c.Request().Context(), clients)
	if err != nil {
		return &echo.HTTPError{
			Message: err.Error(),
			Code:    echo.ErrInternalServerError.Code,
		}
	}

	return c.JSON(201, clients)
}

func (h clientsHandler) GetByID(c echo.Context) error {
	var clientID dto.ClientGet
	if err := c.Bind(&clientID); err != nil {
		return err
	}

	if err := h.validator.ValidateData(clientID); err != nil {
		return err
	}

	client, err := h.clientService.GetByID(c.Request().Context(), clientID.ClientID)
	if err != nil {
		return err
	}

	return c.JSON(200, client)
}

func (h clientsHandler) Setup(group *echo.Group) {
	group.GET("/:clientId", h.GetByID)
	group.POST("/bulk", h.upsertBulk)
}
