package media

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"gitlab.com/ttpcodes/prismriver/internal/app/db"
	"gitlab.com/ttpcodes/prismriver/internal/app/downloader"
	"net/http"
	"strconv"
)

// UpdateHandler handles requests for updating Media items.
func UpdateHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	media, err := db.GetMedia(vars["id"], vars["type"])
	if err != nil {
		message := fmt.Sprintf("could not find media with id %v and type %v", vars["id"], vars["type"])
		logrus.Infof(message)
		http.Error(w, message, http.StatusNotFound)
		return
	}
	r.ParseForm()
	video := r.Form.Get("video")
	boolean, err := strconv.ParseBool(video)
	if err != nil {
		message := fmt.Sprintf("could not parse %v as bool", video)
		logrus.Warnf(message)
		http.Error(w, message, http.StatusBadRequest)
		return
	}
	if media.Video == boolean {
		logrus.Warnf("media with id %v and type %v already has video set to %v, ignoring", vars["id"], vars["type"], boolean)
		w.WriteHeader(http.StatusNotModified)
		return
	}
	if boolean {
		newMedia, err := downloader.GetInfo(media.URL, boolean)
		if err != nil {
			message := fmt.Sprintf("could not get info for media with url %v: %v", media.URL, err)
			logrus.Errorf(message)
			http.Error(w, message, http.StatusInternalServerError)
			return
		}
		if !newMedia.Video {
			message := fmt.Sprintf("media with url %v does not support video, ignoring", media.URL)
			logrus.Infof(message)
			http.Error(w, message, http.StatusForbidden)
			return
		}
	}
	media.Video = boolean
	media.Save()
	w.WriteHeader(http.StatusNoContent)
}
