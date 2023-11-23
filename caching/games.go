package caching

import (
	"fajurion.com/voice-node/caching/games"
	"fajurion.com/voice-node/util"
	"github.com/dgraph-io/ristretto"
)

// ! For setting please ALWAYS use cost 1
var sessionsCache *ristretto.Cache

var GamesMap = map[string]games.Game{}

func setupSessionsCache() {
	var err error
	sessionsCache, err = ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e5,     // expecting to store 10k sessions
		MaxCost:     1 << 30, // maximum cost of cache is 1GB
		BufferItems: 64,      // Some random number, check docs
		OnEvict: func(item *ristretto.Item) {
			session := item.Value.(games.GameSession)

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

	*session.(games.GameSession).EventChannel <- games.EventContext{
		Name: "close",
	}

	sessionsCache.Del(sessionId)
	return true
}

func OpenGameSession(connId string, clientId string, roomId string, gameId string) (games.GameSession, bool) {

	game, ok := GamesMap[gameId]
	if !ok {
		return games.GameSession{}, false
	}
	room, valid := GetRoom(roomId)
	if !valid {
		return games.GameSession{}, false
	}
	room.Mutex.Lock()

	room, valid = GetRoom(roomId)
	if !valid {
		return games.GameSession{}, false
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

	channel := game.LaunchFunc(sessionId)
	session := games.GameSession{
		Id:            sessionId,
		Game:          gameId,
		GameState:     games.GameStateLobby,
		EventChannel:  &channel,
		Creator:       connId,
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

func StartGameSession(sessionId string) bool {

	obj, valid := sessionsCache.Get(sessionId)
	if !valid {
		return false
	}
	session := obj.(games.GameSession)
	if session.GameState > games.GameStateLobby {
		return false
	}

	*session.EventChannel <- games.EventContext{
		Name: "start",
	}

	return true
}

func SetGameState(sessionId string, state int) (games.GameSession, bool) {

	obj, valid := sessionsCache.Get(sessionId)
	if !valid {
		return games.GameSession{}, false
	}
	session := obj.(games.GameSession)

	session.GameState = state
	sessionsCache.Set(sessionId, session, 1)

	return session, true
}

func ForwardGameEvent(sessionId string, event games.EventContext) bool {

	session, valid := sessionsCache.Get(sessionId)
	if !valid {
		return false
	}

	*session.(games.GameSession).EventChannel <- event

	return true
}

func GetSession(sessionId string) (games.GameSession, bool) {
	session, valid := sessionsCache.Get(sessionId)
	if !valid {
		return games.GameSession{}, false
	}
	return session.(games.GameSession), valid
}
