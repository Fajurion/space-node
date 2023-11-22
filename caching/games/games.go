package games

import "github.com/Fajurion/pipesfiber"

type Game struct {
	Id         string
	LaunchFunc func() chan EventContext
}

type EventContext struct {
	Client  *pipesfiber.Client
	Name    string
	Session string
	Data    interface{}
}
