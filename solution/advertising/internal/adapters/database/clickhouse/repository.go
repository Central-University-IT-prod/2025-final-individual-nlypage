package clickhouse

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/google/uuid"
)

type Config struct {
	Host     string
	Port     string
	Database string
	Username string
	Password string
	Debug    bool
}

type Repository struct {
	conn driver.Conn
}

func New(cfg Config) (*Repository, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)},
		Auth: clickhouse.Auth{
			Database: cfg.Database,
			Username: cfg.Username,
			Password: cfg.Password,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		Debug: cfg.Debug,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to clickhouse: %w", err)
	}

	if err := createTables(conn); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return &Repository{conn: conn}, nil
}

func createTables(conn driver.Conn) error {
	queries := []string{
		`
        CREATE TABLE IF NOT EXISTS ad_impressions (
            campaign_id UUID,
            advertiser_id UUID,
            client_id UUID,
            income Float64,
            day Int32,
            view_count UInt64,
            PRIMARY KEY (day, campaign_id, client_id)
        ) ENGINE = ReplacingMergeTree()
        ORDER BY (day, campaign_id, client_id)
        `,
		`
        CREATE TABLE IF NOT EXISTS ad_clicks (
            campaign_id UUID,
            advertiser_id UUID,
            client_id UUID,
            income Float64,
            day Int32,
            PRIMARY KEY (day, campaign_id, client_id)
        ) ENGINE = ReplacingMergeTree() 
        ORDER BY (day, campaign_id, client_id)
        `,
	}

	for _, query := range queries {
		if err := conn.Exec(context.Background(), query); err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}
	}

	return nil
}

// RecordImpression записывает показ рекламы или инкрементирует счетчик просмотров если показ уже существует
func (r *Repository) RecordImpression(ctx context.Context, show *AdImpression) error {
	// Используем INSERT ... SELECT для атомарного инкремента
	query := `
		INSERT INTO ad_impressions (
			campaign_id,
			advertiser_id,
			client_id,
			income,
			day,
			view_count
		)
		SELECT 
			campaign_id,
			advertiser_id,
			client_id,
			income,
			day,
			view_count + 1
		FROM 
		(
			SELECT 
				? as campaign_id,
				? as advertiser_id,
				? as client_id,
				? as income,
				? as day,
				coalesce(max(view_count), 0) as view_count
			FROM ad_impressions FINAL
			WHERE campaign_id = ? AND client_id = ? AND day = ?
		)
	`

	if err := r.conn.Exec(ctx, query,
		show.CampaignID,
		show.AdvertiserID,
		show.ClientID,
		show.Income,
		show.Day,
		show.CampaignID,
		show.ClientID,
		show.Day,
	); err != nil {
		return fmt.Errorf("failed to record impression: %w", err)
	}

	return nil
}

// RecordClick записывает клик по рекламе, возвращая ошибку если клик уже существует
func (r *Repository) RecordClick(ctx context.Context, click *AdClick) error {
	// Проверяем показана ли реклама
	checkImpressionQuery := `
        SELECT count(*)
        FROM ad_impressions
        WHERE campaign_id = ? AND client_id = ?
    `

	var impressionCount uint64
	row := r.conn.QueryRow(ctx, checkImpressionQuery, click.CampaignID, click.ClientID)
	if err := row.Scan(&impressionCount); err != nil {
		return fmt.Errorf("failed to check impression existence: %w", err)
	}

	if impressionCount == 0 {
		return ErrClickAdNotShown
	}

	// Проверяем существование клика
	checkClickQuery := `
        SELECT count(*)
        FROM ad_clicks
        WHERE campaign_id = ? AND client_id = ?
    `

	var clickCount uint64
	row = r.conn.QueryRow(ctx, checkClickQuery, click.CampaignID, click.ClientID)
	if err := row.Scan(&clickCount); err != nil {
		return fmt.Errorf("failed to check existing click: %w", err)
	}

	if clickCount > 0 {
		return ErrClickAlreadyExists
	}

	// Если клика нет, записываем его
	query := `
        INSERT INTO ad_clicks (
            campaign_id,
            advertiser_id,
            client_id,
            income,
            day
        ) VALUES (?, ?, ?, ?, ?)
    `

	if err := r.conn.Exec(ctx, query,
		click.CampaignID,
		click.AdvertiserID,
		click.ClientID,
		click.Income,
		click.Day,
	); err != nil {
		return fmt.Errorf("failed to record click: %w", err)
	}

	return nil
}

func (r *Repository) DeleteStatsByCampaignID(ctx context.Context, campaignID uuid.UUID) error {
	// Удаляем показы рекламы
	deleteImpressionsQuery := `
        ALTER TABLE ad_impressions 
        DELETE WHERE campaign_id = ?
    `
	if err := r.conn.Exec(ctx, deleteImpressionsQuery, campaignID); err != nil {
		return fmt.Errorf("failed to delete impressions: %w", err)
	}

	// Удаляем клики по рекламе
	deleteClicksQuery := `
        ALTER TABLE ad_clicks 
        DELETE WHERE campaign_id = ?
    `
	if err := r.conn.Exec(ctx, deleteClicksQuery, campaignID); err != nil {
		return fmt.Errorf("failed to delete clicks: %w", err)
	}

	return nil
}

// CampaignStats возвращает статистику по кампании
func (r *Repository) CampaignStats(ctx context.Context, campaignID uuid.UUID) (*Stats, error) {
	query := `
		WITH 
			impressions AS (
				SELECT 
					count(*) as imp_count,
					sum(income) as imp_income
				FROM ad_impressions 
				WHERE campaign_id = ?
			),
			clicks AS (
				SELECT 
					count(*) as click_count,
					sum(income) as click_income
				FROM ad_clicks 
				WHERE campaign_id = ?
			)
		SELECT 
			imp_count as impressions,
			click_count as clicks,
			if(imp_count > 0, click_count/imp_count * 100, 0) as conversion,
			coalesce(imp_income, 0) as impression_income,
			coalesce(click_income, 0) as click_income,
			coalesce(imp_income, 0) + coalesce(click_income, 0) as total_income
		FROM impressions
		CROSS JOIN clicks
	`

	var stats Stats
	row := r.conn.QueryRow(ctx, query, campaignID, campaignID)
	if err := row.Scan(&stats.ImpressionsCount, &stats.ClicksCount, &stats.Conversion, &stats.SpentImpressions, &stats.SpentClicks, &stats.SpentTotal); err != nil {
		return nil, fmt.Errorf("failed to get campaign stats: %w", err)
	}

	return &stats, nil
}

// CampaignDailyStats возвращает ежедневную статистику по кампании
func (r *Repository) CampaignDailyStats(ctx context.Context, campaignID uuid.UUID) ([]*StatsDaily, error) {
	query := `
		WITH 
			daily_impressions AS (
				SELECT 
					day,
					count(*) as imp_count,
					sum(income) as imp_income
				FROM ad_impressions 
				WHERE campaign_id = ?
				GROUP BY day
			),
			daily_clicks AS (
				SELECT 
					day,
					count(*) as click_count,
					sum(income) as click_income
				FROM ad_clicks 
				WHERE campaign_id = ?
				GROUP BY day
			)
		SELECT 
			di.day,
			di.imp_count as impressions,
			COALESCE(dc.click_count, 0) as clicks,
			if(di.imp_count > 0, COALESCE(dc.click_count, 0)/di.imp_count * 100, 0) as conversion,
			COALESCE(di.imp_income, 0) as impression_income,
			COALESCE(dc.click_income, 0) as click_income,
			COALESCE(di.imp_income, 0) + COALESCE(dc.click_income, 0) as total_income
		FROM daily_impressions di
		LEFT JOIN daily_clicks dc ON dc.day = di.day
		ORDER BY di.day
	`

	rows, err := r.conn.Query(ctx, query, campaignID, campaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to query daily campaign stats: %w", err)
	}
	defer rows.Close()

	var stats []*StatsDaily
	for rows.Next() {
		var stat StatsDaily
		if err := rows.Scan(&stat.Date, &stat.ImpressionsCount, &stat.ClicksCount, &stat.Conversion, &stat.SpentImpressions, &stat.SpentClicks, &stat.SpentTotal); err != nil {
			return nil, fmt.Errorf("failed to scan daily campaign stats: %w", err)
		}
		stats = append(stats, &stat)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating daily campaign stats: %w", err)
	}

	return stats, nil
}

// AdvertiserStats возвращает статистику по рекламодателю
func (r *Repository) AdvertiserStats(ctx context.Context, advertiserID uuid.UUID) (*Stats, error) {
	query := `
		WITH 
			impressions AS (
				SELECT 
					count(*) as imp_count,
					sum(income) as imp_income
				FROM ad_impressions 
				WHERE advertiser_id = ?
			),
			clicks AS (
				SELECT 
					count(*) as click_count,
					sum(income) as click_income
				FROM ad_clicks 
				WHERE advertiser_id = ?
			)
		SELECT 
			imp_count as impressions,
			click_count as clicks,
			if(imp_count > 0, click_count/imp_count * 100, 0) as conversion,
			coalesce(imp_income, 0) as impression_income,
			coalesce(click_income, 0) as click_income,
			coalesce(imp_income, 0) + coalesce(click_income, 0) as total_income
		FROM impressions
		CROSS JOIN clicks
	`

	var stats Stats
	row := r.conn.QueryRow(ctx, query, advertiserID, advertiserID)
	if err := row.Scan(&stats.ImpressionsCount, &stats.ClicksCount, &stats.Conversion, &stats.SpentImpressions, &stats.SpentClicks, &stats.SpentTotal); err != nil {
		return nil, fmt.Errorf("failed to get advertiser stats: %w", err)
	}

	return &stats, nil
}

// AdvertiserDailyStats возвращает ежедневную статистику по рекламодателю
func (r *Repository) AdvertiserDailyStats(ctx context.Context, advertiserID uuid.UUID) ([]*StatsDaily, error) {
	query := `
		WITH 
			daily_impressions AS (
				SELECT 
					day,
					count(*) as imp_count,
					sum(income) as imp_income
				FROM ad_impressions 
				WHERE advertiser_id = ?
				GROUP BY day
			),
			daily_clicks AS (
				SELECT 
					day,
					count(*) as click_count,
					sum(income) as click_income
				FROM ad_clicks 
				WHERE advertiser_id = ?
				GROUP BY day
			)
		SELECT 
			di.day,
			di.imp_count as impressions,
			COALESCE(dc.click_count, 0) as clicks,
			if(di.imp_count > 0, COALESCE(dc.click_count, 0)/di.imp_count * 100, 0) as conversion,
			COALESCE(di.imp_income, 0) as impression_income,
			COALESCE(dc.click_income, 0) as click_income,
			COALESCE(di.imp_income, 0) + COALESCE(dc.click_income, 0) as total_income
		FROM daily_impressions di
		LEFT JOIN daily_clicks dc ON dc.day = di.day
		ORDER BY di.day
	`

	rows, err := r.conn.Query(ctx, query, advertiserID, advertiserID)
	if err != nil {
		return nil, fmt.Errorf("failed to query daily advertiser stats: %w", err)
	}
	defer rows.Close()

	var stats []*StatsDaily
	for rows.Next() {
		var stat StatsDaily
		if err := rows.Scan(&stat.Date, &stat.ImpressionsCount, &stat.ClicksCount, &stat.Conversion, &stat.SpentImpressions, &stat.SpentClicks, &stat.SpentTotal); err != nil {
			return nil, fmt.Errorf("failed to scan daily advertiser stats: %w", err)
		}
		stats = append(stats, &stat)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating daily advertiser stats: %w", err)
	}

	return stats, nil
}

func (r *Repository) UserCampaignsStats(ctx context.Context, campaignIDs []uuid.UUID, userID uuid.UUID) (map[uuid.UUID]*UserCampaignStats, error) {
	if len(campaignIDs) == 0 {
		return make(map[uuid.UUID]*UserCampaignStats), nil
	}

	// Преобразуем UUID в строки для SQL запроса
	campaignIDStrings := make([]string, len(campaignIDs))
	for i, id := range campaignIDs {
		campaignIDStrings[i] = fmt.Sprintf("toUUID('%s')", id.String())
	}

	query := `
		SELECT
			campaign_id,
			countIf(type = 'impression') as impressions_count,
			countIf(type = 'click') as clicks_count,
			max(is_viewed_by_user) as is_viewed,
			max(is_clicked_by_user) as is_clicked
		FROM
		(
			-- Показы
			SELECT
				campaign_id,
				'impression' as type,
				1 as is_viewed_by_user,
				0 as is_clicked_by_user
			FROM ad_impressions
			WHERE campaign_id IN (%s)
				AND client_id = toUUID(?)
			
			UNION ALL
			
			-- Клики
			SELECT
				campaign_id,
				'click' as type,
				0 as is_viewed_by_user,
				1 as is_clicked_by_user
			FROM ad_clicks
			WHERE campaign_id IN (%s)
				AND client_id = toUUID(?)
			
			UNION ALL
			
			-- Общая статистика показов
			SELECT
				campaign_id,
				'impression' as type,
				0 as is_viewed_by_user,
				0 as is_clicked_by_user
			FROM ad_impressions
			WHERE campaign_id IN (%s)
			
			UNION ALL
			
			-- Общая статистика кликов
			SELECT
				campaign_id,
				'click' as type,
				0 as is_viewed_by_user,
				0 as is_clicked_by_user
			FROM ad_clicks
			WHERE campaign_id IN (%s)
		)
		GROUP BY campaign_id
	`

	// Форматируем запрос, подставляя списки UUID
	campaignIDsStr := strings.Join(campaignIDStrings, ", ")
	formattedQuery := fmt.Sprintf(query, campaignIDsStr, campaignIDsStr, campaignIDsStr, campaignIDsStr)

	// Выполняем запрос с параметрами для client_id
	args := []interface{}{
		userID.String(), // Для проверки показов пользователю
		userID.String(), // Для проверки кликов пользователя
	}

	// Выполняем запрос
	rows, err := r.conn.Query(ctx, formattedQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	// Создаем мапу для результатов
	results := make(map[uuid.UUID]*UserCampaignStats)

	// Инициализируем статистику для всех кампаний
	for _, campaignID := range campaignIDs {
		results[campaignID] = &UserCampaignStats{
			CampaignID: campaignID,
		}
	}

	// Читаем результаты
	for rows.Next() {
		var (
			campaignIDStr    string
			impressionsCount uint64
			clicksCount      uint64
			isViewed         bool
			isClicked        bool
		)

		if err := rows.Scan(&campaignIDStr, &impressionsCount, &clicksCount, &isViewed, &isClicked); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		campaignID, err := uuid.Parse(campaignIDStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse campaign ID: %w", err)
		}

		// Обновляем статистику
		stats := results[campaignID]
		if stats != nil {
			stats.ImpressionsCount = impressionsCount
			stats.ClicksCount = clicksCount
			stats.IsViewedByUser = isViewed
			stats.IsClickedByUser = isClicked
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

// GetCampaignsSortedByUserViews возвращает кампании, сгруппированные по количеству просмотров
func (r *Repository) GetCampaignsSortedByUserViews(ctx context.Context, campaignIDs []uuid.UUID, userID uuid.UUID) ([]ViewsGroup, error) {
	if len(campaignIDs) == 0 {
		return []ViewsGroup{}, nil
	}

	// Преобразуем слайс ID в строку для SQL IN clause
	campaignIDStrings := make([]string, len(campaignIDs))
	for i, id := range campaignIDs {
		campaignIDStrings[i] = fmt.Sprintf("'%s'", id)
	}
	campaignsStr := strings.Join(campaignIDStrings, ",")

	// Получаем количество просмотров для каждой кампании
	query := fmt.Sprintf(`
		SELECT 
			campaign_id,
			sum(view_count) as total_views
		FROM ad_impressions FINAL
		WHERE campaign_id IN (%s)
			AND client_id = ?
		GROUP BY campaign_id
		ORDER BY total_views ASC
	`, campaignsStr)

	rows, err := r.conn.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get campaigns sorted by views: %w", err)
	}
	defer rows.Close()

	// Группируем кампании по количеству просмотров
	viewsMap := make(map[uint64][]uuid.UUID)
	seenCampaigns := make(map[uuid.UUID]struct{})

	for rows.Next() {
		var campaignID uuid.UUID
		var views uint64
		if err := rows.Scan(&campaignID, &views); err != nil {
			return nil, fmt.Errorf("failed to scan campaign views: %w", err)
		}
		viewsMap[views] = append(viewsMap[views], campaignID)
		seenCampaigns[campaignID] = struct{}{}
	}

	// Добавляем кампании без просмотров
	var notViewedCampaigns []uuid.UUID
	for _, id := range campaignIDs {
		if _, exists := seenCampaigns[id]; !exists {
			notViewedCampaigns = append(notViewedCampaigns, id)
		}
	}

	// Формируем результат, начиная с кампаний без просмотров
	result := make([]ViewsGroup, 0)
	if len(notViewedCampaigns) > 0 {
		result = append(result, ViewsGroup{
			ViewCount: 0,
			Campaigns: notViewedCampaigns,
		})
	}

	// Добавляем остальные группы в порядке возрастания просмотров
	var viewCounts []uint64
	for views := range viewsMap {
		viewCounts = append(viewCounts, views)
	}
	sort.Slice(viewCounts, func(i, j int) bool {
		return viewCounts[i] < viewCounts[j]
	})

	for _, views := range viewCounts {
		result = append(result, ViewsGroup{
			ViewCount: views,
			Campaigns: viewsMap[views],
		})
	}

	return result, nil
}

func (r *Repository) Close() error {
	return r.conn.Close()
}
