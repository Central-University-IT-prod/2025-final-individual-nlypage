package telegram

import tele "gopkg.in/telebot.v3"

type Handler interface {
	Setup(g *tele.Group)
}
