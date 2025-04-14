package app

import (
	"encoding/json"
	"github.com/labstack/echo/v4"
	"nlypage-final/pkg/closer"
	"sync"
)

// App is an interface that represents the app
type App interface {
	Start()
}

// app is a struct that represents the app
type app struct {
	serviceProvider ServiceProvider
}

// New is a function that creates a new app struct
func New() App {
	return &app{
		serviceProvider: newServiceProvider(),
	}
}

// route is a function that sets up the routes
func (a *app) route() {
	e := a.serviceProvider.Echo()

	a.serviceProvider.Logger().Debug("Setting up routes")
	e.GET("/ping", func(c echo.Context) error {
		return c.JSON(200, map[string]interface{}{
			"status": "ok",
		})
	})
	a.serviceProvider.ClientsHandler().Setup(e.Group("/clients"))
	a.serviceProvider.AdvertisersHandler().Setup(e.Group("/advertisers"))
	a.serviceProvider.MlScoreHandler().Setup(e.Group("/ml-scores"))
	a.serviceProvider.CampaignsHandler().Setup(e.Group("/advertisers"))
	a.serviceProvider.AdsHandler().Setup(e.Group("/ads"))
	a.serviceProvider.StatsHandler().Setup(e.Group("/stats"))
	a.serviceProvider.AiHandler().Setup(e.Group("/ai"))
	a.serviceProvider.TimeHandler().Setup(e.Group("/time"))
	a.serviceProvider.ModerationHandler().Setup(e.Group("/moderation"))

	routes, err := json.MarshalIndent(e.Routes(), "", "  ")
	if err == nil {
		a.serviceProvider.Logger().Debugf("Routes: %s", routes)
	}
}

// Start is a function that starts the app
func (a *app) Start() {
	// Close all resources on stop
	defer func() {
		closer.CloseAll()
		closer.Wait()
	}()

	a.serviceProvider.Logger().Info("Starting app")

	// Set up routes
	a.route()

	// Migrate database trough DI
	a.serviceProvider.DB()

	// Migrate clickhouse database trough DI
	a.serviceProvider.Clickhouse()

	// Start server listening
	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()

		a.serviceProvider.Logger().Info("Starting echo server")
		a.serviceProvider.Logger().Error(
			a.serviceProvider.Echo().
				Start(":" + a.serviceProvider.Viper().GetString("service.backend.port")),
		)
	}()

	//go func() {
	//	defer wg.Done()
	//	closer.Add(a.serviceProvider.AdScoringService().Stop)
	//
	//	a.serviceProvider.Logger().Info("Starting ad-scoring service")
	//	a.serviceProvider.Logger().Error(
	//		a.serviceProvider.AdScoringService().Start(context.Background()),
	//	)
	//}()

	go func() {
		defer wg.Done()

		a.serviceProvider.Logger().Info("Starting bot")
		a.serviceProvider.Bot().Start()
	}()

	wg.Wait()
}
