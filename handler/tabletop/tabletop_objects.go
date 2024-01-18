package tabletop_handlers

import (
	"fajurion.com/voice-node/caching"
	"github.com/Fajurion/pipes"
	"github.com/Fajurion/pipesfiber/wshandler"
)

// Action: tobj_create
func createObject(message wshandler.Message) {

	if message.ValidateForm("location", "type", "data") {
		wshandler.ErrorResponse(message, "invalid")
		return
	}

	location := message.Data["location"].(string)
	objType := message.Data["type"].(string)
	objData := message.Data["data"].(string)

	object := &caching.TableObject{
		Location: location,
		Type:     objType,
		Data:     objData,
		Creator:  message.Client.ID,
	}
	err := caching.AddObjectToTable(message.Client.Session, object)
	if err != nil {
		wshandler.ErrorResponse(message, "server.error")
		return
	}

	// Notify other clients
	valid := SendEventToMembers(message.Client.Session, pipes.Event{
		Name: "tobj_created",
		Data: map[string]interface{}{
			"id":   object.ID,
			"loc":  location,
			"type": objType,
			"data": objData,
		},
	})
	if !valid {
		wshandler.ErrorResponse(message, "server.error")
		return
	}

	wshandler.SuccessResponse(message)
}

// Action: tobj_delete
func deleteObject(message wshandler.Message) {

	if message.ValidateForm("id") {
		wshandler.ErrorResponse(message, "invalid")
		return
	}

	err := caching.RemoveObjectFromTable(message.Client.Session, message.Data["id"].(string))
	if err != nil {
		wshandler.ErrorResponse(message, "server.error")
		return
	}

	// Notify other clients
	valid := SendEventToMembers(message.Client.Session, pipes.Event{
		Name: "tobj_deleted",
		Data: map[string]interface{}{
			"id": message.Data["id"].(string),
		},
	})
	if !valid {
		wshandler.ErrorResponse(message, "server.error")
		return
	}

	wshandler.SuccessResponse(message)
}

// Action: tobj_modify
func modifyObject(message wshandler.Message) {

	if message.ValidateForm("id", "data") {
		wshandler.ErrorResponse(message, "invalid")
		return
	}

	err := caching.ModifyTableObject(message.Client.Session, message.Data["id"].(string), message.Data["data"].(string))
	if err != nil {
		wshandler.ErrorResponse(message, "server.error")
		return
	}

	// Notify other clients
	valid := SendEventToMembers(message.Client.Session, pipes.Event{
		Name: "tobj_modified",
		Data: map[string]interface{}{
			"id":   message.Data["id"].(string),
			"data": message.Data["data"].(string),
		},
	})
	if !valid {
		wshandler.ErrorResponse(message, "server.error")
		return
	}

	wshandler.SuccessResponse(message)
}

// Action: tobj_move
func moveObject(message wshandler.Message) {

	if message.ValidateForm("id", "location") {
		wshandler.ErrorResponse(message, "invalid")
		return
	}

	err := caching.ModifyTableObject(message.Client.Session, message.Data["id"].(string), message.Data["location"].(string))
	if err != nil {
		wshandler.ErrorResponse(message, "server.error")
		return
	}

	// Notify other clients
	valid := SendEventToMembers(message.Client.Session, pipes.Event{
		Name: "tobj_modified",
		Data: map[string]interface{}{
			"id":   message.Data["id"].(string),
			"data": message.Data["location"].(string),
		},
	})
	if !valid {
		wshandler.ErrorResponse(message, "server.error")
		return
	}

	wshandler.SuccessResponse(message)
}
