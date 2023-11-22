package wordgrid

import (
	"fajurion.com/voice-node/caching/games"
	"fajurion.com/voice-node/util"
)

func LaunchWordGrid() chan games.EventContext {
	channel := make(chan games.EventContext)
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
}
