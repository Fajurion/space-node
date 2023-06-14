package server

import (
	"errors"
	"net"

	"fajurion.com/voice-node/caching"
	"fajurion.com/voice-node/util"
)

var Channels map[byte]func(*Context) error = map[byte]func(*Context) error{}

type Context struct {
	Account string
	Bytes   []byte
	Addr    net.Addr
}

func ExecuteChannel(account string, bytes []byte, addr net.Addr) error {

	prefix := bytes[0]
	channel := Channels[prefix]
	if channel == nil {
		return errors.New("channel not found")
	}

	if len(bytes) < 2 {
		return errors.New("invalid packet")
	}

	var context = Context{
		Account: account,
		Bytes:   bytes[2:],
		Addr:    addr,
	}

	err := channel(&context)
	if err != nil {
		return err
	}

	return nil
}

func SetupChannels() {
	Channels['r'] = RefreshChannel
}

// Create a refresh channel function
func RefreshChannel(c *Context) error {

	if caching.LastConnectionRefresh(c.Addr.String()).Seconds() > caching.UserTTL.Seconds()/2 {
		return errors.New("too many refreshes")
	}

	caching.RefreshConnection(c.Addr.String())
	caching.RefreshUser(c.Account)
	util.Log.Println("[udp] Refreshed connection for", c.Account)

	return nil
}
