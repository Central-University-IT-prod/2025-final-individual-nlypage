package service

import (
	"context"
	"nlypage-final/internal/adapters/database/clickhouse"
	"nlypage-final/internal/adapters/database/postgres/ent"
	"nlypage-final/internal/adapters/database/postgres/ent/campaign"
	"nlypage-final/internal/adapters/database/postgres/ent/mlscore"
	"nlypage-final/internal/adapters/database/postgres/ent/targeting"
	"nlypage-final/internal/adapters/database/redis/ads"
	"nlypage-final/internal/domain/common/errorz"
	"nlypage-final/internal/domain/dto"
	"nlypage-final/pkg/ad_scoring"
	"nlypage-final/pkg/logger"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/shopspring/decimal"
)

type adTimeService interface {
	Now() *dto.CurrentDate
}

type CampaignStats struct {
	CampaignID       uuid.UUID
	ImpressionsCount int64
	ClicksCount      int64
	IsViewedByUser   bool
	IsClickedByUser  bool
}

type adClickhouseRepository interface {
	RecordImpression(ctx context.Context, show *clickhouse.AdImpression) error
	RecordClick(ctx context.Context, click *clickhouse.AdClick) error
	UserCampaignsStats(ctx context.Context, campaignIDs []uuid.UUID, userID uuid.UUID) (map[uuid.UUID]*clickhouse.UserCampaignStats, error)
	GetCampaignsSortedByUserViews(ctx context.Context, campaignIDs []uuid.UUID, userID uuid.UUID) ([]clickhouse.ViewsGroup, error)
}

type adsStorage interface {
	Get(ctx context.Context, userID uuid.UUID) (ads.Ad, error)
	Add(ctx context.Context, userID uuid.UUID, ad ads.Ad) error
	Remove(ctx context.Context, userID uuid.UUID, adID uuid.UUID)
}

type AdService interface {
	SelectAd(ctx context.Context, clientID dto.ClientAdGet) (*dto.Ad, error)
	RecordClick(ctx context.Context, click dto.ClientAdClick) error
}

type adService struct {
	db                   *ent.Client
	adScoring            ad_scoring.Scorer
	adsStorage           adsStorage
	clickhouseRepository adClickhouseRepository
	timeService          adTimeService
}

func NewAdService(
	db *ent.Client,
	adScoring ad_scoring.Scorer,
	adsStorage adsStorage,
	clickhouseRepository adClickhouseRepository,
	timeService adTimeService,
) AdService {
	return &adService{
		db:                   db,
		adScoring:            adScoring,
		adsStorage:           adsStorage,
		clickhouseRepository: clickhouseRepository,
		timeService:          timeService,
	}
}

func (a *adService) SelectAd(ctx context.Context, clientID dto.ClientAdGet) (*dto.Ad, error) {
	user, err := a.db.User.Get(ctx, clientID.ClientID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errorz.ErrNotFound
		}
		logger.Log.Errorf("failed to get user: %v", err)
		return nil, errorz.ErrInternal
	}

	logger.Log.Debugw("Get user",
		"client_id", user.ID,
		"age", user.Age,
		"location", user.Location,
		"gender", user.Gender,
	)

	// Получаем все активные кампании
	campaigns, err := a.db.Campaign.Query().
		Where(
			campaign.And(
				campaign.StartDateLTE(a.timeService.Now().CurrentDate),
				campaign.EndDateGTE(a.timeService.Now().CurrentDate),
				campaign.ModeratedEQ(true),
				campaign.HasTargetingWith(
					targeting.And(
						targeting.Or(
							targeting.AgeFromIsNil(),
							targeting.AgeFromLTE(user.Age),
						),
						targeting.Or(
							targeting.AgeToIsNil(),
							targeting.AgeToGTE(user.Age),
						),
						targeting.Or(
							targeting.LocationEQ(user.Location),
							targeting.LocationIsNil(),
						),
						targeting.Or(
							targeting.GenderEQ(targeting.Gender(user.Gender)),
							targeting.GenderEQ(targeting.GenderALL),
							targeting.GenderIsNil(),
						),
					),
				),
			),
		).
		All(ctx)

	if err != nil {
		logger.Log.Errorf("failed to get campaigns: %v", err)
		return nil, errorz.ErrInternal
	}
	logger.Log.Debugw("Found campaigns",
		"count", len(campaigns),
	)
	if len(campaigns) == 0 {
		return nil, errorz.ErrNotFound
	}

	// Получаем статистику для всех кампаний сразу
	campaignIDs := make([]uuid.UUID, len(campaigns))
	advertiserIDs := make([]uuid.UUID, len(campaigns))
	for i, camp := range campaigns {
		campaignIDs[i] = camp.ID
		advertiserIDs[i] = camp.AdvertiserID
	}

	// Bulk fetch ML scores for all campaigns
	mlScores, err := a.db.MlScore.Query().
		Where(
			mlscore.And(
				mlscore.UserID(user.ID),
				mlscore.AdvertiserIDIn(advertiserIDs...),
			),
		).
		All(ctx)
	if err != nil && !ent.IsNotFound(err) {
		logger.Log.Warnw("Failed to get ML scores",
			"user_id", user.ID.String(),
			"error", err,
		)
		return nil, errorz.ErrInternal
	}

	// Create a map for quick ML score lookup
	mlScoreMap := make(map[uuid.UUID]int64)
	for _, score := range mlScores {
		mlScoreMap[score.AdvertiserID] = score.Score
	}

	// Bulk fetch campaign stats and user interaction data
	campaignStats, err := a.clickhouseRepository.UserCampaignsStats(ctx, campaignIDs, user.ID)
	if err != nil {
		logger.Log.Warnw("Failed to get campaign stats",
			"error", err,
		)
		return nil, errorz.ErrInternal
	}

	logger.Log.Debugw("Found user campaigns stats",
		"stats", campaignStats,
	)

	// Find a suitable campaign and calculate its score
	type campaignWithScore struct {
		Campaign *ent.Campaign   `json:"campaign"`
		Score    decimal.Decimal `json:"score"`
	}
	selectedCampaignsMap := make(map[uuid.UUID]campaignWithScore)
	var filteredCampaignIDs []uuid.UUID

	for _, camp := range campaigns {
		stats, exists := campaignStats[camp.ID]
		if !exists {
			logger.Log.Warnw("No stats found for campaign",
				"campaign_id", camp.ID.String(),
			)
			continue
		}

		// Skip if user already clicked
		if stats.IsClickedByUser {
			continue
		}

		var costPerImpression float64
		if !stats.IsViewedByUser {
			costPerImpression = camp.CostPerImpression
		}

		// Calculate score for this campaign
		score := a.adScoring.CalculateScore(ad_scoring.Ad{
			MlScore:           mlScoreMap[camp.AdvertiserID],
			ImpressionsCount:  int(stats.ImpressionsCount),
			ImpressionsTarget: camp.ImpressionsLimit,
			CostPerImpression: costPerImpression,
			CostPerClick:      camp.CostPerClick,
			ClicksCount:       int(stats.ClicksCount),
			ClicksTarget:      camp.ClicksLimit,
		})

		logger.Log.Debugw("Calculated score for campaign",
			"campaign_id", camp.ID.String(),
			"score", score,
		)

		if score.GreaterThanOrEqual(a.adScoring.CalculateThreshold()) {
			selectedCampaignsMap[camp.ID] = campaignWithScore{
				Campaign: camp,
				Score:    score,
			}
			filteredCampaignIDs = append(filteredCampaignIDs, camp.ID)
		}
	}

	if len(filteredCampaignIDs) == 0 {
		return nil, errorz.ErrNotFound
	}

	viewGroups, err := a.clickhouseRepository.GetCampaignsSortedByUserViews(ctx, filteredCampaignIDs, user.ID)
	if err != nil {
		logger.Log.Warnw("Failed to get campaign stats",
			"error", err,
		)
		return nil, errorz.ErrInternal
	}

	logger.Log.Debugw("Campaign selection data",
		"campaigns_with_scores", selectedCampaignsMap,
		"view_groups", viewGroups,
	)

	// Проходим по группам от минимального количества просмотров к максимальному
	for _, group := range viewGroups {
		// Из текущей группы выбираем кампанию с наивысшим скором
		var bestCampaign *ent.Campaign
		var bestScore decimal.Decimal

		for _, campaignID := range group.Campaigns {
			if camp, exists := selectedCampaignsMap[campaignID]; exists {
				if bestCampaign == nil || camp.Score.GreaterThan(bestScore) {
					bestCampaign = camp.Campaign
					bestScore = camp.Score
				}
			}
		}

		if bestCampaign != nil {
			logger.Log.Infow("Selected campaign",
				"campaign_id", bestCampaign.ID.String(),
				"score", bestScore,
				"view_count", group.ViewCount,
			)

			if err := a.clickhouseRepository.RecordImpression(ctx, &clickhouse.AdImpression{
				CampaignID:   bestCampaign.ID,
				AdvertiserID: bestCampaign.AdvertiserID,
				ClientID:     user.ID,
				Income:       bestCampaign.CostPerImpression,
				Day:          a.timeService.Now().CurrentDate,
			}); err != nil {
				logger.Log.Warnw("Failed to record impression",
					"error", err,
				)
			}

			return &dto.Ad{
				AdID:         bestCampaign.ID,
				AdTitle:      bestCampaign.AdTitle,
				AdText:       bestCampaign.AdText,
				ImageURL:     bestCampaign.ImageURL,
				AdvertiserID: bestCampaign.AdvertiserID,
			}, nil
		}
	}

	logger.Log.Warnw("No suitable campaign found",
		"total_campaigns", len(campaigns),
		"filtered_campaigns", len(filteredCampaignIDs),
		"view_groups", len(viewGroups),
	)
	return nil, errorz.ErrNotFound
}

func (a *adService) RecordClick(ctx context.Context, click dto.ClientAdClick) error {
	camp, err := a.db.Campaign.Get(ctx, click.AdID)
	if err != nil {
		return &echo.HTTPError{
			Message: "campaign not found",
			Code:    echo.ErrNotFound.Code,
		}
	}

	if err := a.clickhouseRepository.RecordClick(ctx, &clickhouse.AdClick{
		CampaignID:   click.AdID,
		AdvertiserID: camp.AdvertiserID,
		ClientID:     click.ClientID,
		Income:       camp.CostPerClick,
		Day:          a.timeService.Now().CurrentDate,
	}); err != nil {
		return &echo.HTTPError{
			Message: err.Error(),
			Code:    echo.ErrConflict.Code,
		}
	}
	a.adsStorage.Remove(ctx, click.ClientID, click.AdID)

	return nil
}

// getPositionInList возвращает позицию элемента в списке (1-based)
func getPositionInList(list []uuid.UUID, item uuid.UUID) int {
	for i, v := range list {
		if v == item {
			return i + 1
		}
	}
	return 0
}

type campaignWithScore struct {
	Campaign *ent.Campaign   `json:"campaign"`
	Score    decimal.Decimal `json:"score"`
}
