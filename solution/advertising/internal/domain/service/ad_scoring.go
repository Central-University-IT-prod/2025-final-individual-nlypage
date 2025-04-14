package service

import (
	"context"
	"fmt"
	"nlypage-final/internal/adapters/database/clickhouse"
	"nlypage-final/internal/adapters/database/postgres/ent"
	"nlypage-final/internal/adapters/database/postgres/ent/campaign"
	"nlypage-final/internal/adapters/database/postgres/ent/mlscore"
	"nlypage-final/internal/adapters/database/redis/ads"
	"nlypage-final/internal/domain/dto"
	"nlypage-final/pkg/ad_scoring"
	"nlypage-final/pkg/logger"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"go.uber.org/zap"
)

// AdScoringService предоставляет интерфейс для управления сервисом подсчета рейтинга объявлений
//
// Deprecated: использовался для периодического перерасчёта рейтинга объявлений и сохранения в редис
// В связи с ограничениями тестовой системы - перестал использоваться
// Теперь не используется, хотя реализация осталась и идея хорошая под продакшн, было много идей, что здесь можно накрутить
// Но пришлось прийти к более простому решению, которое я не успел реализовать на том уровне, на котором хотелось
type AdScoringService interface {
	// Start запускает периодическое обновление рейтингов объявлений
	Start(ctx context.Context) error
	// Stop останавливает обновление рейтингов
	Stop() error
	// ForceUpdate принудительно запускает обновление рейтингов объявлений
	ForceUpdate(ctx context.Context) error
	// Recalculate принудительно пересчитывает рейтинг конкретного объявления у пользователя
	Recalculate(ctx context.Context, userID uuid.UUID, campaignID uuid.UUID) error
}

// adScoringTimeService предоставляет интерфейс для получения текущей даты
type adScoringTimeService interface {
	Now() *dto.CurrentDate
}

// adScoringClickhouse предоставляет интерфейс для работы со статистикой в Clickhouse
type adScoringClickhouse interface {
	CampaignStats(ctx context.Context, campaignID uuid.UUID) (*clickhouse.Stats, error)
	//IsViewed(ctx context.Context, adID uuid.UUID, userID uuid.UUID) (bool, error)
	//IsClicked(ctx context.Context, adID uuid.UUID, userID uuid.UUID) (bool, error)
}

// adsStorage предоставляет интерфейс для работы с объявлениями в Redis
type adScoringAdsStorage interface {
	UpdateAll(ctx context.Context, ads []ads.Ad) error
	GetMedianScore(ctx context.Context, userID uuid.UUID) (decimal.Decimal, error)
	Remove(ctx context.Context, userID uuid.UUID, adID uuid.UUID)
	Add(ctx context.Context, userID uuid.UUID, ad ads.Ad) error
}

// adScoringService реализует сервис подсчета рейтинга объявлений
type adScoringService struct {
	timeService adScoringTimeService
	db          *ent.Client
	clickhouse  adScoringClickhouse
	scorer      ad_scoring.Scorer
	adsStorage  adScoringAdsStorage
	logger      *logger.Logger
	interval    time.Duration
	cancelFunc  context.CancelFunc
	mu          sync.Mutex
}

// NewAdScoringService создает новый экземпляр сервиса подсчета рейтинга объявлений
func NewAdScoringService(
	timeService adScoringTimeService,
	db *ent.Client,
	clickhouse adScoringClickhouse,
	scorer ad_scoring.Scorer,
	adsStorage adScoringAdsStorage,
	logger *logger.Logger,
	interval time.Duration,
) AdScoringService {
	return &adScoringService{
		timeService: timeService,
		db:          db,
		clickhouse:  clickhouse,
		scorer:      scorer,
		adsStorage:  adsStorage,
		logger:      logger,
		interval:    interval,
	}
}

func (s *adScoringService) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cancelFunc != nil {
		return fmt.Errorf("scoring service is already running")
	}

	ctx, cancel := context.WithCancel(ctx)
	s.cancelFunc = cancel

	return s.run(ctx)
}

func (s *adScoringService) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cancelFunc == nil {
		return fmt.Errorf("scoring service is not running")
	}

	s.cancelFunc()
	s.cancelFunc = nil
	return nil
}

// ForceUpdate принудительно запускает обновление рейтингов объявлений
func (s *adScoringService) ForceUpdate(ctx context.Context) error {
	s.logger.Info("Starting forced ad score update")
	s.updateScores(ctx)
	return nil
}

func (s *adScoringService) Recalculate(ctx context.Context, userID uuid.UUID, campaignID uuid.UUID) error {
	s.logger.Debugw("Recalculating ad score", "user_id", userID, "campaign_id", campaignID)

	// Получаем пользователя
	u, err := s.db.User.Get(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Получаем кампанию с таргетингом
	c, err := s.db.Campaign.Query().
		Where(campaign.ID(campaignID)).
		Where(
			campaign.And(
				campaign.ModeratedEQ(true),
				campaign.StartDateLTE(s.timeService.Now().CurrentDate),
				campaign.EndDateGTE(s.timeService.Now().CurrentDate),
			),
		).
		WithTargeting().
		Only(ctx)
	if err != nil {
		return fmt.Errorf("failed to get campaign: %w", err)
	}

	// Проверяем таргетинг
	//if !s.matchesTargeting(u, c.Edges.Targeting) {
	//	return nil // Пропускаем, если таргетинг не подходит
	//}

	// Создаем campaignData для обработки
	cd := campaignData{
		Campaign: c,
		userID:   userID,
	}

	// Получаем статистику по кампании
	campStats, err := s.clickhouse.CampaignStats(ctx, campaignID)
	if err != nil {
		return fmt.Errorf("failed to get campaign stats: %w", err)
	}

	cd.impressionsCount = int(campStats.ImpressionsCount)
	cd.clicksCount = int(campStats.ClicksCount)

	var mlScoreInt int64
	mlScore, err := s.db.MlScore.Query().
		Where(
			mlscore.And(
				mlscore.UserID(u.ID),
				mlscore.AdvertiserID(c.AdvertiserID),
			),
		).
		First(ctx)
	if err != nil && !ent.IsNotFound(err) {
		s.logger.Warnw("Failed to get ML score",
			"user_id", u.ID.String(),
			"campaign_id", c.ID.String(),
			"error", err,
		)
		return fmt.Errorf("failed to get ML score: %w", err)
	}

	if mlScore != nil {
		mlScoreInt = mlScore.Score
	}

	// Вычисляем скор
	cd.score = s.scorer.CalculateScore(ad_scoring.Ad{
		MlScore:           mlScoreInt,
		ImpressionsCount:  cd.impressionsCount,
		ImpressionsTarget: c.ImpressionsLimit,
		CostPerImpression: c.CostPerImpression,
		ClicksCount:       cd.clicksCount,
		ClicksTarget:      c.ClicksLimit,
		CostPerClick:      c.CostPerClick,
	})

	medianScore, err := s.adsStorage.GetMedianScore(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get median score: %w", err)
	}

	if cd.score.LessThan(medianScore) {
		return nil
	}

	// Обновляем объявление
	if err := s.adsStorage.Add(ctx, userID, ads.Ad{
		UserID:            userID,
		AdID:              c.ID,
		AdTitle:           c.AdTitle,
		AdText:            c.AdText,
		ImageURL:          c.ImageURL,
		AdvertiserID:      c.AdvertiserID,
		CostPerImpression: c.CostPerImpression,
		Score:             cd.score,
	}); err != nil {
		return fmt.Errorf("failed to update ad score: %w", err)
	}
	return nil
}

func (s *adScoringService) run(ctx context.Context) error {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			s.updateScores(ctx)
		}
	}
}

// campaignData содержит данные кампании и статистику показов/кликов
type campaignData struct {
	*ent.Campaign
	impressionsCount int
	clicksCount      int
	score            decimal.Decimal
	userID           uuid.UUID
}

// updateScores обновляет рейтинги всех объявлений для всех пользователей
func (s *adScoringService) updateScores(ctx context.Context) {
	s.logger.Info("Starting ad score update")

	// Получаем всех пользователей
	users, err := s.db.User.Query().All(ctx)
	if err != nil {
		s.logger.Error("Failed to get users", zap.Error(err))
		return
	}

	// Получаем все активные кампании с таргетингом
	campaigns, err := s.db.Campaign.Query().
		Where(
			campaign.And(
				campaign.ModeratedEQ(true),
				campaign.StartDateLTE(s.timeService.Now().CurrentDate),
				campaign.EndDateGTE(s.timeService.Now().CurrentDate),
			),
		).
		WithTargeting().
		All(ctx)
	if err != nil {
		s.logger.Error("Failed to get campaigns", zap.Error(err))
		return
	}

	// Собираем статистику по кампаниям
	var campaignsData []campaignData
	for _, c := range campaigns {
		campStats, err := s.clickhouse.CampaignStats(ctx, c.ID)
		if err != nil {
			s.logger.Errorw("Failed to get campaign stats",
				"campaign_id", c.ID.String(),
				"error", err,
			)
			continue
		}

		campaignsData = append(campaignsData, campaignData{
			Campaign:         c,
			impressionsCount: int(campStats.ImpressionsCount),
			clicksCount:      int(campStats.ClicksCount),
		})
	}

	// Для каждого пользователя собираем подходящие объявления
	var (
		adsToAdd []ads.Ad
		usersAds []campaignData
		// addedAdsCountMap используется для отслеживания количества объявлений, которые будут добавлены в Redis
		// Это делается для того чтобы не добавлять в редис больше объявлений чем нужно
		addedAdsCountMap = make(map[uuid.UUID]int)
	)
	for _, u := range users {
		select {
		case <-ctx.Done():
			return
		default:
			userAds, err := s.collectUserAds(ctx, u, campaignsData)
			if err != nil {
				s.logger.Errorw("Failed to collect ads for user",
					"user_id", u.ID.String(),
					"error", err,
				)
				continue
			}

			usersAds = append(usersAds, userAds...)
		}
	}

	// Сортируем объявления по рейтингу
	sort.Slice(usersAds, func(i, j int) bool {
		return usersAds[i].score.GreaterThan(usersAds[j].score)
	})

	// Добавляем объявления в Redis
	for _, a := range usersAds {
		if _, ok := addedAdsCountMap[a.ID]; !ok {
			addedAdsCountMap[a.ID] = 0
		}

		// Добавляем объявление в Redis только если количество потенциальных показов не превышает лимит
		if addedAdsCountMap[a.ID] < a.ImpressionsLimit-a.impressionsCount {
			//addedAdsCountMap[a.ID]++
			adsToAdd = append(adsToAdd, ads.Ad{
				UserID:            a.userID,
				AdID:              a.ID,
				AdTitle:           a.AdTitle,
				AdText:            a.AdText,
				ImageURL:          a.ImageURL,
				AdvertiserID:      a.AdvertiserID,
				CostPerImpression: a.CostPerImpression,
				Score:             a.score,
			})
		}
	}

	// Обновляем все объявления в Redis одним вызовом
	if err := s.adsStorage.UpdateAll(ctx, adsToAdd); err != nil {
		s.logger.Error("Failed to update all ads in Redis", zap.Error(err))
		return
	}
	s.logger.Info("Ad scores updated in Redis")
}

// calculateMedianScore вычисляет медианный скор из списка объявлений
func (s *adScoringService) calculateMedianScore(ads []campaignData) decimal.Decimal {
	if len(ads) == 0 {
		return decimal.Zero
	}

	scores := make([]decimal.Decimal, len(ads))
	for i, ad := range ads {
		scores[i] = ad.score
	}
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].LessThan(scores[j])
	})

	if len(scores)%2 == 0 {
		// Для четного количества берем среднее двух средних значений
		mid := len(scores) / 2
		return scores[mid-1].Add(scores[mid]).Div(decimal.NewFromInt(2))
	}
	// Для нечетного количества берем среднее значение
	return scores[len(scores)/2]
}

// collectUserAds собирает подходящие объявления для конкретного пользователя
func (s *adScoringService) collectUserAds(ctx context.Context, u *ent.User, campaigns []campaignData) ([]campaignData, error) {
	var matchedAds []campaignData

	for _, c := range campaigns {
		_, err := c.Campaign.Edges.TargetingOrErr()
		if err != nil {
			s.logger.Errorw("Failed to get targeting",
				"campaign_id", c.ID.String(),
				"error", err,
			)
			continue
		}
		// Проверяем таргетинг
		//if !s.matchesTargeting(u, t) {
		//	continue
		//}

		var costPerImpression float64
		// Проверяем, было ли показано объявление
		//isViewed, err := s.clickhouse.IsViewed(ctx, c.ID, u.ID)
		//if err != nil {
		//	s.logger.Errorw("Failed to get isViewed",
		//		"campaign_id", c.ID.String(),
		//		"error", err,
		//	)
		//	continue
		//}
		//if !isViewed {
		//	costPerImpression = c.Campaign.CostPerImpression
		//}

		// Проверяем, был ли клик по объявлению
		//isClicked, err := s.clickhouse.IsClicked(ctx, c.ID, u.ID)
		//if err != nil {
		//	s.logger.Errorw("Failed to get isClicked",
		//		"campaign_id", c.ID.String(),
		//		"error", err,
		//	)
		//	continue
		//}
		//if isClicked {
		//	continue
		//}

		// Получаем ML score для пары пользователь-рекламодатель
		var mlScoreInt int64
		mlScore, err := s.db.MlScore.Query().
			Where(
				mlscore.And(
					mlscore.UserID(u.ID),
					mlscore.AdvertiserID(c.Campaign.AdvertiserID),
				),
			).
			First(ctx)

		if err != nil && !ent.IsNotFound(err) {
			s.logger.Warnw("Failed to get ML score",
				"campaign_id", c.Campaign.ID.String(),
				"error", err,
			)
			continue
		}

		// Если ML score найден, используем его значение, иначе используем 0
		if mlScore != nil {
			mlScoreInt = mlScore.Score
		}

		score := s.scorer.CalculateScore(ad_scoring.Ad{
			MlScore:           mlScoreInt,
			ImpressionsCount:  c.impressionsCount,
			ImpressionsTarget: c.Campaign.ImpressionsLimit,
			ClicksCount:       c.clicksCount,
			ClicksTarget:      c.Campaign.ClicksLimit,
			CostPerImpression: costPerImpression,
			CostPerClick:      c.Campaign.CostPerClick,
		})

		s.logger.Debugw(
			"Calculated score",
			"campaign_id", c.Campaign.ID.String(),
			"score", score,
		)

		matchedAds = append(matchedAds, campaignData{
			Campaign:         c.Campaign,
			impressionsCount: c.impressionsCount,
			clicksCount:      c.clicksCount,
			score:            score,
			userID:           u.ID,
		})
	}

	if len(matchedAds) == 0 {
		return matchedAds, nil
	}

	medianScore := s.calculateMedianScore(matchedAds)
	s.logger.Debugw("Calculated median score", "median_score", medianScore)

	// Убираем объявления, у которых скор меньше медианного
	var filteredAds []campaignData
	for _, ad := range matchedAds {
		if ad.score.GreaterThanOrEqual(medianScore) {
			filteredAds = append(filteredAds, ad)
		}
	}

	return filteredAds, nil
}

//// matchesTargeting проверяет, соответствует ли пользователь таргетингу кампании
//func (s *adScoringService) matchesTargeting(u *ent.User, t *ent.Targeting) bool {
//	if t == nil {
//		return true
//	}
//
//	// Проверяем пол
//	if t.Gender.String() != "ALL" && t.Gender.String() != u.Gender.String() {
//		return false
//	}
//
//	// Проверяем возраст
//	if u.Age < t.AgeFrom {
//		return false
//	}
//	if u.Age > t.AgeTo {
//		return false
//	}
//
//	// Проверяем локацию
//	if t.Location != "" && t.Location != u.Location {
//		return false
//	}
//
//	return true
//}
