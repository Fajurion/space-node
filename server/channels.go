package server

import (
	"errors"
	"net"

	"fajurion.com/voice-node/util"
	"github.com/Fajurion/pipes"
)

var Channels map[byte]func(*Context) error = map[byte]func(*Context) error{}

type Context struct {
	Account  string
	ClientID string
	Bytes    []byte
	Addr     net.Addr
}

func ExecuteChannel(account string, clientID string, bytes []byte, addr net.Addr) error {

	prefix := bytes[0]
	channel := Channels[prefix]
	if channel == nil {
		return errors.New("channel not found")
	}

	if len(bytes) < 2 {
		return errors.New("invalid packet")
	}

	var context = Context{
		ClientID: clientID,
		Account:  account,
		Bytes:    bytes[2:],
		Addr:     addr,
	}

	err := channel(&context)
	if err != nil {
		return err
	}

	if pipes.DebugLogs {
		util.Log.Println("[udp] Executed channel", string(prefix), "for", account)
	}

	return nil
}

func SetupChannels() {
	Channels['r'] = RefreshChannel
	Channels['c'] = ConfirmChannel
	Channels['v'] = VoiceChannel
}
