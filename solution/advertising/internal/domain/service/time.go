package service

import (
	"errors"
	"nlypage-final/internal/domain/dto"
)

type timeStorage interface {
	Now() (int, error)
	Set(day int)
}

type TimeService interface {
	Now() *dto.CurrentDate
	Set(day int) (*dto.CurrentDate, error)
}

type timeService struct {
	day         int
	timeStorage timeStorage
}

func NewTimeService(timeStorage timeStorage) (TimeService, error) {
	currentDay, err := timeStorage.Now()
	if err != nil {
		currentDay = 0
	}
	return &timeService{
		day:         currentDay,
		timeStorage: timeStorage,
	}, nil
}

// Now returns current day
func (s *timeService) Now() *dto.CurrentDate {
	return &dto.CurrentDate{CurrentDate: s.day}
}

func (s *timeService) Set(day int) (*dto.CurrentDate, error) {
	if day <= s.day {
		return nil, errors.New("new day must be greater than current day")
	}
	s.day = day
	s.timeStorage.Set(day)

	return &dto.CurrentDate{CurrentDate: day}, nil
}
