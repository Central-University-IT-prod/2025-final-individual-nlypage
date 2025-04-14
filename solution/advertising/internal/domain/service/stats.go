package service

import (
	"context"
	"github.com/google/uuid"
	"nlypage-final/internal/adapters/database/clickhouse"
	"nlypage-final/internal/domain/dto"
)

type statsTimeService interface {
	Now() *dto.CurrentDate
}

type statsClickhouseRepository interface {
	CampaignStats(ctx context.Context, campaignID uuid.UUID) (*clickhouse.Stats, error)
	CampaignDailyStats(ctx context.Context, campaignID uuid.UUID) ([]*clickhouse.StatsDaily, error)
	AdvertiserStats(ctx context.Context, advertiserID uuid.UUID) (*clickhouse.Stats, error)
	AdvertiserDailyStats(ctx context.Context, advertiserID uuid.UUID) ([]*clickhouse.StatsDaily, error)
}

type StatsService interface {
	Campaign(ctx context.Context, campaignID uuid.UUID) (*dto.Stats, error)
	CampaignDaily(ctx context.Context, campaignID uuid.UUID) ([]*dto.StatsDaily, error)
	Advertiser(ctx context.Context, advertiserID uuid.UUID) (*dto.Stats, error)
	AdvertiserDaily(ctx context.Context, advertiserID uuid.UUID) ([]*dto.StatsDaily, error)
}

type statsService struct {
	timeService          statsTimeService
	clickhouseRepository statsClickhouseRepository
}

func NewStatsService(timeService statsTimeService, clickhouseRepository statsClickhouseRepository) StatsService {
	return &statsService{
		timeService:          timeService,
		clickhouseRepository: clickhouseRepository,
	}
}

func (s *statsService) Campaign(ctx context.Context, campaignID uuid.UUID) (*dto.Stats, error) {
	stats, err := s.clickhouseRepository.CampaignStats(ctx, campaignID)
	if err != nil {
		return nil, err
	}
	return &dto.Stats{
		ImpressionsCount: int(stats.ImpressionsCount),
		ClicksCount:      int(stats.ClicksCount),
		Conversion:       stats.Conversion,
		SpentImpressions: stats.SpentImpressions,
		SpentClicks:      stats.SpentClicks,
		SpentTotal:       stats.SpentTotal,
	}, nil
}

func (s *statsService) CampaignDaily(ctx context.Context, campaignID uuid.UUID) ([]*dto.StatsDaily, error) {
	stats, err := s.clickhouseRepository.CampaignDailyStats(ctx, campaignID)
	if err != nil {
		return nil, err
	}

	var statsDaily []*dto.StatsDaily
	for _, stat := range stats {
		statsDaily = append(statsDaily, &dto.StatsDaily{
			Stats: dto.Stats{
				ImpressionsCount: int(stat.ImpressionsCount),
				ClicksCount:      int(stat.ClicksCount),
				Conversion:       stat.Conversion,
				SpentImpressions: stat.SpentImpressions,
				SpentClicks:      stat.SpentClicks,
				SpentTotal:       stat.SpentTotal,
			},
			Date: int(stat.Date),
		})
	}
	return statsDaily, nil
}

func (s *statsService) Advertiser(ctx context.Context, advertiserID uuid.UUID) (*dto.Stats, error) {
	stats, err := s.clickhouseRepository.AdvertiserStats(ctx, advertiserID)
	if err != nil {
		return nil, err
	}
	return &dto.Stats{
		ImpressionsCount: int(stats.ImpressionsCount),
		ClicksCount:      int(stats.ClicksCount),
		Conversion:       stats.Conversion,
		SpentImpressions: stats.SpentImpressions,
		SpentClicks:      stats.SpentClicks,
		SpentTotal:       stats.SpentTotal,
	}, nil
}

func (s *statsService) AdvertiserDaily(ctx context.Context, advertiserID uuid.UUID) ([]*dto.StatsDaily, error) {
	stats, err := s.clickhouseRepository.AdvertiserDailyStats(ctx, advertiserID)
	if err != nil {
		return nil, err
	}

	var statsDaily []*dto.StatsDaily
	for _, stat := range stats {
		statsDaily = append(statsDaily, &dto.StatsDaily{
			Stats: dto.Stats{
				ImpressionsCount: int(stat.ImpressionsCount),
				ClicksCount:      int(stat.ClicksCount),
				Conversion:       stat.Conversion,
				SpentImpressions: stat.SpentImpressions,
				SpentClicks:      stat.SpentClicks,
				SpentTotal:       stat.SpentTotal,
			},
			Date: int(stat.Date),
		})
	}
	return statsDaily, nil
}
