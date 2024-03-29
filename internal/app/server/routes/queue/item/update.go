package item

import (
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/Safety-Third/prismriver/internal/app/player"
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
	case "bottom":
		queue.MoveTo(int(index), -1)
	case "down":
		queue.MoveTo(int(index), int(index+1))
	case "top":
		queue.MoveTo(int(index), 1)
	case "up":
		queue.MoveTo(int(index), int(index-1))
	default:
		to, err := strconv.ParseUint(move, 10, 8)
		if err != nil {
			logrus.Warnf("could not parse %v as a valid move instruction, ignoring request", move)
			return
		}
		queue.MoveTo(int(index), int(to))
	}
}
