package campaigns

import (
	"context"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"io"
	v1 "nlypage-final/internal/adapters/controller/api/v1"
	"nlypage-final/internal/adapters/controller/api/validator"
	"nlypage-final/internal/domain/dto"
)

type campaignService interface {
	Create(ctx context.Context, campaign *dto.CampaignCreate) (*dto.Campaign, error)
	GetByID(ctx context.Context, campaignID uuid.UUID, advertiserID uuid.UUID) (*dto.Campaign, error)
	Get(ctx context.Context, advertiserID uuid.UUID, size, page int) ([]*dto.Campaign, error)
	Delete(ctx context.Context, campaignID uuid.UUID, advertiserID uuid.UUID) error
	Update(ctx context.Context, campaignUpdate *dto.CampaignUpdate) (*dto.Campaign, error)
	UploadImage(ctx context.Context, uploadImageRequest *dto.CampaignUploadImageRequest, imageData io.Reader) (*dto.CampaignImageURL, error)
	RemoveImage(ctx context.Context, removeImageRequest *dto.CampaignRemoveImageRequest) error
}

type campaignsHandler struct {
	service   campaignService
	validator *validator.Validator
}

func NewCampaignsHandler(service campaignService, validator *validator.Validator) v1.Handler {
	return &campaignsHandler{
		service:   service,
		validator: validator,
	}
}

func (h campaignsHandler) create(c echo.Context) error {
	var campaign dto.CampaignCreate
	if err := c.Bind(&campaign); err != nil {
		return err
	}

	if err := h.validator.ValidateData(&campaign); err != nil {
		return err
	}

	createdCampaign, err := h.service.Create(c.Request().Context(), &campaign)
	if err != nil {
		return err
	}

	return c.JSON(201, createdCampaign)
}

func (h campaignsHandler) get(c echo.Context) error {
	var campaignsGet dto.CampaignsGet
	if err := c.Bind(&campaignsGet); err != nil {
		return err
	}
	switch {
	case campaignsGet.Size == 0:
		campaignsGet.Size = 10
	case campaignsGet.Page == 0:
		campaignsGet.Page = 1
	}

	if err := h.validator.ValidateData(campaignsGet); err != nil {
		return err
	}

	campaigns, err := h.service.Get(
		c.Request().Context(),
		campaignsGet.AdvertiserID,
		campaignsGet.Size,
		campaignsGet.Page,
	)
	if err != nil {
		return err
	}

	if len(campaigns) == 0 {
		return c.JSON(200, []dto.Campaign{})
	}

	return c.JSON(200, campaigns)
}

func (h campaignsHandler) getByID(c echo.Context) error {
	var campaignGet dto.CampaignGet
	if err := c.Bind(&campaignGet); err != nil {
		return err
	}
	if err := h.validator.ValidateData(campaignGet); err != nil {
		return err
	}

	campaign, err := h.service.GetByID(
		c.Request().Context(),
		campaignGet.CampaignID,
		campaignGet.AdvertiserID,
	)
	if err != nil {
		return err
	}

	return c.JSON(200, campaign)
}

func (h campaignsHandler) delete(c echo.Context) error {
	var campaignDelete dto.CampaignDelete
	if err := c.Bind(&campaignDelete); err != nil {
		return err
	}
	if err := h.validator.ValidateData(campaignDelete); err != nil {
		return err
	}

	if err := h.service.Delete(
		c.Request().Context(),
		campaignDelete.CampaignID,
		campaignDelete.AdvertiserID,
	); err != nil {
		return err
	}

	return c.NoContent(204)
}

func (h campaignsHandler) update(c echo.Context) error {
	var campaignUpdate dto.CampaignUpdate
	if err := c.Bind(&campaignUpdate); err != nil {
		return err
	}
	if err := h.validator.ValidateData(campaignUpdate); err != nil {
		return err
	}

	updatedCampaign, err := h.service.Update(c.Request().Context(), &campaignUpdate)
	if err != nil {
		return err
	}

	return c.JSON(200, updatedCampaign)
}

func (h campaignsHandler) uploadImage(c echo.Context) error {
	var uploadImageRequest dto.CampaignUploadImageRequest
	if err := c.Bind(&uploadImageRequest); err != nil {
		return err
	}
	if err := h.validator.ValidateData(uploadImageRequest); err != nil {
		return err
	}

	image, err := c.FormFile("image")
	if err != nil {
		return &echo.HTTPError{
			Message: "failed to get image",
			Code:    echo.ErrBadRequest.Code,
		}
	}

	imageData, err := image.Open()
	if err != nil {
		return &echo.HTTPError{
			Message: "failed to open image",
			Code:    echo.ErrInternalServerError.Code,
		}
	}
	defer imageData.Close()

	imageURL, err := h.service.UploadImage(c.Request().Context(), &uploadImageRequest, imageData)
	if err != nil {
		return err
	}

	return c.JSON(200, imageURL)
}

func (h campaignsHandler) removeImage(c echo.Context) error {
	var removeImageRequest dto.CampaignRemoveImageRequest
	if err := c.Bind(&removeImageRequest); err != nil {
		return err
	}
	if err := h.validator.ValidateData(removeImageRequest); err != nil {
		return err
	}

	if err := h.service.RemoveImage(c.Request().Context(), &removeImageRequest); err != nil {
		return err
	}

	return c.NoContent(204)
}

func (h campaignsHandler) Setup(group *echo.Group) {
	group.POST("/:advertiserId/campaigns", h.create)
	group.GET("/:advertiserId/campaigns", h.get)
	group.GET("/:advertiserId/campaigns/:campaignId", h.getByID)
	group.DELETE("/:advertiserId/campaigns/:campaignId", h.delete)
	group.PUT("/:advertiserId/campaigns/:campaignId", h.update)
	group.POST("/:advertiserId/campaigns/:campaignId/image", h.uploadImage)
	group.DELETE("/:advertiserId/campaigns/:campaignId/image", h.removeImage)
}
