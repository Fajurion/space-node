package games

import (
	"github.com/Fajurion/pipesfiber/wshandler"
)

func SetupActions() {
	wshandler.Routes["game_init"] = initGame
	wshandler.Routes["game_event"] = gameEvent
}
