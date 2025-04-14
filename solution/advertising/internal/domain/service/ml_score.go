package service

import (
	"context"
	"github.com/labstack/echo/v4"
	"nlypage-final/internal/adapters/database/postgres/ent"
	"nlypage-final/internal/adapters/database/postgres/ent/mlscore"
	"nlypage-final/internal/domain/common/errorz"
	"nlypage-final/internal/domain/dto"
)

type MlScoreService interface {
	Upsert(ctx context.Context, upsertMlScore dto.MlScoreUpsert) error
}

type mlScoreService struct {
	db *ent.Client
}

func NewMlScoreService(db *ent.Client) MlScoreService {
	return &mlScoreService{
		db: db,
	}
}

func (s *mlScoreService) Upsert(ctx context.Context, upsertMlScore dto.MlScoreUpsert) error {
	err := s.db.MlScore.
		Create().
		SetAdvertiserID(upsertMlScore.AdvertiserID).
		SetUserID(upsertMlScore.ClientID).
		SetScore(upsertMlScore.Score).
		OnConflictColumns(mlscore.FieldAdvertiserID, mlscore.FieldUserID).UpdateNewValues().Exec(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			return &echo.HTTPError{
				Message: "constraint error, advertiser or client not found",
				Code:    echo.ErrNotFound.Code,
			}
		}
		return errorz.ErrInternal
	}
	return nil
}
