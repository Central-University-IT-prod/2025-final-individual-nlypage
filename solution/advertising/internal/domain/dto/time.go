package dto

type CurrentDate struct {
	CurrentDate int `json:"current_date" validate:"gte=0"`
}
