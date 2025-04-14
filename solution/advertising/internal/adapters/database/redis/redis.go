package redis

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"nlypage-final/internal/adapters/database/redis/ads"
	"nlypage-final/internal/adapters/database/redis/states"
	"nlypage-final/internal/adapters/database/redis/time"
)

type Client struct {
	Time   time.Storage
	States states.Storage
	Ads    ads.Storage
	Cache  *redis.Client
}

type Options struct {
	Host     string
	Port     string
	Password string
}

func New(opts Options) (*Client, error) {
	timeRedis := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", opts.Host, opts.Port),
		Password: opts.Password,
		DB:       0,
	})
	if err := timeRedis.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping time storage: %w", err)
	}

	statesRedis := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", opts.Host, opts.Port),
		Password: opts.Password,
		DB:       1,
	})
	if err := statesRedis.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping states storage: %w", err)
	}

	adsRedis := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", opts.Host, opts.Port),
		Password: opts.Password,
		DB:       2,
	})
	if err := adsRedis.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping ads storage: %w", err)
	}

	cacheRedis := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", opts.Host, opts.Port),
		Password: opts.Password,
		DB:       3,
	})
	if err := cacheRedis.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping cache storage: %w", err)
	}

	return &Client{
		Time:   time.NewStorage(timeRedis),
		States: states.NewStorage(statesRedis),
		Ads:    ads.NewStorage(adsRedis),
		Cache:  cacheRedis,
	}, nil
}

// CloseAll closes all redis clients
func (c *Client) CloseAll() error {
	_ = c.Time.Close()
	_ = c.States.Close()
	_ = c.Ads.Close()
	return nil
}
