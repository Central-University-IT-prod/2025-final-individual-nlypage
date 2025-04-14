package service

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"nlypage-final/internal/domain/dto"
	"nlypage-final/internal/domain/utils"
	"nlypage-final/pkg/gigachat"
)

type generateAdvertiserService interface {
	GetByID(ctx context.Context, advertiserID uuid.UUID) (*dto.Advertiser, error)
}

type GenerateService interface {
	GenerateAdText(ctx context.Context, generateAdText *dto.GenerateAdTextRequest) (*dto.GenerateAdTextResponse, error)
}

type generateService struct {
	gigachatClient    *gigachat.Client
	advertiserService generateAdvertiserService
}

func NewGenerateService(gigachatClient *gigachat.Client, advertiserService generateAdvertiserService) GenerateService {
	return &generateService{
		gigachatClient:    gigachatClient,
		advertiserService: advertiserService,
	}
}

func (s *generateService) GenerateAdText(ctx context.Context, generateAdText *dto.GenerateAdTextRequest) (*dto.GenerateAdTextResponse, error) {
	advertiser, err := s.advertiserService.GetByID(ctx, generateAdText.AdvertiserID)
	if err != nil {
		return nil, &echo.HTTPError{
			Message: fmt.Errorf("advertiser not found"),
			Code:    echo.ErrBadRequest.Code,
		}
	}

	prompt :=
		`Создайте привлекательный рекламный текст для следующего объявления:
Рекламодатель: %s
Заголовок объявления: %s
%s
				
Требования к тексту:
1. Текст должен быть убедительным и привлекательным и при этом коротким
2. Язык текста: %s
3. Текст должен быть оптимизирован для целевой аудитории
4. Не добавляй форматирование текста
				
Пожалуйста, создайте рекламный текст, который соответствует этим требованиям.`

	additionalInfo := ""
	if generateAdText.AdditionalInfo != "" {
		additionalInfo = "Дополнительная информация:\n" + generateAdText.AdditionalInfo
	}

	formattedPrompt := fmt.Sprintf(prompt, advertiser.Name, generateAdText.AdTitle, additionalInfo, utils.DetectLanguage(generateAdText.AdTitle))

	errAuth := s.gigachatClient.AuthWithContext(ctx)
	if errAuth != nil {
		return nil, &echo.HTTPError{
			Message: fmt.Errorf("failed to generate ad text: %w", errAuth).Error(),
			Code:    echo.ErrUnauthorized.Code,
		}
	}
	response, err := s.gigachatClient.ChatWithContext(ctx, &gigachat.ChatRequest{
		Model: "GigaChat",
		Messages: []gigachat.Message{
			{
				Role:    gigachat.UserRole,
				Content: formattedPrompt,
			},
		},
	})

	if err != nil {
		return nil, &echo.HTTPError{
			Message: fmt.Errorf("failed to generate ad text: %w", err).Error(),
			Code:    echo.ErrInternalServerError.Code,
		}
	}

	return &dto.GenerateAdTextResponse{
		AdText: response.Choices[0].Message.Content,
	}, nil
}
