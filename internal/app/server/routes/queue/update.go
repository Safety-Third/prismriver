package queue

import (
	"github.com/sirupsen/logrus"
	"github.com/Safety-Third/prismriver/internal/app/player"
	"net/http"
	"strconv"
)

// UpdateHandler handles requests for updating Queue instance settings.
func UpdateHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	balancing, err := strconv.ParseBool(r.Form.Get("balancing"))
	if err != nil {
		logrus.Warnf("error parsing boolean from balancing input, defaulting to true")
		balancing = true
	}
	queue := player.GetQueue()
	queue.SetBalancing(balancing)
}
