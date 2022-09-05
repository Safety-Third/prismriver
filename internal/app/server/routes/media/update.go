package media

import (
	"fmt"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/Safety-Third/prismriver/internal/app/db"
	"github.com/Safety-Third/prismriver/internal/app/downloader"
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
	str := r.Form.Get("video")
	video, videoErr := strconv.ParseBool(str)
	if videoErr != nil {
		logrus.Infof("could not parse %v as bool, ignoring", str)
	}
	str = r.Form.Get("length")
	length, err := strconv.ParseBool(str)
	if err != nil {
		logrus.Infof("could not parse %v as bool, ignoring", str)
	}
	str = r.Form.Get("title")
	title, err := strconv.ParseBool(str)
	if err != nil {
		logrus.Infof("could not parse %v as bool, ignoring", str)
	}
	modified := false
	if videoErr == nil && video != media.Video && video || length || title {
		newMedia, err := downloader.GetInfo(media.URL, video)
		if err != nil {
			message := fmt.Sprintf("could not get info for media with url %v: %v", media.URL, err)
			logrus.Errorf(message)
			http.Error(w, message, http.StatusInternalServerError)
			return
		}
		if videoErr == nil && video != media.Video && video && newMedia.Video != media.Video {
			media.Video = newMedia.Video
			modified = true
		}
		if length && newMedia.Length != media.Length {
			media.Length = newMedia.Length
			modified = true
		}
		if title && newMedia.Title != media.Title {
			media.Title = newMedia.Title
			modified = true
		}
	}
	if videoErr == nil && video != media.Video && !video {
		media.Video = video
		modified = true
	}
	if !modified {
		logrus.Infof("media with id %v and type %v has no fields to update, ignoring", vars["id"], vars["type"])
		w.WriteHeader(http.StatusNotModified)
		return
	}
	response, err := json.Marshal(media)
	if err != nil {
		message := fmt.Sprintf("could not generate media response: %v", err)
		logrus.Errorf(message)
		http.Error(w, message, http.StatusInternalServerError)
		return
	}
	media.Save()
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}
