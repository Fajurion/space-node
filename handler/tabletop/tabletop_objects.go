package tabletop_handlers

import (
	"fajurion.com/voice-node/caching"
	"github.com/Fajurion/pipes"
	"github.com/Fajurion/pipesfiber/wshandler"
)

// Action: tobj_create
func createObject(message wshandler.Message) {

	if message.ValidateForm("x", "y", "w", "h", "type", "data") {
		wshandler.ErrorResponse(message, "invalid")
		return
	}

	x := message.Data["x"].(float64)
	y := message.Data["y"].(float64)
	width := message.Data["w"].(float64)
	height := message.Data["h"].(float64)
	objType := int(message.Data["type"].(float64))
	objData := message.Data["data"].(string)

	object := &caching.TableObject{
		LocationX: x,
		LocationY: y,
		Width:     width,
		Height:    height,
		Type:      objType,
		Data:      objData,
		Creator:   message.Client.ID,
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
			"x":    x,
			"y":    y,
			"w":    width,
			"h":    height,
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

	if message.ValidateForm("id", "x", "y") {
		wshandler.ErrorResponse(message, "invalid")
		return
	}

	x := message.Data["x"].(float64)
	y := message.Data["y"].(float64)

	err := caching.MoveTableObject(message.Client.Session, message.Data["id"].(string), x, y)
	if err != nil {
		wshandler.ErrorResponse(message, "server.error")
		return
	}

	// Notify other clients
	valid := SendEventToMembers(message.Client.Session, pipes.Event{
		Name: "tobj_moved",
		Data: map[string]interface{}{
			"id": message.Data["id"].(string),
			"s":  message.Client.ID,
			"x":  x,
			"y":  y,
		},
	})
	if !valid {
		wshandler.ErrorResponse(message, "server.error")
		return
	}

	wshandler.SuccessResponse(message)
}