package player

import (
	"encoding/json"
	"math/rand"
	"os"
	"path"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"gitlab.com/ttpcodes/prismriver/internal/app/constants"
	"gitlab.com/ttpcodes/prismriver/internal/app/db"
	"gitlab.com/ttpcodes/prismriver/internal/app/downloader"
)

var queueInstance *Queue
var queueOnce sync.Once

// Download represents a download occurring for a QueueItem.
type Download struct {
	doneCh   chan struct{}
	err      string
	progress int
}

// Queue represents a queue of Media items waiting to be played.
type Queue struct {
	sync.RWMutex

	balancing bool
	downloads map[string]*Download
	items     []*QueueItem
	Update    chan []byte
}

// QueueItem represents a Media item waiting to be played in the Queue.
type QueueItem struct {
	balanced bool
	err      string
	// go doesn't have a method for returning a random generic uint for some reason
	id       uint32
	Media    db.Media
	owner    uint32
	ready    chan bool
	queue    *Queue
}

// QueueResponse represents a Queue containing the necessary fields to be exported via JSON.
type QueueResponse struct {
	Balancing bool                `json:"balancing"`
	Items     []QueueItemResponse `json:"items"`
}

// QueueItemResponse represents a QueueItem containing the necessary fields to be exported via JSON.
type QueueItemResponse struct {
	Downloading bool     `json:"downloading"`
	Error       string   `json:"error"`
	Id          uint32   `json:"id"`
	Media       db.Media `json:"media"`
	Progress    int      `json:"progress"`
}

// GetQueue returns the single Queue instance of the application.
func GetQueue() *Queue {
	queueOnce.Do(func() {
		logrus.Info("Created queue instance.")
		queueInstance = &Queue{
			balancing: true,
			downloads: make(map[string]*Download),
			items:     make([]*QueueItem, 0),
			Update:    make(chan []byte),
		}
	})
	return queueInstance
}

// Add adds a new Media item to the Queue as a QueueItem. If the item is detected to not be ready, it will instantiate
// a download of the Media. Add is thread-safe.
func (q *Queue) Add(media db.Media, owner uint32) {
	q.Lock()
	var id uint32 = 0
	for q.contains(id) {
		id = rand.Uint32()
	}
	item := &QueueItem{
		balanced: q.balancing,
		id:       id,
		Media:    media,
		owner:    owner,
		ready:    make(chan bool),
		queue:    q,
	}
	if q.balancing {
		q.items = InsertQueueItemBalanced(item, q.items)
	} else {
		q.items = InsertQueueItemDefault(item, q.items)
	}

	dataDir := viper.GetString(constants.DATA)
	source := downloader.GetSource(media.Type)
	ext := ".opus"
	if media.Video && source.HasVideo() {
		if viper.GetBool(constants.VIDEOTRANSCODING) {
			ext = ".mp4"
		} else {
			ext = ".video"
		}
	}
	filePath := path.Join(dataDir, media.Type, media.ID+ext)
	_, err := os.Stat(filePath)
	download, ok := q.downloads[item.Media.ID]
	if item.Media.Type != "internal" && (os.IsNotExist(err) || ok) {
		if ok {
			go func() {
				<-download.doneCh
				if download.err != "" {
					q.Lock()
					item.err = download.err
					q.Unlock()
					q.sendQueueUpdate()
					return
				}
				item.ready <- true
				close(item.ready)
			}()
		} else {
			download := &Download{
				doneCh: make(chan struct{}),
			}
			q.downloads[item.Media.ID] = download

			progressChan, doneChan, err := source.DownloadMedia(media)
			if err != nil {
				logrus.Errorf("error when downloading media: %v", err)
				return
			}

			go func() {
				for progress := range progressChan {
					q.Lock()
					download.progress = int(progress)
					q.Unlock()
					q.sendQueueUpdate()
				}
				if err := <-doneChan; err != nil {
					q.Lock()
					download.err = err.Error()
					item.err = err.Error()
					delete(q.downloads, item.Media.ID)
					close(download.doneCh)
					q.Unlock()
					q.sendQueueUpdate()
					return
				}
				q.Lock()
				delete(q.downloads, item.Media.ID)
				close(download.doneCh)
				q.Unlock()
				q.sendQueueUpdate()
				item.ready <- true
				close(item.ready)
			}()
		}
	} else {
		logrus.Debug("Queue item ready. Sending on channel.")
		go func() {
			item.ready <- true
			close(item.ready)
		}()
	}
	player := GetPlayer()
	if player.State == STOPPED && len(q.items) == 1 {
		go player.Play(item)
	}
	q.Unlock()
	q.sendQueueUpdate()
	logrus.Info("Added " + media.Title + " to queue.")
}

// Advance moves the Queue up by one and plays the next item if it exists. Advance is thread-safe.
func (q *Queue) Advance() {
	q.Lock()
	q.items = q.items[1:]
	if len(q.items) > 0 {
		player := GetPlayer()
		go player.Play(q.items[0])
	}
	q.Unlock()
	q.sendQueueUpdate()
}

// BeQuiet replaces the currently playing item with the BeQuiet Media and plays it. BeQuiet is thread-safe.
func (q *Queue) BeQuiet() {
	player := GetPlayer()
	q.Lock()
	defer q.Unlock()
	if len(q.items) == 0 {
		q.Add(*db.BeQuiet, 0)
		q.Unlock()
		return
	} else if player.State == LOADING {
		return
	}
	quietQueue := make([]*QueueItem, 0)
	quietItem := &QueueItem{
		Media: *db.BeQuiet,
		ready: make(chan bool, 1),
		queue: q,
	}
	quietItem.ready <- true
	close(quietItem.ready)
	quietQueue = append(quietQueue, q.items[0], quietItem)
	quietQueue = append(quietQueue, q.items[1:]...)
	q.items = quietQueue
	player.Skip()
}

// MoveTo moves a QueueItem to a specific position in the Queue. MoveTo is thread-safe.
func (q *Queue) MoveTo(index int, to int) {
	if index == 0 || index == to || to == 0 {
		return
	}
	q.Lock()
	if index < len(q.items) && to < len(q.items) {
		item := q.items[index]
		q.items = append(q.items[:index], q.items[index + 1:]...)
		if to == len(q.items) {
			q.items = append(q.items, item)
		} else {
			q.items = append(q.items[:to], q.items[to - 1:]...)
			q.items[to] = item
		}
		item.balanced = false
	} else {
		logrus.Warnf("user provided invalid request to move item at index %v to new index %v", index, to)
	}
	q.Unlock()
	q.sendQueueUpdate()
}

// GenerateResponse generates a JSON response of all the QueueItems in the Queue. GenerateResponse is thread-safe.
func (q *Queue) GenerateResponse() []byte {
	// Cannot return a nil slice or the frontend will have issues.
	items := make([]QueueItemResponse, 0)
	q.RLock()
	for _, item := range q.items {
		response := item.GenerateResponse()
		items = append(items, response)
	}
	wrapper := QueueResponse{
		Balancing: q.balancing,
		Items:     items,
	}
	q.RUnlock()
	response, err := json.Marshal(wrapper)
	if err != nil {
		logrus.Error("Error generating JSON response:")
		logrus.Error(err)
	}
	return response
}

// Remove removes a QueueItem from the Queue. Remove is thread-safe. Note that slice indices must be integers in Go.
func (q *Queue) Remove(index int) {
	q.Lock()
	if index < len(q.items) {
		q.items = append(q.items[:index], q.items[index+1:]...)
	} else {
		logrus.Infof("user attempted to remove now nonexistent queue item at index %v, ignoring", index)
	}
	q.Unlock()
	q.sendQueueUpdate()
}

// SetBalancing turns on and off balancing queue ordering. SetBalancing is thread-safe.
func (q *Queue) SetBalancing(balancing bool) {
	q.Lock()
	q.balancing = balancing
	if q.balancing {
		orderedItems := make([]*QueueItem, 0)
		for _, item := range q.items {
			item.balanced = true
			orderedItems = InsertQueueItemBalanced(item, orderedItems)
		}
		q.items = orderedItems
	} else {
		for _, item := range q.items {
			item.balanced = false
		}
	}
	q.Unlock()
	q.sendQueueUpdate()
}

func (q *Queue) contains(id uint32) bool {
	for _, item := range q.items {
		if item.id == id {
			return true
		}
	}
	return false
}

// sendQueueUpdate is thread-safe.
func (q *Queue) sendQueueUpdate() {
	response := q.GenerateResponse()
	q.Update <- response
}

// GenerateResponse returns the QueueItemResponse form of the QueueItem.
func (q QueueItem) GenerateResponse() QueueItemResponse {
	downloading, progress := q.Progress()
	return QueueItemResponse{
		Downloading: downloading,
		Error:       q.err,
		Id:          q.id,
		Media:       q.Media,
		Progress:    progress,
	}
}

// Progress returns the download progress of the QueueItem.
func (q QueueItem) Progress() (bool, int) {
	download, ok := q.queue.downloads[q.Media.ID]
	if !ok {
		return false, 100
	}
	return true, download.progress
}

// Shuffle performs a shuffle on the items in the Queue. Shuffle is thread-safe.
func (q *Queue) Shuffle() {
	q.Lock()
	if len(q.items) > 1 {
		// Offset by one since we don't want to modify the currently playing item.
		rand.Shuffle(len(q.items)-1, func(i, j int) {
			q.items[i+1], q.items[j+1] = q.items[j+1], q.items[i+1]
		})
		q.Unlock()
		q.sendQueueUpdate()
	} else {
		q.Unlock()
	}
}
