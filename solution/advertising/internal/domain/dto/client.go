package dto

import "github.com/google/uuid"

// NOTE: Большинство DTO сгенерировано ИИ и подредактировано мной

// Client представляет клиента системы
type Client struct {
	ClientID uuid.UUID `json:"client_id" validate:"required"`
	Login    string    `json:"login" validate:"required"`
	Age      int       `json:"age" validate:"required,gte=0"`
	Location string    `json:"location" validate:"required"`
	Gender   string    `json:"gender" validate:"required,oneof=MALE FEMALE"`
}

type ClientID struct {
	ClientID uuid.UUID `json:"client_id" validate:"required"`
}

type ClientGet struct {
	ClientID uuid.UUID `param:"clientId" validate:"required"`
}

// ClientUpsert представляет DTO для создания/обновления клиента
type ClientUpsert struct {
	ClientID uuid.UUID `json:"client_id" validate:"required"`
	Login    string    `json:"login" validate:"required"`
	Age      int       `json:"age" validate:"required,gte=0"`
	Location string    `json:"location" validate:"required"`
	Gender   string    `json:"gender" validate:"required,oneof=MALE FEMALE"`
}
