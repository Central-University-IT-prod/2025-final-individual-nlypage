package ads

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Ad struct {
	UserID            uuid.UUID
	AdID              uuid.UUID       `json:"ad_id"`
	AdTitle           string          `json:"ad_title"`
	AdText            string          `json:"ad_text"`
	ImageURL          string          `json:"image_url"`
	AdvertiserID      uuid.UUID       `json:"advertiser_id"`
	CostPerImpression float64         `json:"cost_per_impression"`
	Score             decimal.Decimal `json:"score"`
}

type Storage interface {
	Add(ctx context.Context, userID uuid.UUID, ad Ad) error
	UpdateAll(ctx context.Context, ads []Ad) error
	Remove(ctx context.Context, userID uuid.UUID, adID uuid.UUID)
	Get(ctx context.Context, userID uuid.UUID) (Ad, error)
	Close() error
	GetMedianScore(ctx context.Context, userID uuid.UUID) (decimal.Decimal, error)
}

type storage struct {
	redis *redis.Client
}

func NewStorage(client *redis.Client) Storage {
	return &storage{redis: client}
}

func (s *storage) Add(ctx context.Context, userID uuid.UUID, ad Ad) error {
	key := fmt.Sprintf("user:%s:ads", userID.String())
	adKey := fmt.Sprintf("ad:%s", ad.AdID.String())

	// Store the complete ad data
	adData, err := json.Marshal(ad)
	if err != nil {
		return fmt.Errorf("failed to marshal ad data: %w", err)
	}

	err = s.redis.Set(ctx, adKey, adData, 0).Err()
	if err != nil {
		return fmt.Errorf("failed to store ad data: %w", err)
	}

	// Store the score in sorted set for ranking
	member := &redis.Z{
		Score:  ad.Score.InexactFloat64(),
		Member: ad.AdID.String(),
	}
	return s.redis.ZAdd(ctx, key, member).Err()
}

func (s *storage) UpdateAll(ctx context.Context, ads []Ad) error {
	if len(ads) == 0 {
		return nil
	}

	// Group ads by UserID
	adsByUser := make(map[uuid.UUID][]Ad)
	oldKeys := make([]string, 0)
	for _, ad := range ads {
		adsByUser[ad.UserID] = append(adsByUser[ad.UserID], ad)
		oldKeys = append(oldKeys, fmt.Sprintf("user:%s:ads", ad.UserID.String()))
	}

	// Start a Redis pipeline for adding new data
	pipe := s.redis.Pipeline()

	// Add all new ads
	for userID, userAds := range adsByUser {
		key := fmt.Sprintf("user:%s:ads:new", userID.String()) // Временный ключ для новых данных
		members := make([]*redis.Z, len(userAds))

		for i, ad := range userAds {
			adKey := fmt.Sprintf("ad:%s", ad.AdID.String())

			// Store the complete ad data
			adData, err := json.Marshal(ad)
			if err != nil {
				return fmt.Errorf("failed to marshal ad data: %w", err)
			}

			pipe.Set(ctx, adKey, adData, 0)

			members[i] = &redis.Z{
				Score:  ad.Score.InexactFloat64(),
				Member: ad.AdID.String(),
			}
		}

		// Add new ads to the temporary sorted set
		if len(members) > 0 {
			pipe.ZAdd(ctx, key, members...)
		}
	}

	// Execute pipeline to add new data
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to add new data: %w", err)
	}

	// Now rename temporary keys to real keys and delete old data
	pipe = s.redis.Pipeline()
	for userID := range adsByUser {
		oldKey := fmt.Sprintf("user:%s:ads", userID.String())
		newKey := fmt.Sprintf("user:%s:ads:new", userID.String())

		// Get all old ad IDs to delete their data
		oldAds, err := s.redis.ZRange(ctx, oldKey, 0, -1).Result()
		if err != nil && !errors.Is(err, redis.Nil) {
			return fmt.Errorf("failed to get old ads: %w", err)
		}

		// Delete old ad data
		for _, adID := range oldAds {
			pipe.Del(ctx, fmt.Sprintf("ad:%s", adID))
		}

		// Rename new key to old key (атомарная операция)
		pipe.Rename(ctx, newKey, oldKey)
	}

	// Execute pipeline to switch to new data
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to switch to new data: %w", err)
	}

	return nil
}

func (s *storage) Remove(ctx context.Context, userID uuid.UUID, adID uuid.UUID) {
	key := fmt.Sprintf("user:%s:ads", userID.String())
	adKey := fmt.Sprintf("ad:%s", adID.String())

	pipe := s.redis.Pipeline()
	pipe.ZRem(ctx, key, adID.String())
	pipe.Del(ctx, adKey)
	_, _ = pipe.Exec(ctx)
}

func (s *storage) Get(ctx context.Context, userID uuid.UUID) (Ad, error) {
	key := fmt.Sprintf("user:%s:ads", userID.String())

	// First, get the highest scoring ad to determine our threshold
	topResults, err := s.redis.ZRevRangeByScoreWithScores(ctx, key, &redis.ZRangeBy{
		Min:    "-inf",
		Max:    "+inf",
		Offset: 0,
		Count:  1,
	}).Result()
	if err != nil {
		return Ad{}, err
	}

	if len(topResults) == 0 {
		return Ad{}, ErrAdNotFound
	}

	// Calculate the minimum acceptable score (90% of the highest score)
	topScore := topResults[0].Score
	minAcceptableScore := topScore * 0.9

	// Get all ads within the acceptable score range
	results, err := s.redis.ZRevRangeByScoreWithScores(ctx, key, &redis.ZRangeBy{
		Min: fmt.Sprintf("%f", minAcceptableScore),
		Max: fmt.Sprintf("%f", topScore),
	}).Result()
	if err != nil {
		return Ad{}, err
	}

	if len(results) == 0 {
		return Ad{}, ErrAdNotFound
	}

	// Randomly select one ad from the results using crypto/rand
	selectedResult := results[rand.Intn(len(results))]
	adID := selectedResult.Member.(string)
	adKey := fmt.Sprintf("ad:%s", adID)

	// Get the complete ad data
	adData, errGet := s.redis.Get(ctx, adKey).Bytes()
	if errGet != nil {
		return Ad{}, fmt.Errorf("failed to get ad data: %w", errGet)
	}

	var ad Ad
	if err := json.Unmarshal(adData, &ad); err != nil {
		return Ad{}, fmt.Errorf("failed to unmarshal ad data: %w", err)
	}

	// Remove the ad from Redis after retrieval
	pipe := s.redis.Pipeline()
	pipe.ZRem(ctx, key, adID)
	pipe.Del(ctx, adKey)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return Ad{}, fmt.Errorf("failed to remove ad after retrieval: %w", err)
	}

	return ad, nil
}

func (s *storage) GetMedianScore(ctx context.Context, userID uuid.UUID) (decimal.Decimal, error) {
	key := fmt.Sprintf("user:%s:ads", userID.String())

	// Get all scores from the sorted set
	scores, err := s.redis.ZRangeWithScores(ctx, key, 0, -1).Result()
	if err != nil {
		return decimal.Decimal{}, fmt.Errorf("failed to get scores from Redis: %w", err)
	}

	if len(scores) == 0 {
		return decimal.Decimal{}, nil
	}

	// Extract just the scores into a slice
	scoreValues := make([]float64, len(scores))
	for i, score := range scores {
		scoreValues[i] = score.Score
	}

	// Calculate median
	n := len(scoreValues)
	if n%2 == 0 {
		// Even number of scores - average the two middle values
		return decimal.NewFromFloat(scoreValues[n/2-1]).Add(decimal.NewFromFloat(scoreValues[n/2])).Div(decimal.New(2, 0)), nil
	}
	// Odd number of scores - return the middle value
	return decimal.NewFromFloat(scoreValues[n/2]), nil
}

func (s *storage) Close() error {
	return s.redis.Close()
}
