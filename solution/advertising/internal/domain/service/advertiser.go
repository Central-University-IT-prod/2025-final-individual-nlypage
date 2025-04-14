package service

import (
	"context"
	"github.com/google/uuid"
	"nlypage-final/internal/adapters/database/postgres/ent"
	"nlypage-final/internal/adapters/database/postgres/ent/user"
	"nlypage-final/internal/domain/common/errorz"
	"nlypage-final/internal/domain/dto"
)

type AdvertiserService interface {
	GetByID(ctx context.Context, advertiserID uuid.UUID) (*dto.Advertiser, error)
	UpsertBulk(ctx context.Context, upsertAdvertisers []dto.AdvertiserUpsert) error
}

type advertiserService struct {
	db *ent.Client
}

func NewAdvertiserService(db *ent.Client) AdvertiserService {
	return &advertiserService{
		db: db,
	}
}

func (s *advertiserService) GetByID(ctx context.Context, advertiserID uuid.UUID) (*dto.Advertiser, error) {
	advertiser, err := s.db.Advertiser.Get(ctx, advertiserID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errorz.ErrNotFound
		}
		return nil, errorz.ErrInternal
	}

	return &dto.Advertiser{
		AdvertiserID: advertiser.ID,
		Name:         advertiser.Name,
	}, nil
}

func (s *advertiserService) UpsertBulk(ctx context.Context, upsertAdvertisers []dto.AdvertiserUpsert) error {
	err := s.db.Advertiser.MapCreateBulk(upsertAdvertisers, func(c *ent.AdvertiserCreate, i int) {
		c.SetID(upsertAdvertisers[i].AdvertiserID).
			SetName(upsertAdvertisers[i].Name)
	}).OnConflictColumns(user.FieldID).UpdateNewValues().Exec(ctx)
	if err != nil {
		return errorz.ErrInternal
	}
	return nil
}
