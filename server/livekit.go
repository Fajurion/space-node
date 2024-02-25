package server

import (
	"context"
	"log"
	"os"

	"github.com/livekit/protocol/auth"
	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go"
)

var RoomClient *lksdk.RoomServiceClient

func InitLiveKit() {
	RoomClient = lksdk.NewRoomServiceClient(os.Getenv("LK_URL"), os.Getenv("LK_KEY"), os.Getenv("LK_SECRET"))

	room, err := RoomClient.CreateRoom(context.Background(), &livekit.CreateRoomRequest{
		Name:         "test",
		EmptyTimeout: 60,
	})
	if err != nil {
		panic(err)
	}

	token := RoomClient.CreateToken()
	token.AddGrant(&auth.VideoGrant{
		RoomJoin: true,
		Room:     room.Name,
	})
	token.SetName("test 1")
	token.SetIdentity("test 1")

	tokenString, err := token.ToJWT()
	if err != nil {
		panic(err)
	}

	log.Println(tokenString)

	token = RoomClient.CreateToken()
	token.AddGrant(&auth.VideoGrant{
		RoomJoin: true,
		Room:     room.Name,
	})
	token.SetName("test 2")
	token.SetIdentity("test 2")

	tokenString, err = token.ToJWT()
	if err != nil {
		panic(err)
	}

	log.Println(tokenString)
}
