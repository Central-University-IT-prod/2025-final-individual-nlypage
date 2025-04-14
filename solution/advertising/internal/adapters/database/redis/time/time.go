package time

import (
	"context"
	"github.com/go-redis/redis/v8"
)

type Storage interface {
	Now() (int, error)
	Set(day int)
	Close() error
}

type storage struct {
	redis *redis.Client
}

func NewStorage(client *redis.Client) Storage {
	return &storage{redis: client}
}

func (s *storage) Now() (int, error) {
	return s.redis.Get(context.Background(), "current-day").Int()
}

func (s *storage) Set(day int) {
	s.redis.Set(context.Background(), "current-day", day, 0)
}

func (s *storage) Close() error {
	return s.redis.Close()
}
