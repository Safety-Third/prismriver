package queue

import (
	"encoding/binary"
	"net"
	"net/http"
	"strconv"

	"github.com/sirupsen/logrus"

	"gitlab.com/ttpcodes/prismriver/internal/app/db"
	"gitlab.com/ttpcodes/prismriver/internal/app/downloader"
	"gitlab.com/ttpcodes/prismriver/internal/app/player"
)

// StoreHandler handles requests for adding new QueueItems.
func StoreHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		logrus.Warnf("error parsing form data from POST /queue: %v", err)
		return
	}
	id := r.Form.Get("id")
	kind := r.Form.Get("type")
	url := r.Form.Get("url")
	video, err := strconv.ParseBool(r.Form.Get("video"))
	if err != nil {
		logrus.Warnf("error parsing boolean from video input, defaulting to false")
		video = false
	}

	var ip uint32 = 0
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		if parsed := net.ParseIP(host); parsed != nil {
			if len(parsed) == 16 {
				ip = binary.BigEndian.Uint32(parsed[12:16])
			} else {
				ip = binary.BigEndian.Uint32(parsed)
			}
		}
	}

	if len(id) > 0 && len(kind) > 0 {
		media, err := db.GetMedia(id, kind)
		if err == nil {
			queue := player.GetQueue()
			queue.Add(media, ip)
			return
		}
	}
	if len(url) != 0 {
		id, name, source := downloader.FindSource(url)
		if id != "" {
			queue := player.GetQueue()
			media, err := db.GetMedia(id, name)
			if err == nil {
				queue.Add(media, ip)
				return
			}
			media, err = source.GetInfo(id, video)
			if err != nil {
				logrus.Errorf("could not get video info: %v", err)
				return
			}
			if err := db.AddMedia(media); err != nil {
				logrus.Errorf("error storing new media item; %v", err)
				return
			}
			queue.Add(media, ip)
			return
		}
		logrus.Warnf("client attempted to add unsupported media %v, ignoring", url)
		return
	}
	logrus.Warn("User sent an empty POST request, ignoring.")
}
