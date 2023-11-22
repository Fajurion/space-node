package handler

import (
	games_actions "fajurion.com/voice-node/handler/games"
	"github.com/Fajurion/pipesfiber/wshandler"
)

func Initialize() {
	wshandler.Initialize()
	games_actions.SetupActions()

	wshandler.Routes["set_data"] = setData
	wshandler.Routes["setup"] = setupUDP
	wshandler.Routes["update"] = update
}
