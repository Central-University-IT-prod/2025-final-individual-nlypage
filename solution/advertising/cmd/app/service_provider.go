package app

import (
	"ariga.io/entcache"
	"context"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"errors"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/nlypage/intele"
	"github.com/spf13/viper"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/layout"
	"log"
	"nlypage-final/internal/adapters/config"
	apiV1 "nlypage-final/internal/adapters/controller/api/v1"
	"nlypage-final/internal/adapters/controller/api/v1/ads"
	"nlypage-final/internal/adapters/controller/api/v1/advertisers"
	"nlypage-final/internal/adapters/controller/api/v1/ai"
	"nlypage-final/internal/adapters/controller/api/v1/campaigns"
	"nlypage-final/internal/adapters/controller/api/v1/clients"
	"nlypage-final/internal/adapters/controller/api/v1/ml_score"
	"nlypage-final/internal/adapters/controller/api/v1/moderation"
	"nlypage-final/internal/adapters/controller/api/v1/stats"
	timeHandler "nlypage-final/internal/adapters/controller/api/v1/time"
	"nlypage-final/internal/adapters/controller/api/validator"
	"nlypage-final/internal/adapters/database/clickhouse"
	"nlypage-final/internal/adapters/database/minio"
	"nlypage-final/internal/adapters/database/postgres/ent"
	"nlypage-final/internal/adapters/database/redis"
	"nlypage-final/internal/domain/dto"
	"nlypage-final/internal/domain/service"
	"nlypage-final/pkg/ad_scoring"
	"nlypage-final/pkg/closer"
	"nlypage-final/pkg/gigachat"
	"nlypage-final/pkg/logger"
	"os"
	"time"

	_ "github.com/lib/pq"
)

// ServiceProvider is a simple DI container
type ServiceProvider interface {
	Echo() *echo.Echo
	Bot() *tele.Bot

	Viper() *viper.Viper
	Location() *time.Location
	Layout() *layout.Layout
	InputManager() *intele.InputManager

	LoggerConfig() config.LoggerConfig
	PGConfig() config.PGConfig
	ClickhouseConfig() config.ClickHouseConfig
	GigaChatConfig() config.GigachatConfig
	AdScoringConfig() config.AdScoringConfig

	Validator() *validator.Validator
	Logger() *logger.Logger
	GigaChat() *gigachat.Client

	DB() *ent.Client
	Redis() *redis.Client
	Clickhouse() *clickhouse.Repository
	AdImagesRepository() minio.AdImagesRepository
	AdScorer() ad_scoring.Scorer

	TimeService() service.TimeService
	ClientService() service.ClientService
	AdvertiserService() service.AdvertiserService
	MlScoreService() service.MlScoreService
	CampaignService() service.CampaignService
	AdService() service.AdService
	StatsService() service.StatsService
	GenerateService() service.GenerateService
	ModerationService() service.ModerationService
	AdScoringService() service.AdScoringService

	TimeHandler() apiV1.Handler
	ClientsHandler() apiV1.Handler
	AdvertisersHandler() apiV1.Handler
	MlScoreHandler() apiV1.Handler
	CampaignsHandler() apiV1.Handler
	AdsHandler() apiV1.Handler
	StatsHandler() apiV1.Handler
	AiHandler() apiV1.Handler
	ModerationHandler() apiV1.Handler
}

type serviceProvider struct {
	echo *echo.Echo
	bot  *tele.Bot

	viper        *viper.Viper
	location     *time.Location
	layout       *layout.Layout
	inputManager *intele.InputManager

	pgConfig         config.PGConfig
	loggerConfig     config.LoggerConfig
	clickhouseConfig config.ClickHouseConfig
	gigachatConfig   config.GigachatConfig
	minioConfig      config.MinioConfig
	adScoringConfig  config.AdScoringConfig

	validator *validator.Validator
	logger    *logger.Logger
	gigachat  *gigachat.Client
	adScorer  ad_scoring.Scorer

	db                 *ent.Client
	clickhouse         *clickhouse.Repository
	redis              *redis.Client
	adImagesRepository minio.AdImagesRepository

	timeService       service.TimeService
	clientService     service.ClientService
	advertiserService service.AdvertiserService
	mlScoreService    service.MlScoreService
	campaignService   service.CampaignService
	adService         service.AdService
	statsService      service.StatsService
	generateService   service.GenerateService
	moderationService service.ModerationService
	adScoringService  service.AdScoringService

	timeHandler        apiV1.Handler
	clientsHandler     apiV1.Handler
	advertisersHandler apiV1.Handler
	mlScoreHandler     apiV1.Handler
	campaignsHandler   apiV1.Handler
	adsHandler         apiV1.Handler
	statsHandler       apiV1.Handler
	aiHandler          apiV1.Handler
	moderationHandler  apiV1.Handler
}

func newServiceProvider() ServiceProvider {
	return &serviceProvider{}
}

func (s *serviceProvider) Echo() *echo.Echo {
	if s.echo == nil {
		e := echo.New()
		e.HTTPErrorHandler = func(err error, c echo.Context) {
			code := echo.ErrBadRequest.Code
			var he *echo.HTTPError
			if errors.As(err, &he) {
				code = he.Code
			}

			if !c.Response().Committed {
				_ = c.JSON(code, dto.HTTPStatus{
					Code:    code,
					Message: err.Error(),
				})
			}
		}

		apiV1Logger := s.Logger().Named("api_v1")
		e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
			LogURI:      true,
			LogStatus:   true,
			LogMethod:   true,
			HandleError: true,
			LogError:    true,
			LogRemoteIP: true,
			LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
				if v.Error == nil {
					apiV1Logger.Infow("request completed",
						"ip", v.RemoteIP,
						"method", v.Method,
						"uri", v.URI,
						"status", v.Status,
					)
				} else {
					apiV1Logger.Errorw("request failed",
						"ip", v.RemoteIP,
						"method", v.Method,
						"uri", v.URI,
						"status", v.Status,
						"error", v.Error.Error(),
					)
				}
				return nil
			},
		}))

		s.echo = e
	}
	return s.echo
}

func (s *serviceProvider) Bot() *tele.Bot {
	if s.bot == nil {
		if err := os.Setenv("BOT_TOKEN", s.Viper().GetString("service.bot.token")); err != nil {
			s.Logger().Panicf("failed to set bot token: %v", err)
		}

		settings := s.Layout().Settings()
		botLogger := s.Logger().Named("bot")
		settings.OnError = func(err error, ctx tele.Context) {
			if ctx.Callback() == nil {
				botLogger.Errorf("(user: %d) | Error: %v", ctx.Sender().ID, err)
			} else {
				botLogger.Errorf("(user: %d) | unique: %s | Error: %v", ctx.Sender().ID, ctx.Callback().Unique, err)
			}
		}

		b, err := tele.NewBot(settings)
		if err != nil {
			botLogger.Panicf("failed to init bot: %v", err)
		}

		if cmds := s.Layout().Commands(); cmds != nil {
			if err = b.SetCommands(cmds); err != nil {
				botLogger.Panicf("failed to set commands: %v", err)
			}
		}

		s.bot = b
	}
	return s.bot
}

func (s *serviceProvider) Viper() *viper.Viper {
	if s.viper == nil {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")

		if err := viper.ReadInConfig(); err != nil {
			log.Panicf("failed to read config: %v", err)
		}
		s.viper = viper.GetViper()
	}

	return s.viper
}

func (s *serviceProvider) Location() *time.Location {
	if s.location == nil {
		location, err := time.LoadLocation(viper.GetString("settings.timezone"))
		if err != nil {
			s.Logger().Panicf("failed to load location: %v", err)
		}
		s.location = location
	}

	return s.location
}

func (s *serviceProvider) Layout() *layout.Layout {
	if s.layout == nil {
		l, err := layout.New("telegram.yml")
		if err != nil {
			s.Logger().Panicf("failed to init layout: %v", err)
		}
		s.layout = l
	}

	return s.layout
}

func (s *serviceProvider) InputManager() *intele.InputManager {
	if s.inputManager == nil {
		s.inputManager = intele.NewInputManager(intele.InputOptions{
			Storage: s.Redis().States,
		})
	}
	return s.inputManager
}

func (s *serviceProvider) LoggerConfig() config.LoggerConfig {
	if s.loggerConfig == nil {
		cfg := config.NewLoggerConfig(s.Viper(), s.Location())
		s.loggerConfig = cfg
	}

	return s.loggerConfig
}

func (s *serviceProvider) PGConfig() config.PGConfig {
	if s.pgConfig == nil {
		s.pgConfig = config.NewPGConfig(s.Viper())
	}

	return s.pgConfig
}

func (s *serviceProvider) ClickhouseConfig() config.ClickHouseConfig {
	if s.clickhouseConfig == nil {
		s.clickhouseConfig = config.NewClickHouseConfig(s.Viper())
	}

	return s.clickhouseConfig
}

func (s *serviceProvider) GigaChatConfig() config.GigachatConfig {
	if s.gigachatConfig == nil {
		s.gigachatConfig = config.NewGigachatConfig(s.Viper())
	}

	return s.gigachatConfig
}

func (s *serviceProvider) AdScoringConfig() config.AdScoringConfig {
	if s.adScoringConfig == nil {
		s.adScoringConfig = config.NewAdScoringConfig(s.Viper())
	}

	return s.adScoringConfig
}

func (s *serviceProvider) MinioConfig() config.MinioConfig {
	if s.minioConfig == nil {
		s.minioConfig = config.NewMinioConfig(s.Viper())
	}

	return s.minioConfig
}

func (s *serviceProvider) Logger() *logger.Logger {
	if s.logger == nil {
		l, err := logger.Init(logger.Config{
			Debug:        s.LoggerConfig().Debug(),
			TimeLocation: s.LoggerConfig().TimeLocation(),
			LogToFile:    s.LoggerConfig().LogToFile(),
			LogsDir:      s.LoggerConfig().LogsDir(),
		})
		if err != nil {
			s.Logger().Panicf("failed to init logger: %v", err)
		}

		s.logger = l
		if s.LoggerConfig().Debug() {
			s.logger.Debug("Debug mode enabled")
		}
	}

	return s.logger
}

func (s *serviceProvider) DB() *ent.Client {
	if s.db == nil {
		db, err := sql.Open(dialect.Postgres, s.PGConfig().DSN())
		if err != nil {
			s.Logger().Panicf("failed to open database: %v", err)
		}

		drv := entcache.NewDriver(
			db,
			entcache.TTL(time.Second*3),
			entcache.Levels(
				entcache.NewLRU(256),
				entcache.NewRedis(s.Redis().Cache),
			),
		)
		client := ent.NewClient(ent.Driver(drv))

		loggerCfg := s.LoggerConfig()
		if loggerCfg.Debug() {
			client = client.Debug()
		}
		if errMigrate := client.Schema.Create(entcache.Skip(context.Background())); errMigrate != nil {
			s.Logger().Panicf("Failed to run migrations: %v", errMigrate)
		}

		closer.Add(client.Close)
		s.db = client
	}

	return s.db
}

func (s *serviceProvider) Redis() *redis.Client {
	if s.redis == nil {
		r, err := redis.New(redis.Options{
			Host:     s.Viper().GetString("service.redis.host"),
			Port:     s.Viper().GetString("service.redis.port"),
			Password: s.Viper().GetString("service.redis.password"),
		})
		if err != nil {
			s.Logger().Panicf("failed to init redis: %v", err)
		}

		closer.Add(r.CloseAll)
		s.redis = r
	}
	return s.redis
}

func (s *serviceProvider) Clickhouse() *clickhouse.Repository {
	if s.clickhouse == nil {
		cfg := s.ClickhouseConfig()
		var err error
		s.clickhouse, err = clickhouse.New(clickhouse.Config{
			Host:     cfg.Host(),
			Port:     cfg.Port(),
			Database: cfg.Database(),
			Username: cfg.Username(),
			Password: cfg.Password(),
			Debug:    cfg.Debug(),
		})
		if err != nil {
			s.Logger().Panicf("failed to init clickhouse: %v", err)
		}

		closer.Add(s.clickhouse.Close)
	}
	return s.clickhouse
}

func (s *serviceProvider) AdImagesRepository() minio.AdImagesRepository {
	if s.adImagesRepository == nil {
		imagesRepository, err := minio.NewAdImagesRepository(minio.Config{
			Endpoint:     s.MinioConfig().Endpoint(),
			HTTPEndpoint: s.MinioConfig().HTTPEndpoint(),
			AccessKey:    s.MinioConfig().AccessKey(),
			SecretKey:    s.MinioConfig().SecretKey(),
			BucketName:   s.MinioConfig().BucketName(),
			UseSSL:       s.MinioConfig().UseSSL(),
		})
		if err != nil {
			s.Logger().Panicf("failed to init ad images repository: %v", err)
		}
		s.adImagesRepository = imagesRepository
	}
	return s.adImagesRepository
}

func (s *serviceProvider) AdScorer() ad_scoring.Scorer {
	if s.adScorer == nil {
		s.adScorer = ad_scoring.NewScorer(ad_scoring.Config{
			PlatformProfitWeight: s.AdScoringConfig().PlatformProfitWeight(),
			RelevanceWeight:      s.AdScoringConfig().RelevanceWeight(),
			PerformanceWeight:    s.AdScoringConfig().PerformanceWeight(),
		})
	}
	return s.adScorer
}

func (s *serviceProvider) Validator() *validator.Validator {
	if s.validator == nil {
		s.validator = validator.New()
	}

	return s.validator
}

func (s *serviceProvider) GigaChat() *gigachat.Client {
	if s.gigachat == nil {
		giga, err := gigachat.NewInsecureClientWithAuthKey(s.GigaChatConfig().AuthKey())
		if err != nil {
			s.Logger().Panicf("failed to init gigachat: %v", err)
		}
		s.gigachat = giga
	}
	return s.gigachat
}

// ----------------------------------Services----------------------------------start

func (s *serviceProvider) TimeService() service.TimeService {
	if s.timeService == nil {
		timeSrvc, err := service.NewTimeService(s.Redis().Time)
		if err != nil {
			s.Logger().Panicf("failed to init time service: %v", err)
		}
		s.timeService = timeSrvc

	}
	return s.timeService
}

func (s *serviceProvider) ClientService() service.ClientService {
	if s.clientService == nil {
		s.clientService = service.NewClientService(s.DB())
	}
	return s.clientService
}

func (s *serviceProvider) AdvertiserService() service.AdvertiserService {
	if s.advertiserService == nil {
		s.advertiserService = service.NewAdvertiserService(s.DB())
	}
	return s.advertiserService
}

func (s *serviceProvider) MlScoreService() service.MlScoreService {
	if s.mlScoreService == nil {
		s.mlScoreService = service.NewMlScoreService(s.DB())
	}
	return s.mlScoreService
}

func (s *serviceProvider) CampaignService() service.CampaignService {
	if s.campaignService == nil {
		s.campaignService = service.NewCampaignService(
			s.DB(),
			s.TimeService(),
			s.Clickhouse(),
			s.AdImagesRepository(),
			s.Viper().GetBool("service.backend.settings.campaign-moderation"),
		)
	}
	return s.campaignService
}

func (s *serviceProvider) AdService() service.AdService {
	if s.adService == nil {
		s.adService = service.NewAdService(
			s.DB(),
			s.AdScorer(),
			s.Redis().Ads,
			s.Clickhouse(),
			s.TimeService(),
		)
	}
	return s.adService
}

func (s *serviceProvider) StatsService() service.StatsService {
	if s.statsService == nil {
		s.statsService = service.NewStatsService(s.TimeService(), s.Clickhouse())
	}
	return s.statsService
}

func (s *serviceProvider) GenerateService() service.GenerateService {
	if s.generateService == nil {
		s.generateService = service.NewGenerateService(s.GigaChat(), s.AdvertiserService())
	}
	return s.generateService
}

func (s *serviceProvider) ModerationService() service.ModerationService {
	if s.moderationService == nil {
		s.moderationService = service.NewModerationService(s.DB())
	}
	return s.moderationService
}

func (s *serviceProvider) AdScoringService() service.AdScoringService {
	if s.adScoringService == nil {
		s.adScoringService = service.NewAdScoringService(
			s.TimeService(),
			s.DB(),
			s.Clickhouse(),
			s.AdScorer(),
			s.Redis().Ads,
			s.Logger().Named("ad-scoring"),
			s.AdScoringConfig().UpdateInterval(),
		)
	}
	return s.adScoringService
}

// ----------------------------------Services----------------------------------end

// ----------------------------------Handlers----------------------------------start

func (s *serviceProvider) TimeHandler() apiV1.Handler {
	if s.timeHandler == nil {
		s.timeHandler = timeHandler.NewTimeHandler(s.TimeService(), s.AdScoringService(), s.Validator())
	}
	return s.timeHandler
}

func (s *serviceProvider) ClientsHandler() apiV1.Handler {
	if s.clientsHandler == nil {
		s.clientsHandler = clients.NewClientsHandler(s.ClientService(), s.Validator())
	}
	return s.clientsHandler
}

func (s *serviceProvider) AdvertisersHandler() apiV1.Handler {
	if s.advertisersHandler == nil {
		s.advertisersHandler = advertisers.NewAdvertisersHandler(s.AdvertiserService(), s.Validator())
	}
	return s.advertisersHandler
}

func (s *serviceProvider) MlScoreHandler() apiV1.Handler {
	if s.mlScoreHandler == nil {
		s.mlScoreHandler = ml_score.NewMlScoreHandler(s.MlScoreService(), s.Validator())
	}
	return s.mlScoreHandler
}

func (s *serviceProvider) CampaignsHandler() apiV1.Handler {
	if s.campaignsHandler == nil {
		s.campaignsHandler = campaigns.NewCampaignsHandler(s.CampaignService(), s.Validator())
	}
	return s.campaignsHandler
}

func (s *serviceProvider) AdsHandler() apiV1.Handler {
	if s.adsHandler == nil {
		s.adsHandler = ads.NewAdsHandler(s.AdService(), s.Validator())
	}
	return s.adsHandler
}

func (s *serviceProvider) StatsHandler() apiV1.Handler {
	if s.statsHandler == nil {
		s.statsHandler = stats.NewStatsHandler(s.StatsService(), s.Validator())
	}
	return s.statsHandler
}

func (s *serviceProvider) AiHandler() apiV1.Handler {
	if s.aiHandler == nil {
		s.aiHandler = ai.NewAiHandler(s.GenerateService(), s.Validator())
	}
	return s.aiHandler
}

func (s *serviceProvider) ModerationHandler() apiV1.Handler {
	if s.moderationHandler == nil {
		s.moderationHandler = moderation.NewModerationHandler(s.ModerationService(), s.Validator())
	}
	return s.moderationHandler
}

// ----------------------------------Handlers----------------------------------end
