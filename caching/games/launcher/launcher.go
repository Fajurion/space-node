package launcher

import (
	"fajurion.com/voice-node/caching"
	"fajurion.com/voice-node/caching/games"
	"fajurion.com/voice-node/caching/games/wordgrid"
)

func InitGames() {
	caching.GamesMap["wordgrid"] = games.Game{
		Id:         "wordgrid",
		LaunchFunc: LaunchWorldGrid,
		MinPlayers: 1,
		MaxPlayers: 100,
	}
}

func LaunchWorldGrid(session string) chan games.EventContext {
	return wordgrid.LaunchWordGrid(session)
}
