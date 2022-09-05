package routes

import (
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/Safety-Third/prismriver/internal/app/player"
	"github.com/Safety-Third/prismriver/internal/app/server/ws"
	"net/http"
	"sync"
)

var queueHub *ws.Hub
var queueOnce sync.Once

var queueUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// GetQueueHub returns the single Hub instance used for handling Queue-related WebSocket requests.
func GetQueueHub() *ws.Hub {
	queueOnce.Do(func() {
		queueHub = ws.CreateHub()
		go queueHub.Execute()

		go (func() {
			queue := player.GetQueue()
			for response := range queue.Update {
				queueHub.Broadcast <- response
			}
		})()
	})
	return queueHub
}

// WebsocketQueueHandler handles requests for getting Queue WebSocket updates.
func WebsocketQueueHandler(w http.ResponseWriter, r *http.Request) {
	GetQueueHub()
	conn, err := queueUpgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.Error("Error when upgrading client to WS connection:")
		logrus.Error(err)
		return
	}
	client := &ws.Client{
		Conn: conn,
		Hub:  queueHub,
		Send: make(chan []byte, 256),
	}
	client.Hub.Register <- client

	go client.RunRead()
	go client.RunWrite()

	queue := player.GetQueue()
	response, err := queue.List()
	if err != nil {
		logrus.Errorf("error generating queue response: %v", err)
		return
	}
	client.Send <- response
	logrus.Debug("Sent initial message on WS connection.")
}
