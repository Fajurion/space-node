package server

import (
	"encoding/base64"
	"errors"

	"fajurion.com/voice-node/caching"
	"github.com/Fajurion/pipes"
	"github.com/Fajurion/pipes/send"
)

func VoiceChannel(c *Context) error {

	client, valid := caching.GetUser(c.Account)
	if !valid {
		return errors.New("user not found")
	}

	if client.TargetType != caching.TargetRoom {
		return errors.New("not implemented")
	}

	room, valid := caching.GetRoom(client.Target)
	if !valid {
		return errors.New("room not found")
	}

	// Check if voice is a reasonable size (1kb)
	if len(c.Bytes) > 1024 {
		return errors.New("voice too large")
	}

	// Encode voice data
	voiceString := base64.StdEncoding.EncodeToString(c.Bytes)

	send.Pipe("udp", pipes.Message{
		Local:   true,
		NoSelf:  true,
		Channel: pipes.Conversation(room.Members, []string{pipes.CurrentNode.ID}),
		Event: pipes.Event{
			Name:   "voice",
			Sender: client.ID,
			Data: map[string]interface{}{
				"data": voiceString,
			},
		},
	})

	return nil
}
