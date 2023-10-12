package handler

import (
	"fajurion.com/voice-node/caching"
	"github.com/Fajurion/pipes"
	"github.com/Fajurion/pipes/send"
	"github.com/Fajurion/pipesfiber/wshandler"
)

// Action: set_data
func setData(message wshandler.Message) {

	if message.ValidateForm("data") {
		wshandler.ErrorResponse(message, "invalid")
		return
	}

	// Set data
	valid := caching.SetRoomData(message.Client.Session, message.Data["data"].(string))
	if !valid {
		wshandler.ErrorResponse(message, "invalid")
		return
	}

	if !SendRoomData(message.Client.Session) {
		wshandler.ErrorResponse(message, "invalid")
		return
	}

	wshandler.SuccessResponse(message)
}

func SendRoomData(id string) bool {
	adapters, event, valid := GetRoomData(id, "room_data")
	if !valid {
		return false
	}

	// Send to all
	err := send.Pipe(send.ProtocolWS, pipes.Message{
		Channel: pipes.BroadcastChannel(adapters),
		Local:   true,
		Event:   event,
	})
	return err == nil
}

func GetRoomData(id string, eventName string) ([]string, pipes.Event, bool) {
	room, validRoom := caching.GetRoom(id)
	members, valid := caching.GetAllConnections(id)
	if !valid || !validRoom {
		return []string{}, pipes.Event{}, false
	}

	// Get all members
	adapters := make([]string, len(members))
	returnableMembers := make([]caching.ReturnableMember, len(members))
	i := 0
	for _, member := range members {
		returnableMembers[i] = member.ToReturnableMember()
		adapters[i] = member.Adapter
		i++
	}

	// Send to all
	return adapters, pipes.Event{
		Name: eventName,
		Data: map[string]interface{}{
			"start":   room.Start,
			"room":    room.Data,
			"members": returnableMembers,
		},
	}, true
}
