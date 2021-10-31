package queue

import (
	"github.com/sirupsen/logrus"
	"gitlab.com/ttpcodes/prismriver/internal/app/player"
	"net/http"
)

// IndexHandler handles requests for listing all QueueItems.
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	queue := player.GetQueue()
	response, err := queue.List()
	if err != nil {
		logrus.Errorf("error generating queue response: %v", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}
