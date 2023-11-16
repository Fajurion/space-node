package games

import (
	"github.com/Fajurion/pipesfiber"
	"github.com/Fajurion/pipesfiber/wshandler"
)

type Game struct {
	Id       string
	Session  string
	Handlers map[string]func(*EventContext)
}

type EventContext struct {
	Client  *pipesfiber.Client
	Name    string
	Session string
	Data    map[string]interface{}
}

func SetupActions() {
	wshandler.Routes["game_init"] = nil
	wshandler.Routes["game_event"] = nil
}
