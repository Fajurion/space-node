package handler

import "github.com/Fajurion/pipesfiber/wshandler"

func Initialize() {
	wshandler.Initialize()

	wshandler.Routes["set_data"] = setData
}
