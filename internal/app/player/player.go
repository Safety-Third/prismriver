package player

import (
	"encoding/json"
	"errors"
	"path"
	"sync"
	"time"

	"github.com/adrg/libvlc-go"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"gitlab.com/ttpcodes/prismriver/internal/app/constants"
)

var playerInstance *Player
var playerOnce sync.Once
var playerTicker *time.Ticker

// Represents various states that the Player can exist in.
const (
	// STOPPED represents a stopped state when nothing is playing.
	STOPPED = iota
	// PLAYING represents a playing state.
	PLAYING = iota
	// PAUSED represents a paused state.
	PAUSED = iota
	// LOADING represents a loading state before playback begins.
	LOADING = iota
)

// Player represents a player for Media items.
type Player struct {
	doneChan chan bool
	player   *vlc.Player
	State    int
	Update   chan []byte
	Volume   int
}

// State represents status information about the Player, such as the time, state, and volume.
type State struct {
	CurrentTime int
	TotalTime   int
	State       int
	Volume      int
}

// GetPlayer returns the single Player instance used by the application.
func GetPlayer() *Player {
	playerOnce.Do(func() {
		playerInstance = &Player{
			doneChan: make(chan bool),
			State:    STOPPED,
			Update:   make(chan []byte),
			Volume:   100,
		}
		playerTicker = time.NewTicker(30 * time.Second)
		go func() {
			for range playerTicker.C{
				response := playerInstance.GenerateResponse()
				playerInstance.Update <- response
			}
		}()
	})
	return playerInstance
}

// GenerateResponse generates a JSON response representing the Player's current status.
func (p Player) GenerateResponse() []byte {
	if p.State == PLAYING {
		currentTime, err := p.player.MediaTime()
		if err != nil {
			logrus.Error("Error getting player's media time:")
			logrus.Error(err)
		}
		totalTime, err := p.player.MediaLength()
		if err != nil {
			logrus.Error("Error getting player's media length:")
			logrus.Error(err)
		}
		response, err := json.Marshal(State{
			CurrentTime: currentTime,
			State:       p.State,
			TotalTime:   totalTime,
			Volume:      p.Volume,
		})
		if err != nil {
			logrus.Error("Error generating JSON response:")
			logrus.Error(err)
		}
		return response
	}

	response, err := json.Marshal(State{
		CurrentTime: 0,
		State:       p.State,
		TotalTime:   0,
		Volume:      p.Volume,
	})
	if err != nil {
		logrus.Error("Error generating JSON response:")
		logrus.Error(err)
	}
	return response
}

// Play begins playback on a QueueItem.
func (p *Player) Play(item *QueueItem) error {
	defer func() {
		p.State = STOPPED
		queue := GetQueue()
		queue.Advance()
	}()
	p.State = LOADING
	dataDir := viper.GetString(constants.DATA)
	ext := ".opus"
	if item.Media.Video {
		if viper.GetBool(constants.VIDEO_TRANSCODING) {
			ext = ".mp4"
		} else {
			ext = ".video"
		}
	}
	filePath := path.Join(dataDir, item.Media.Type, item.Media.ID+ext)
	select {
	case <-item.ctx.Done():
		logrus.Infof("context canceled, not playing media")
		return nil
	case <-item.ready:
	}

	if err := vlc.Init("--quiet", "--fullscreen"); err != nil {
		logrus.Error("Error initializing vlc:")
		logrus.Error(err)
		return err
	}

	defer func() {
		if err := vlc.Release(); err != nil {
			logrus.Errorf("error releasing vlc instance: %v", err)
		}
		p.player = nil
	}()

	var err error
	p.player, err = vlc.NewPlayer()
	if err != nil {
		logrus.Error("Error creating player:")
		logrus.Error(err)
		return err
	}
	defer func() {
		if err := p.player.Stop(); err != nil {
			logrus.Errorf("error stopping vlc player: %v", err)
		}
		if err := p.player.Release(); err != nil {
			logrus.Errorf("error releasing vlc player: %v", err)
		}
	}()

	vlcMedia, err := p.player.LoadMediaFromPath(filePath)
	if err != nil {
		logrus.Error("Error loading media file:")
		logrus.Error(err)
		return err
	}
	defer func() {
		if err := vlcMedia.Release(); err != nil {
			logrus.Errorf("error releasing media item: %v", err)
		}
	}()

	if err := p.player.Play(); err != nil {
		logrus.Error("Error playing media file:")
		logrus.Error(err)
		return err
	}

	if err := p.player.SetVolume(p.Volume); err != nil {
		logrus.Errorf("error setting volume: %v", err)
		return err
	}

	if err := p.player.SetFullScreen(true); err != nil {
		logrus.Errorf("error setting fullscreen: %v", err)
		return err
	}

	eventManager, err := p.player.EventManager()
	if err != nil {
		logrus.Errorf("error retrieving vlc EventManager: %v", err)
		return err
	}

	eventID, err := eventManager.Attach(vlc.MediaPlayerEndReached, func(event vlc.Event, userData interface{}) {
		item.cancel()
		logrus.Debugf("playback finished")
	}, nil)
	if err != nil {
		logrus.Errorf("error registering MediaPlayerEndReached event: %v", err)
		return err
	}
	defer eventManager.Detach(eventID)

	time.Sleep(1 * time.Second)
	length, err := p.player.MediaLength()
	if err != nil || length == 0 {
		length = 1000 * 60
	}
	p.State = PLAYING
	p.sendPlayerUpdate()

	<-item.ctx.Done()

	p.State = STOPPED
	p.sendPlayerUpdate()

	return nil
}

// UpVolume increments the volume of the Player by 5, up to a maximum of 100.
func (p *Player) UpVolume() {
	if p.Volume == 100 {
		return
	}
	if p.State == PLAYING {
		if err := p.player.SetVolume(p.Volume + 5); err != nil {
			logrus.Errorf("error setting volume: %v", err)
			return
		}
	}
	p.Volume += 5
	p.sendPlayerUpdate()
}

// DownVolume decrements the volume of the Player by 5, down to a minimum of 0.
func (p *Player) DownVolume() {
	if p.Volume == 0 {
		return
	}
	if p.State == PLAYING {
		if err := p.player.SetVolume(p.Volume - 5); err != nil {
			logrus.Errorf("error setting volume: %v", err)
			return
		}
	}
	p.Volume -= 5
	p.sendPlayerUpdate()
}

// Seek sets the player to a certain time.
func (p *Player) Seek(milliseconds int) error {
	if p.State != PLAYING {
		return errors.New("cannot seek player that isn't playing")
	}
	if err := p.player.SetMediaTime(milliseconds); err != nil {
		return err
	}
	return nil
}

func (p Player) sendPlayerUpdate() {
	response := p.GenerateResponse()
	p.Update <- response
	logrus.Debug("Sent player update event.")
}
