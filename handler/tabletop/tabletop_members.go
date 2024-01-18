package tabletop_handlers

import (
	"log"

	"fajurion.com/voice-node/caching"
	"github.com/Fajurion/pipes"
	"github.com/Fajurion/pipesfiber/wshandler"
)

// Action: table_join
func joinTable(message wshandler.Message) {

	err := caching.JoinTable(message.Client.Session, message.Client.ID)
	if err != nil {
		log.Println("Couldn't join table of room", message.Client.Session, ":", err.Error())
		wshandler.ErrorResponse(message, "server.error")
		return
	}

	wshandler.SuccessResponse(message)

	// Send all the objects
	objects, err := caching.TableObjects(message.Client.Session)
	if err != nil {
		log.Println("Couldn't get objects of room", message.Client.Session, ":", err.Error())
		return
	}

	err = message.Client.SendEvent(pipes.Event{
		Name: "table_obj",
		Data: map[string]interface{}{
			"obj": objects,
		},
	})
	if err != nil {
		log.Println("Couldn't send objects of room through event", message.Client.Session, ":", err.Error())
	}
}

// Action: table_leave
func leaveTable(message wshandler.Message) {
	err := caching.LeaveTable(message.Client.Session, message.Client.ID)
	if err != nil {
		log.Println("Couldn't leave table of room", message.Client.Session, ":", err.Error())
		wshandler.ErrorResponse(message, "server.error")
		return
	}

	wshandler.SuccessResponse(message)
}
