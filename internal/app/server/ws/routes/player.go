package routes

import (
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"gitlab.com/ttpcodes/prismriver/internal/app/player"
	"gitlab.com/ttpcodes/prismriver/internal/app/server/ws"
	"net/http"
	"sync"
)

var playerHub *ws.Hub
var playerOnce sync.Once

var playerUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// GetPlayerHub returns the single Hub instance used for handling Player-related WebSocket requests.
func GetPlayerHub() *ws.Hub {
	playerOnce.Do(func() {
		playerHub = ws.CreateHub()
		go playerHub.Execute()

		go (func() {
			playerInstance := player.GetPlayer()
			for response := range playerInstance.Update {
				playerHub.Broadcast <- response
			}
		})()
	})
	return playerHub
}

// WebsocketPlayerHandler handles requests for getting Player WebSocket updates.
func WebsocketPlayerHandler(w http.ResponseWriter, r *http.Request) {
	GetPlayerHub()
	conn, err := playerUpgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.Error("Error when upgrading client to WS connection:")
		logrus.Error(err)
		return
	}
	client := &ws.Client{
		Conn: conn,
		Hub:  playerHub,
		Send: make(chan []byte, 256),
	}
	client.Hub.Register <- client

	go client.RunRead()
	go client.RunWrite()

	playerInstance := player.GetPlayer()
	response := playerInstance.GenerateResponse()
	client.Send <- response
	logrus.Debug("Sent initial message on WS connection.")
}
