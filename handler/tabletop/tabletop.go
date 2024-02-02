package tabletop_handlers

import (
	"fajurion.com/voice-node/caching"
	"github.com/Fajurion/pipes"
	"github.com/Fajurion/pipes/send"
	"github.com/Fajurion/pipesfiber/wshandler"
)

func SetupHandler() {

	// Table member management
	wshandler.Routes["table_join"] = joinTable
	wshandler.Routes["table_leave"] = leaveTable

	// Table object management
	wshandler.Routes["tobj_create"] = createObject
	wshandler.Routes["tobj_delete"] = deleteObject
	wshandler.Routes["tobj_modify"] = modifyObject
	wshandler.Routes["tobj_move"] = moveObject

	// Table cursor sending
	wshandler.Routes["tc_move"] = moveCursor
}

// Send an event to all table members
func SendEventToMembers(room string, event pipes.Event) bool {
	valid, members := caching.TableMembers(room)
	if !valid {
		return false
	}

	return send.Pipe(send.ProtocolWS, pipes.Message{
		Channel: pipes.BroadcastChannel(members),
		Local:   true,
		Event:   event,
	}) == nil
}

// Send an event to all table members (excluding the sender)
func SendEventToMembersNoSelf(sender string, room string, event pipes.Event) bool {
	valid, members := caching.TableMembers(room)
	if !valid {
		return false
	}

	// Remove the sender from the list
	for i, member := range members {
		if member == sender {
			members = append(members[:i], members[i+1:]...)
			break
		}
	}

	return send.Pipe(send.ProtocolWS, pipes.Message{
		Channel: pipes.BroadcastChannel(members),
		Local:   true,
		Event:   event,
	}) == nil
}
