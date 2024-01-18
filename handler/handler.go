package handler

import (
	games_actions "fajurion.com/voice-node/handler/games"
	tabletop_handlers "fajurion.com/voice-node/handler/tabletop"
	"github.com/Fajurion/pipesfiber/wshandler"
)

func Initialize() {
	wshandler.Initialize()
	games_actions.SetupActions()

	wshandler.Routes["set_data"] = setData
	wshandler.Routes["setup"] = setupUDP
	wshandler.Routes["update"] = update

	tabletop_handlers.SetupHandler()
}
