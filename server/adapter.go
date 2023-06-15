package server

import (
	"crypto/cipher"
	"errors"
	"net"

	"fajurion.com/voice-node/caching"
	"github.com/Fajurion/pipes/adapter"
	pipesUtil "github.com/Fajurion/pipes/util"
)

// AddAdapter adds an adapter for pipes
func AddAdapter(client caching.ConnectedClient) error {

	adapter.AdaptUDP(adapter.Adapter{
		ID:      client.ID,
		Receive: sendFromAdapter,
		Data:    &client,
	})

	return nil
}

// Handles receiving messages from the adapter and passes them to the client
func sendFromAdapter(context *adapter.Context) error {
	client := context.Adapter.Data.(*caching.ConnectedClient)
	return SendToIP(client.Address, &client.Key, context.Message)
}

// SendToClient sends a message to a client through UDP
func SendToClient(account string, bytes []byte) error {

	client, valid := caching.GetUser(account)
	if !valid {
		return errors.New("user not found")
	}

	return SendToIP(client.Address, &client.Key, bytes)
}

// SendToIP sends a message to a address through UDP
func SendToIP(address string, key *cipher.Block, bytes []byte) error {

	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return err
	}

	// Encrypt using AES
	bytes, err = pipesUtil.EncryptAES(*key, bytes)
	if err != nil {
		return err
	}

	_, err = udpServ.WriteToUDP(bytes, addr)
	return err
}
