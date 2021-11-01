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
// Not all Player fields are accessed by multiple threads. player is safely accessible without the lock.
type Player struct {
	sync.RWMutex

	doneChan chan struct{}
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
			doneChan: make(chan struct{}),
			State:    STOPPED,
			Update:   make(chan []byte),
			Volume:   100,
		}
		playerTicker = time.NewTicker(30 * time.Second)
		go func() {
			for range playerTicker.C{
				response, err := playerInstance.generateResponse()
				if err != nil {
					logrus.Errorf("could not generate player response: %v", err)
					continue
				}
				playerInstance.Update <- response
			}
		}()
	})
	return playerInstance
}

// generateResponse generates a JSON response representing the Player's current status.
func (p *Player) generateResponse() ([]byte, error) {
	if p.State == PLAYING {
		currentTime, err := p.player.MediaTime()
		if err != nil {
			return nil, err
		}
		totalTime, err := p.player.MediaLength()
		if err != nil {
			return nil, err
		}
		response, err := json.Marshal(State{
			CurrentTime: currentTime,
			State:       p.State,
			TotalTime:   totalTime,
			Volume:      p.Volume,
		})
		if err != nil {
			return nil, err
		}
		return response, nil
	}

	response, err := json.Marshal(State{
		CurrentTime: 0,
		State:       p.State,
		TotalTime:   0,
		Volume:      p.Volume,
	})
	if err != nil {
		return nil, err
	}
	return response, nil
}

// Get returns the Player's current status. Get is thread-safe.
func (p *Player) Get() ([]byte, error) {
	p.RLock()
	defer p.RUnlock()
	response, err := p.generateResponse()
	if err != nil {
		return nil, err
	}
	return response, nil
}

// Play begins playback on a QueueItem. Play is thread-safe.
func (p *Player) Play(item *QueueItem) error {
	defer func() {
		p.Lock()
		p.State = STOPPED
		p.sendPlayerUpdate()
		p.Unlock()
		p.doneChan <- struct{}{}
	}()
	p.Lock()
	p.State = LOADING
	p.Unlock()
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

	p.Lock()
	p.State = PLAYING
	p.Unlock()

	if err := p.player.Play(); err != nil {
		logrus.Error("Error playing media file:")
		logrus.Error(err)
		return err
	}

	p.RLock()
	if err := p.player.SetVolume(p.Volume); err != nil {
		logrus.Errorf("error setting volume: %v", err)
		p.RUnlock()
		return err
	}
	p.RUnlock()

	if err := p.player.SetFullScreen(true); err != nil {
		logrus.Errorf("error setting fullscreen: %v", err)
		return err
	}

	eventManager, err := p.player.EventManager()
	if err != nil {
		logrus.Errorf("error retrieving vlc EventManager: %v", err)
		return err
	}

	// play() does not guarantee that metadata will be available, so we wait for mediaplayerplaying instead
	eventID, err := eventManager.Attach(vlc.MediaPlayerPlaying, func(event vlc.Event, userData interface{}) {
		p.RLock()
		p.sendPlayerUpdate()
		p.RUnlock()
	}, nil)
	if err != nil {
		logrus.Errorf("error registering mediaplayerplaying event: %v", err)
		return err
	}
	defer eventManager.Detach(eventID)
	eventID, err = eventManager.Attach(vlc.MediaPlayerEndReached, func(event vlc.Event, userData interface{}) {
		item.cancel()
		logrus.Debugf("playback finished")
	}, nil)
	if err != nil {
		logrus.Errorf("error registering MediaPlayerEndReached event: %v", err)
		return err
	}
	defer eventManager.Detach(eventID)

	<-item.ctx.Done()
	return nil
}

// UpVolume increments the volume of the Player by 5, up to a maximum of 100. UpVolume is thread-safe.
func (p *Player) UpVolume() {
	p.Lock()
	defer p.Unlock()
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

// DownVolume decrements the volume of the Player by 5, down to a minimum of 0. DownVolume is thread-safe.
func (p *Player) DownVolume() {
	p.Lock()
	defer p.Unlock()
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

// Seek sets the player to a certain time. Seek is thread-safe.
func (p *Player) Seek(milliseconds int) error {
	p.Lock()
	defer p.Unlock()
	if p.State != PLAYING {
		return errors.New("cannot seek player that isn't playing")
	}
	if err := p.player.SetMediaTime(milliseconds); err != nil {
		return err
	}
	return nil
}

func (p *Player) sendPlayerUpdate() {
	response, err := p.generateResponse()
	if err != nil {
		logrus.Errorf("could not generate player response: %v", err)
		return
	}
	p.Update <- response
	logrus.Debug("Sent player update event.")
}
