package states

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type Storage interface {
	Get(userID int64) (string, error)
	Set(userID int64, state string, expiration time.Duration) error
	Delete(userID int64)
	Close() error
}

type storage struct {
	redis *redis.Client
}

func NewStorage(client *redis.Client) Storage {
	return &storage{
		redis: client,
	}
}

func (s *storage) Get(userID int64) (string, error) {
	data, err := s.redis.Get(context.Background(), fmt.Sprintf("%d", userID)).Result()
	if err != nil {
		return "", err
	}
	return data, nil
}

func (s *storage) Set(userID int64, state string, expiration time.Duration) error {
	return s.redis.Set(context.Background(), fmt.Sprintf("%d", userID), state, expiration).Err()
}

func (s *storage) Delete(userID int64) {
	s.redis.Del(context.Background(), fmt.Sprintf("%d", userID))
}

func (s *storage) Close() error {
	return s.redis.Close()
}
