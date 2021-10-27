package item

import (
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"gitlab.com/ttpcodes/prismriver/internal/app/player"
	"net/http"
	"strconv"
)

// UpdateHandler handles requests for moving QueueItems around in the Queue.
func UpdateHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	index, err := strconv.ParseUint(vars["id"], 10, 8)
	if err != nil {
		logrus.Warn("Error parsing int in UpdateHandler, user likely provided incorrect input.")
		return
	}
	queue := player.GetQueue()
	r.ParseForm()
	move := r.Form.Get("move")
	switch move {
	case "down":
		queue.MoveUp(int(index))
	case "up":
		queue.MoveDown(int(index))
		}
	}
}
