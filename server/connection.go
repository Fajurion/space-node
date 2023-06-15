package server

import (
	"errors"

	"fajurion.com/voice-node/caching"
	"fajurion.com/voice-node/util"
)

// RefreshChannel is a channel that refreshes the connection of a user
func RefreshChannel(c *Context) error {

	if caching.LastConnectionRefresh(c.Addr.String()).Seconds() > caching.UserTTL.Seconds()/2 {
		return errors.New("too many refreshes")
	}

	caching.RefreshConnection(c.Addr.String())
	caching.RefreshUser(c.Account)
	util.Log.Println("[udp] Refreshed connection for", c.Account)

	return nil
}

// ConfirmChannel is a channel that confirms the connection of a user
func ConfirmChannel(c *Context) error {

	if caching.LastConnectionRefresh(c.Addr.String()).Seconds() > caching.UserTTL.Seconds()/2 {
		return errors.New("too many refreshes")
	}

	caching.RefreshConnection(c.Addr.String())
	caching.RefreshUser(c.Account)
	util.Log.Println("[udp] Confirmed connection for", c.Account)

	return nil
}
