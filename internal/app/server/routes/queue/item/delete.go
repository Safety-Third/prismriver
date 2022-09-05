package item

import (
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/Safety-Third/prismriver/internal/app/player"
	"net/http"
	"strconv"
)

// DeleteHandler handles requests for deleting QueueItems.
func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	index, err := strconv.ParseUint(vars["id"], 10, 8)
	if err != nil {
		logrus.Warn("Error parsing int in DeleteHandler, user likely provided incorrect input.")
		return
	}
	queue := player.GetQueue()
	queue.Remove(int(index))
}
