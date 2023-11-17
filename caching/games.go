package caching

import (
	"fajurion.com/voice-node/util"
	"github.com/Fajurion/pipesfiber"
	"github.com/dgraph-io/ristretto"
)

type Game struct {
	Id         string
	Session    string
	LaunchFunc func() chan EventContext
}

type EventContext struct {
	Client  *pipesfiber.Client
	Name    string
	Session string
	Data    map[string]interface{}
}

type GameSession struct {
	Id           string
	Game         string
	Members      []string // Client ids
	EventChannel *chan EventContext
}

// ! For setting please ALWAYS use cost 1
var sessionsCache *ristretto.Cache

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

var games = map[string]Game{
	"chess": {
		Id:      "chess",
		Session: "chess",
		LaunchFunc: func() chan EventContext {
			channel := make(chan EventContext)
			go func() {
				for {
					event := <-channel
					if event.Name == "close" {
						break
					}
					util.Log.Println(event)
				}
			}()
			return channel
		},
	},
}

func OpenGameSession(roomId string, gameId string) (string, bool) {

	game, ok := games[gameId]
	if !ok {
		return "", false
	}
	room, valid := GetRoom(roomId)
	if !valid {
		return "", false
	}
	room.Mutex.Lock()

	room, valid = GetRoom(roomId)
	if !valid {
		return "", false
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
		Id:           sessionId,
		Game:         gameId,
		EventChannel: &channel,
	}

	room.Sessions = append(room.Sessions, session.Id)
	roomsCache.Set(roomId, room, 1)
	sessionsCache.Set(sessionId, session, 1)

	roomsCache.Wait()
	room.Mutex.Unlock()

	return sessionId, true
}

func ForwardGameEvent(sessionId string, event EventContext) bool {

	session, valid := sessionsCache.Get(sessionId)
	if !valid {
		return false
	}

	*session.(GameSession).EventChannel <- event

	return true
}
