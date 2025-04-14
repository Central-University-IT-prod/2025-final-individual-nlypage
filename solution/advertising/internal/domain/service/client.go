package service

import (
	"context"
	"nlypage-final/internal/adapters/database/postgres/ent"
	"nlypage-final/internal/adapters/database/postgres/ent/user"
	"nlypage-final/internal/domain/common/errorz"
	"nlypage-final/internal/domain/dto"

	"github.com/google/uuid"
)

type ClientService interface {
	GetByID(ctx context.Context, clientID uuid.UUID) (*dto.Client, error)
	UpsertBulk(ctx context.Context, upsertClients []dto.ClientUpsert) error
}

type clientService struct {
	db *ent.Client
}

func NewClientService(db *ent.Client) ClientService {
	return &clientService{
		db: db,
	}
}

func (s *clientService) GetByID(ctx context.Context, clientID uuid.UUID) (*dto.Client, error) {
	client, err := s.db.User.Get(ctx, clientID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errorz.ErrNotFound
		}
		return nil, errorz.ErrInternal
	}

	return &dto.Client{
		ClientID: client.ID,
		Login:    client.Login,
		Age:      client.Age,
		Location: client.Location,
		Gender:   string(client.Gender),
	}, nil
}

func (s *clientService) UpsertBulk(ctx context.Context, upsertClients []dto.ClientUpsert) error {
	err := s.db.User.MapCreateBulk(upsertClients, func(c *ent.UserCreate, i int) {
		c.SetID(upsertClients[i].ClientID).
			SetLogin(upsertClients[i].Login).
			SetAge(upsertClients[i].Age).
			SetLocation(upsertClients[i].Location).
			SetGender(user.Gender(upsertClients[i].Gender))
	}).OnConflictColumns(user.FieldID).UpdateNewValues().Exec(ctx)
	if err != nil {
		return errorz.ErrInternal
	}
	return nil
}
