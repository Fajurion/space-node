package tabletop_handlers

import (
	"fajurion.com/voice-node/caching"
	"fajurion.com/voice-node/util"
	"github.com/Fajurion/pipes"
	"github.com/Fajurion/pipesfiber/wshandler"
)

// Action: tobj_create
func createObject(message wshandler.Message) {

	if message.ValidateForm("x", "y", "w", "h", "r", "type", "data") {
		wshandler.ErrorResponse(message, "invalid")
		return
	}

	connection, valid := caching.GetConnection(message.Client.ID)
	if !valid {
		wshandler.ErrorResponse(message, "invalid")
		return
	}

	x := message.Data["x"].(float64)
	y := message.Data["y"].(float64)
	width := message.Data["w"].(float64)
	height := message.Data["h"].(float64)
	rotation := message.Data["r"].(float64)
	objType := int(message.Data["type"].(float64))
	objData := message.Data["data"].(string)

	object := &caching.TableObject{
		LocationX: x,
		LocationY: y,
		Width:     width,
		Height:    height,
		Rotation:  rotation,
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
	valid = SendEventToMembers(message.Client.Session, pipes.Event{
		Name: "tobj_created",
		Data: map[string]interface{}{
			"id":   object.ID,
			"x":    x,
			"y":    y,
			"w":    width,
			"h":    height,
			"r":    rotation,
			"type": objType,
			"data": objData,
			"c":    connection.ClientID,
		},
	})
	if !valid {
		wshandler.ErrorResponse(message, "server.error")
		return
	}

	wshandler.NormalResponse(message, map[string]interface{}{
		"success": true,
		"id":      object.ID,
	})
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

// Action: tobj_select
func selectObject(message wshandler.Message) {

	if message.ValidateForm("id") {
		wshandler.ErrorResponse(message, "invalid")
		return
	}

	connection, valid := caching.GetConnection(message.Client.ID)
	if !valid {
		wshandler.ErrorResponse(message, "invalid")
		return
	}

	// Grab hold of it
	err := caching.SelectTableObject(message.Client.Session, message.Data["id"].(string), connection.ClientID)
	if err != nil {
		wshandler.ErrorResponse(message, util.ErrorTabletopInvalidAction)
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

	connection, valid := caching.GetConnection(message.Client.ID)
	if !valid {
		wshandler.ErrorResponse(message, "invalid")
		return
	}

	// Check if the object is held by the client
	obj, valid := caching.GetTableObject(message.Client.Session, message.Data["id"].(string))
	if !valid {
		wshandler.ErrorResponse(message, "invalid")
		return
	}
	if obj.Holder != connection.ClientID {
		wshandler.ErrorResponse(message, util.ErrorTabletopInvalidAction)
		return
	}

	err := caching.ModifyTableObject(message.Client.Session, message.Data["id"].(string), message.Data["data"].(string))
	if err != nil {
		wshandler.ErrorResponse(message, "server.error")
		return
	}

	// Notify other clients
	valid = SendEventToMembers(message.Client.Session, pipes.Event{
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

// Action: tobj_rotate
func rotateObject(message wshandler.Message) {

	if message.ValidateForm("id", "r") {
		wshandler.ErrorResponse(message, "invalid")
		return
	}

	connection, valid := caching.GetConnection(message.Client.ID)
	if !valid {
		wshandler.ErrorResponse(message, "invalid")
		return
	}

	rotation := message.Data["r"].(float64)

	err := caching.RotateTableObject(message.Client.Session, message.Data["id"].(string), rotation)
	if err != nil {
		wshandler.ErrorResponse(message, "server.error")
		return
	}

	// Notify other clients
	valid = SendEventToMembers(message.Client.Session, pipes.Event{
		Name: "tobj_rotated",
		Data: map[string]interface{}{
			"id": message.Data["id"].(string),
			"s":  connection.ClientID,
			"r":  rotation,
		},
	})
	if !valid {
		wshandler.ErrorResponse(message, "server.error")
		return
	}

	wshandler.SuccessResponse(message)
}
