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
