package caching

import (
	"fajurion.com/voice-node/caching/games"
	"fajurion.com/voice-node/caching/games/wordgrid"
	"fajurion.com/voice-node/util"
	"github.com/dgraph-io/ristretto"
)

type GameSession struct {
	Id            string
	Game          string
	ConnectionIds []string
	ClientIds     []string
	EventChannel  *chan games.EventContext
}

// ! For setting please ALWAYS use cost 1
var sessionsCache *ristretto.Cache

var gamesMap = map[string]games.Game{
	"grid": {
		Id:         "grid",
		LaunchFunc: wordgrid.LaunchWordGrid,
	},
}

func setupSessionsCache() {
	var err error
	sessionsCache, err = ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e5,     // expecting to store 10k sessions
		MaxCost:     1 << 30, // maximum cost of cache is 1GB
		BufferItems: 64,      // Some random number, check docs
		OnEvict: func(item *ristretto.Item) {
			session := item.Value.(GameSession)

			util.Log.Println("[cache] session", session.Id, "was deleted")
		},
	})

	if err != nil {
		panic(err)
	}
}

func CloseSession(sessionId string) bool {

	session, valid := sessionsCache.Get(sessionId)
	if !valid {
		return false
	}

	*session.(GameSession).EventChannel <- games.EventContext{
		Name: "close",
	}

	sessionsCache.Del(sessionId)
	return true
}

func OpenGameSession(connId string, clientId string, roomId string, gameId string) (GameSession, bool) {

	game, ok := gamesMap[gameId]
	if !ok {
		return GameSession{}, false
	}
	room, valid := GetRoom(roomId)
	if !valid {
		return GameSession{}, false
	}
	room.Mutex.Lock()

	room, valid = GetRoom(roomId)
	if !valid {
		return GameSession{}, false
	}

	// Create game session
	sessionId := util.GenerateToken(12)
	for {
		_, ok := sessionsCache.Get(sessionId)
		if !ok {
			break
		}
		sessionId = util.GenerateToken(12)
	}

	channel := game.LaunchFunc()
	session := GameSession{
		Id:            sessionId,
		Game:          gameId,
		EventChannel:  &channel,
		ConnectionIds: []string{connId},
		ClientIds:     []string{clientId},
	}

	room.Sessions = append(room.Sessions, session.Id)
	roomsCache.Set(roomId, room, 1)
	sessionsCache.Set(sessionId, session, 1)

	roomsCache.Wait()
	room.Mutex.Unlock()

	return session, true
}

func ForwardGameEvent(sessionId string, event games.EventContext) bool {

	session, valid := sessionsCache.Get(sessionId)
	if !valid {
		return false
	}

	*session.(GameSession).EventChannel <- event

	return true
}

func GetSession(sessionId string) (GameSession, bool) {
	session, valid := sessionsCache.Get(sessionId)
	if !valid {
		return GameSession{}, false
	}
	return session.(GameSession), valid
}
