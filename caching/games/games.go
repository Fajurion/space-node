package games

import "github.com/Fajurion/pipesfiber"

type Game struct {
	Id         string
	LaunchFunc func(string) chan EventContext

	// Lobby settings
	MaxPlayers int
	MinPlayers int
}

type EventContext struct {
	Client *pipesfiber.Client
	Name   string
	Data   interface{}
}

var (
	GameStateLobby = 1
)

type GameSession struct {
	Id            string
	Game          string
	GameState     int
	Creator       string // Connection ID
	ConnectionIds []string
	ClientIds     []string
	EventChannel  *chan EventContext
}
