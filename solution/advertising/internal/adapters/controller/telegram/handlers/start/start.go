package start

import (
	"nlypage-final/internal/adapters/controller/telegram"
	"nlypage-final/pkg/logger"

	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/layout"
)

type handler struct {
	layout *layout.Layout
	logger *logger.Logger
}

func New(lt *layout.Layout, logger *logger.Logger) telegram.Handler {
	return &handler{
		layout: lt,
		logger: logger,
	}
}

func (h handler) Start(c tele.Context) error {
	h.logger.Infof("(user: %d) enter /start", c.Sender().ID)
	return nil
}

func (h handler) Setup(g *tele.Group) {
	g.Handle("/start", h.Start)
}
