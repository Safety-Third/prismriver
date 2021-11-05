package media

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"gitlab.com/ttpcodes/prismriver/internal/app/db"
	"net/http"
	"strconv"
)

type indexResponse struct {
	Media []db.Media `json:"media"`
	Pages uint `json:"pages"`
}

// IndexHandler handles requests to list Media in the database.
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	strParam := params.Get("limit")
	limit, err := strconv.ParseUint(strParam, 10, 8)
	if err != nil {
		logrus.Infof("could not parse %v as limit, defaulting to 12", strParam)
		limit = 12
	}
	query := params.Get("query")
	var response indexResponse
	if query == "" {
		response = indexResponse{
			Media: db.GetRandomMedia(int(limit)),
			Pages: 1,
		}
	} else {
		pageParam := params.Get("page")
		page, err := strconv.ParseUint(pageParam, 10, 8)
		if err != nil {
			logrus.Infof("could not parse %v as page, defaulting to 1", pageParam)
			page = 1
		}
		media, pages := db.FindMedia(query, int(limit), int(page))
		response = indexResponse{
			Media: media,
			Pages: pages,
		}
	}
	data, err := json.Marshal(response)
	if err != nil {
		logrus.Errorf("could not generate media index response: %v", err)
		return
	}
	w.Write(data)
}
