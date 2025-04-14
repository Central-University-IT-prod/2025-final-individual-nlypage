package ad_scoring

import (
	"math"
	"sort"
	"sync"

	"github.com/shopspring/decimal"
)

type Ad struct {
	ID                string
	MlScore           int64
	ImpressionsCount  int
	ImpressionsTarget int
	CostPerImpression float64
	ClicksCount       int
	ClicksTarget      int
	CostPerClick      float64
}

type Config struct {
	PlatformProfitWeight float64
	RelevanceWeight      float64
	PerformanceWeight    float64
}

type Scorer interface {
	CalculateScore(ad Ad) decimal.Decimal
	CalculateThreshold() decimal.Decimal
}

type adHistory struct {
	Score       decimal.Decimal `json:"score"`
	Impressions int             `json:"impressions"`
	Clicks      int             `json:"clicks"`
	ImprTarget  int             `json:"impr_target"`
	ClickTarget int             `json:"click_target"`
}

type scorer struct {
	config  Config
	history []adHistory
	mu      sync.RWMutex
}

func NewScorer(config Config) Scorer {
	return &scorer{config: config}
}

func (s *scorer) CalculateScore(ad Ad) decimal.Decimal {
	//logger.Log.Debugw("Calculating score for ad", "ad", ad)
	// 1. Нормализация весов
	totalWeight := s.config.PlatformProfitWeight + s.config.RelevanceWeight + s.config.PerformanceWeight
	pw := s.config.PlatformProfitWeight / totalWeight
	rw := s.config.RelevanceWeight / totalWeight
	perfw := s.config.PerformanceWeight / totalWeight

	// 2. Релевантность через сигмоид
	relevanceScore := decimal.NewFromFloat((1 / (1 + math.Exp(-float64(ad.MlScore)/1000))) * rw)

	// 2. Прибыль с нормализацией через сигмоид
	revenue := ad.CostPerImpression*float64(ad.ImpressionsTarget) + ad.CostPerClick*float64(ad.ClicksTarget)
	profitScore := decimal.NewFromFloat((1 / (1 + math.Exp(-revenue/10000))) * pw) // Нормализация к 0-1

	// 4. Эффективность с постепенным снижением по мере приближения к целевым показателям и резким снижением после их преодоления
	var impressionRatio, clickRatio = 1.0, 1.0

	calculateDeviation := func(actual, target int) float64 {
		if target == 0 {
			return 1.0
		}

		if float64(actual+1) < float64(target)*1.05 {
			remainingRatio := float64(target-actual) / float64(target)
			return math.Pow(remainingRatio, 0.1)
		}

		// Превышение: очень быстрое снижение
		deviation := float64(actual-target) / float64(target)
		// Увеличиваем коэффициент для более резкого снижения
		return 1.0 - math.Exp(math.Abs(deviation))
	}

	s.mu.Lock()
	impressionRatio = calculateDeviation(ad.ImpressionsCount, ad.ImpressionsTarget)
	clickRatio = calculateDeviation(ad.ClicksCount, ad.ClicksTarget)
	s.mu.Unlock()
	// logger.Log.Debugw("ratio data",
	// 	"impressionRatio", impressionRatio,
	// 	"clickRatio", clickRatio,
	// 	"impressionCount", ad.ImpressionsCount,
	// 	"impressionTarget", ad.ImpressionsTarget,
	// 	"clickTarget", ad.ClicksTarget,
	// 	"clickCount", ad.ClicksCount,
	// )

	performanceScore := decimal.NewFromFloat((impressionRatio + clickRatio) * perfw)

	// 5. Итоговый score с ограничением до 1.0
	// logger.Log.Debugw("calculated scores",
	// 	"profitScore", profitScore,
	// 	"relevanceScore", relevanceScore,
	// 	"performanceScore", performanceScore,
	// )
	totalScore := profitScore.Add(relevanceScore).Add(performanceScore)
	one := decimal.NewFromFloat(1.0)
	if totalScore.GreaterThan(one) {
		totalScore = one
	}

	if totalScore.GreaterThanOrEqual(baseThreshold) {
		s.mu.Lock()
		s.history = append(s.history, adHistory{
			Score:       totalScore,
			Impressions: ad.ImpressionsCount,
			Clicks:      ad.ClicksCount,
			ImprTarget:  ad.ImpressionsTarget,
			ClickTarget: ad.ClicksTarget,
		})
		s.mu.Unlock()
	}

	// Ограничиваем размер истории 1000 записей
	if len(s.history) > 1000 {
		s.mu.Lock()
		s.history = s.history[1:]
		s.mu.Unlock()
	}

	// Рассчитываем порог
	return totalScore
}

var baseThreshold = decimal.NewFromFloat(0.7)

func (s *scorer) CalculateThreshold() decimal.Decimal {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.history) == 0 {
		return baseThreshold
	}

	// Собираем уникальные скоры через map
	uniqueScores := make(map[string]decimal.Decimal)
	for _, h := range s.history {
		// Используем строковое представление для ключа map
		uniqueScores[h.Score.String()] = h.Score
	}

	// Преобразуем map в слайс
	scores := make([]decimal.Decimal, 0, len(uniqueScores))
	for _, score := range uniqueScores {
		scores = append(scores, score)
	}

	// Сортируем скоры
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].LessThan(scores[j])
	})

	// Вычисляем индекс для 80-го перцентиля
	idx := int(float64(len(scores)) * 0.80)
	if idx >= len(scores) {
		idx = len(scores) - 1
	}

	// logger.Log.Debugw("Calculating threshold",
	// 	"total_history", len(s.history),
	// 	"unique_scores", len(scores),
	// 	"percentile_80_index", idx,
	// 	"threshold", scores[idx])

	return scores[idx]
}
