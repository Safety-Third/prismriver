package player

import (
	"context"
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

// DownloadKey represents a identifier for a specific Download via its id, type, and video status.
type DownloadKey struct {
	id        string
	mediaType string
	video     bool
}

// Queue represents a queue of Media items waiting to be played.
type Queue struct {
	sync.RWMutex

	balancing bool
	downloads map[DownloadKey]*Download
	items     []*QueueItem
	Update    chan []byte
}

// QueueItem represents a Media item waiting to be played in the Queue.
type QueueItem struct {
	balanced bool
	cancel   context.CancelFunc
	ctx      context.Context
	err      string
	// go doesn't have a method for returning a random generic uint for some reason
	id       uint32
	Media    db.Media
	owner    uint32
	ready    chan struct{}
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
			downloads: make(map[DownloadKey]*Download),
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
	defer q.Unlock()
	var id uint32 = 0
	for q.contains(id) {
		id = rand.Uint32()
	}
	ctx, cancel := context.WithCancel(context.Background())
	item := &QueueItem{
		balanced: q.balancing,
		cancel:   cancel,
		ctx:      ctx,
		id:       id,
		Media:    media,
		owner:    owner,
		ready:    make(chan struct{}),
		queue:    q,
	}
	if q.balancing {
		q.items = InsertQueueItemBalanced(item, q.items)
	} else {
		q.items = InsertQueueItemDefault(item, q.items)
	}

	dataDir := viper.GetString(constants.DATA)
	ext := ".opus"
	if media.Video {
		if viper.GetBool(constants.VIDEO_TRANSCODING) {
			ext = ".mp4"
		} else {
			ext = ".video"
		}
	}
	filePath := path.Join(dataDir, media.Type, media.ID+ext)
	_, err := os.Stat(filePath)
	key := DownloadKey{
		id:        item.Media.ID,
		mediaType: item.Media.Type,
		video:     item.Media.Video,
	}
	download, ok := q.downloads[key]
	if item.Media.Type != "internal" && (os.IsNotExist(err) || ok) {
		if ok {
			go func() {
				<-download.doneCh
				if download.err != "" {
					q.Lock()
					defer q.Unlock()
					item.err = download.err
					q.sendQueueUpdate()
					return
				}
				close(item.ready)
			}()
		} else {
			download := &Download{
				doneCh: make(chan struct{}),
			}
			q.downloads[key] = download

			progressChan, doneChan, err := downloader.DownloadMedia(media)
			if err != nil {
				logrus.Errorf("error when downloading media: %v", err)
				return
			}

			go func() {
				for progress := range progressChan {
					q.Lock()
					download.progress = int(progress)
					q.sendQueueUpdate()
					q.Unlock()
				}
				if err := <-doneChan; err != nil {
					q.Lock()
					defer q.Unlock()
					download.err = err.Error()
					item.err = err.Error()
					delete(q.downloads, key)
					close(download.doneCh)
					q.sendQueueUpdate()
					return
				}
				q.Lock()
				delete(q.downloads, key)
				close(download.doneCh)
				q.sendQueueUpdate()
				q.Unlock()
				close(item.ready)
			}()
		}
	} else {
		logrus.Debugf("queue item %v ready", item.id)
		go func() {
			close(item.ready)
		}()
	}
	player := GetPlayer()
	if player.State == STOPPED && len(q.items) == 1 {
		go player.Play(item)
	}
	q.sendQueueUpdate()
	logrus.Info("Added " + media.Title + " to queue.")
}

// Advance moves the Queue up by one and plays the next item if it exists. Advance is thread-safe.
func (q *Queue) Advance() {
	q.Lock()
	defer q.Unlock()
	q.items = q.items[1:]
	if len(q.items) > 0 {
		player := GetPlayer()
		go player.Play(q.items[0])
	}
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
		ready: make(chan struct{}),
		queue: q,
	}
	close(quietItem.ready)
	quietQueue = append(quietQueue, q.items[0], quietItem)
	quietQueue = append(quietQueue, q.items[1:]...)
	q.items = quietQueue
	q.items[0].cancel()
}

// List returns all of the items currently on the queue as a JSON response. List is thread safe.
func (q *Queue) List() ([]byte, error) {
	q.RLock()
	defer q.RUnlock()
	response, err := q.generateResponse()
	if err != nil {
		return nil, err
	}
	return response, nil
}

// MoveTo moves a QueueItem to a specific position in the Queue. MoveTo is thread-safe.
func (q *Queue) MoveTo(index int, to int) {
	q.Lock()
	defer q.Unlock()
	if to == -1 {
		to = len(q.items) - 1
	}
	if index == 0 || index == to || to == 0 {
		return
	}
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
		q.sendQueueUpdate()
	} else {
		logrus.Warnf("user provided invalid request to move item at index %v to new index %v", index, to)
	}
}

// generateResponse generates a JSON response of all the QueueItems in the Queue.
func (q *Queue) generateResponse() ([]byte, error) {
	// Cannot return a nil slice or the frontend will have issues.
	items := make([]QueueItemResponse, 0)
	for _, item := range q.items {
		items = append(items, item.generateResponse())
	}
	wrapper := QueueResponse{
		Balancing: q.balancing,
		Items:     items,
	}
	response, err := json.Marshal(wrapper)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// Remove removes a QueueItem from the Queue. Remove is thread-safe.
// Note that slice indices must be integers in Go. Remove can also be used to skip a currently playing QueueItem.
func (q *Queue) Remove(index int) {
	q.Lock()
	defer q.Unlock()
	if index < len(q.items) {
		q.items[index].cancel()
		if index == 0 {
			return
		}
		q.items[index].cancel()
		q.items = append(q.items[:index], q.items[index+1:]...)
		logrus.Debugf("remove item at index %v from queue", index)
		q.sendQueueUpdate()
	} else {
		logrus.Infof("user attempted to remove now nonexistent queue item at index %v, ignoring", index)
	}
}

// SetBalancing turns on and off balancing queue ordering. SetBalancing is thread-safe.
func (q *Queue) SetBalancing(balancing bool) {
	q.Lock()
	defer q.Unlock()
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

func (q *Queue) sendQueueUpdate() {
	response, err := q.generateResponse()
	if err != nil {
		logrus.Errorf("error generating queue response: %v", err)
		return
	}
	q.Update <- response
}

// generateResponse returns the QueueItemResponse form of the QueueItem.
func (q QueueItem) generateResponse() QueueItemResponse {
	downloading, progress := q.progress()
	return QueueItemResponse{
		Downloading: downloading,
		Error:       q.err,
		Id:          q.id,
		Media:       q.Media,
		Progress:    progress,
	}
}

// Progress returns the download progress of the QueueItem.
func (q QueueItem) progress() (bool, int) {
	key := DownloadKey{
		id:        q.Media.ID,
		mediaType: q.Media.Type,
		video:     q.Media.Video,
	}
	download, ok := q.queue.downloads[key]
	if !ok {
		return false, 100
	}
	return true, download.progress
}

// Shuffle performs a shuffle on the items in the Queue. Shuffle is thread-safe.
func (q *Queue) Shuffle() {
	q.Lock()
	defer q.Unlock()
	if len(q.items) > 1 {
		// Offset by one since we don't want to modify the currently playing item.
		rand.Shuffle(len(q.items)-1, func(i, j int) {
			q.items[i+1], q.items[j+1] = q.items[j+1], q.items[i+1]
		})
		q.sendQueueUpdate()
	}
}
