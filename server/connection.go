package server

import (
	"errors"

	"fajurion.com/voice-node/caching"
	"fajurion.com/voice-node/util"
	"github.com/Fajurion/pipes"
	"github.com/Fajurion/pipes/send"
)

// RefreshChannel is a channel that refreshes the connection of a user
func RefreshChannel(c *Context) error {

	if caching.LastConnectionRefresh(c.Addr.String()).Seconds() > caching.UserTTL.Seconds()/2 {
		return errors.New("too many refreshes")
	}

	caching.RefreshConnection(c.Addr.String())
	caching.RefreshUser(c.Account)
	util.Log.Println("[udp] Refreshed connection for", c.Account)

	send.ClientUDP(c.Account, pipes.Event{
		Name:   "refresh",
		Sender: "0",
		Data:   nil,
	})

	return nil
}

// ConfirmChannel is a channel that confirms the connection of a user
// TODO: Ratelimit this
func ConfirmChannel(c *Context) error {

	client, valid := caching.GetUser(c.Account)
	if !valid {
		return errors.New("user not found")
	}

	util.Log.Println("[udp] Confirmed connection for", c.Account)
	SendConfirmation(c.Addr.String(), c.Account, &client.Key)

	return nil
}
