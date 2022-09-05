package downloader

import (
	"io"
	"os"
	"path"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/xfrr/goffmpeg/transcoder"
	youtubedl "gitlab.com/ttpcodes/youtube-dl-go"

	"github.com/Safety-Third/prismriver/internal/app/constants"
	"github.com/Safety-Third/prismriver/internal/app/db"
)

// DownloadMedia runs a download on a given item in a goroutine. This can be tracked using the returned channels.
func DownloadMedia(media db.Media) (chan float64, chan error, error) {
	progressChan := make(chan float64)
	doneChan := make(chan error)
	go func() {
		callDone := func(err error) {
			close(progressChan)
			doneChan <- err
			close(doneChan)
		}
		downloader := youtubedl.NewDownloader(media.URL)

		format := viper.GetString(constants.DOWNLOAD_FORMAT)
		downloader.Format(format)
		downloader.NoPlaylist()
		downloader.Output(path.Join("/tmp", media.Type, media.ID))
		eventChan, closeChan, err := downloader.RunProgress()
		if err != nil {
			logrus.Error("Error starting media download:\n", err)
			callDone(err)
			return
		}
		for progress := range eventChan {
			logrus.Debugf("Download is at %f percent completion", progress)
			progressChan <- progress / 2
		}
		result := <-closeChan
		if result.Err != nil {
			logrus.Error("Error downloading media file:\n", result.Err)
			callDone(result.Err)
			return
		}
		logrus.Debug("Downloaded media file")

		dataDir := viper.GetString(constants.DATA)
		dirPath := path.Join(dataDir, media.Type)
		if err := os.MkdirAll(dirPath, os.ModeDir|0755); err != nil {
			callDone(err)
			return
		}
		if !media.Video || viper.GetBool(constants.VIDEO_TRANSCODING) {
			trans := new(transcoder.Transcoder)
			ext := ".opus"
			if media.Video {
				ext = ".mp4"
			}
			filePath := path.Join(dirPath, media.ID+ext)
			err = trans.Initialize(result.Path, filePath)
			if err != nil {
				logrus.Error("Error starting transcoding process:\n", err)
				callDone(err)
				return
			}
			trans.MediaFile().SetAudioCodec("libopus")
			if media.Video {
				trans.MediaFile().SetVideoCodec("libx264")
				// Needed to enable experimental Opus in the mp4 container format.
				trans.MediaFile().SetStrict(-2)
			} else {
				trans.MediaFile().SetSkipVideo(true)
			}
			logrus.Debug("Instantiated ffmpeg transcoder")

			done := trans.Run(true)
			progress := trans.Output()
			for msg := range progress {
				progressChan <- msg.Progress/2 + 50
				logrus.Debug(msg)
			}
			if err := <-done; err != nil {
				logrus.Error("Error in transcoding process:\n", err)
				callDone(err)
				return
			}
			logrus.Debug("Transcoded media to vorbis audio")
		} else {
			logrus.Debugf("video transcoding disabled, moving file to final destination")
			input, err := os.Open(result.Path)
			if err != nil {
				logrus.Errorf("error reading original video file: %v", err)
				callDone(err)
				return
			}
			defer func() {
				if err := input.Close(); err != nil {
					logrus.Errorf("error closing input file: %v", err)
				}
			}()
			output, err := os.Create(path.Join(dirPath, media.ID+".video"))
			if err != nil {
				logrus.Errorf("error opening destination file: %v", err)
				callDone(err)
				return
			}
			defer func() {
				if err := output.Close(); err != nil {
					logrus.Errorf("error closing output file: %v", err)
				}
			}()
			if _, err := io.Copy(output, input); err != nil {
				logrus.Errorf("error copying video file: %v", err)
				callDone(err)
				return
			}
		}
		if err := os.Remove(result.Path); err != nil {
			logrus.Warnf("error when removing temporary file: %v", err)
		}
		logrus.Debug("removed temporary youtube-dl file")
		logrus.Infof("downloaded new file for media with id %v and type %v", media.ID, media.Type)
		callDone(nil)
	}()
	return progressChan, doneChan, nil
}

// GetInfo retrieves the info for a Media item synchronously.
func GetInfo(url string, video bool) (db.Media, error) {
	downloader := youtubedl.NewDownloader(url)
	downloader.NoPlaylist()
	info, err := downloader.GetInfo()
	if err != nil {
		return db.Media{}, err
	}
	if video && info.VCodec == "none" {
		video = false
	}
	return db.Media{
		ID:     info.ID,
		Length: uint64(info.Duration * float64(time.Millisecond)),
		Title:  info.Title,
		Type:   info.Extractor,
		Video:  video,
		URL:    info.WebpageURL,
	}, nil
}

// ValidateURL checks to see if the given URL is allowed to be played.
func ValidateURL(url string) bool {
	downloader := youtubedl.NewDownloader(url)
	extractor, err := downloader.GetExtractor()
	if err != nil {
		logrus.Warnf("could not determine extractor for %v: %v", url, err)
		return false
	} else if extractor == "generic" {
		return false
	} else {
		extractor = strings.Split(extractor, ":")[0]
		allowedTypes := viper.GetStringSlice(constants.ALLOWED_TYPES)
		for _, mediaType := range allowedTypes {
			if extractor == mediaType {
				return true
			}
		}
		return false
	}
}
