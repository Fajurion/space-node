package caching

import "github.com/Fajurion/pipesfiber"

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

var games = map[string]Game{
	"chess": {},
}

func OpenGameSession(Room string, Game string) {

}
