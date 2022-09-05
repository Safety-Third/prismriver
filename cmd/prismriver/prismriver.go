package main

import (
	"io"
	"os"
	"path"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/Safety-Third/prismriver/assets"
	"github.com/Safety-Third/prismriver/internal/app/constants"
	"github.com/Safety-Third/prismriver/internal/app/server"
)

func main() {
	// Set up configuration framework.
	viper.SetEnvPrefix("prismriver")
	viper.AutomaticEnv()

	viper.SetDefault(constants.ALLOWED_TYPES, []string{"soundcloud", "youtube"})
	viper.SetDefault(constants.DATA, "/var/lib/prismriver")
	viper.SetDefault(constants.DB_HOST, "localhost")
	viper.SetDefault(constants.DB_NAME, "prismriver")
	viper.SetDefault(constants.DB_PASSWORD, "prismriver")
	viper.SetDefault(constants.DB_PORT, "5432")
	viper.SetDefault(constants.DB_USER, "prismriver")
	viper.SetDefault(constants.DOWNLOAD_FORMAT, "bestvideo+bestaudio/best")
	viper.SetDefault(constants.ORIGIN, "")
	viper.SetDefault(constants.VERBOSITY, "info")
	viper.SetDefault(constants.VIDEO_TRANSCODING, true)

	envVars := []string{
		constants.ALLOWED_TYPES,
		constants.DB_HOST,
		constants.DB_NAME,
		constants.DB_PASSWORD,
		constants.DB_PORT,
		constants.DB_USER,
		constants.DOWNLOAD_FORMAT,
		constants.ORIGIN,
		constants.VERBOSITY,
		constants.VIDEO_TRANSCODING,
	}

	for _, env := range envVars {
		if err := viper.BindEnv(env); err != nil {
			logrus.Warnf("error binding to variable %v: %v", env, err)
		}
	}

	viper.SetConfigFile(constants.CONFIG_PATH)
	if err := viper.ReadInConfig(); err != nil {
		logrus.Infof("could not read config file, ignoring: %v", err)
	}

	verbosity := viper.GetString(constants.VERBOSITY)
	level, err := logrus.ParseLevel(verbosity)
	if err != nil {
		logrus.Errorf("Error reading verbosity level in configuration: %v", err)
	}
	logrus.SetLevel(level)
	// trust me, there isn't a nicer way to do this without type hacking or structs to track things like variable
	// privacy.
	logrus.Debugf("current configuration:")
	logrus.Debugf("%v:", constants.ALLOWED_TYPES)
	for _, allowedType := range viper.GetStringSlice(constants.ALLOWED_TYPES) {
		logrus.Debugf("- %v", allowedType)
	}
	logrus.Debugf("%v: %v", constants.DB_HOST, viper.GetString(constants.DB_HOST))
	logrus.Debugf("%v: %v", constants.DB_NAME, viper.GetString(constants.DB_NAME))
	logrus.Debugf("%v: [hidden]", constants.DB_PASSWORD)
	logrus.Debugf("%v: %v", constants.DB_PORT, viper.GetString(constants.DB_PORT))
	logrus.Debugf("%v: %v", constants.DB_USER, viper.GetString(constants.DB_USER))
	logrus.Debugf("%v: %v", constants.DOWNLOAD_FORMAT, viper.GetString(constants.DOWNLOAD_FORMAT))
	logrus.Debugf("%v: %v", constants.ORIGIN, viper.GetString(constants.ORIGIN))
	logrus.Debugf("%v: %v", constants.VERBOSITY, viper.GetString(constants.VERBOSITY))
	logrus.Debugf("%v: %v", constants.VIDEO_TRANSCODING, viper.GetBool(constants.VIDEO_TRANSCODING))

	dataDir := viper.GetString(constants.DATA)
	if err := os.MkdirAll(path.Join(dataDir, "internal"), os.ModeDir|0755); err != nil {
		logrus.Fatalf("error creating data directories: %v", err)
	}

	beQuiet, err := assets.HTTP.Open("bequiet.opus")
	if err != nil {
		logrus.Fatalf("Error reading bequiet.opus in internal filesystem (is this binary corrupted?): %v", err)
	}
	beQuietPath := path.Join(dataDir, "internal", "bequiet.opus")
	beQuietFile, err := os.Create(beQuietPath)
	if err != nil {
		logrus.Fatalf("Error creating application files: %v", err)
	}
	if _, err := io.Copy(beQuietFile, beQuiet); err != nil {
		logrus.Fatalf("error copying bequiet.opus: %v", err)
	}
	if err := beQuietFile.Close(); err != nil {
		logrus.Warnf("error closing reader on bequiet.opus: %v", err)
	}

	server.CreateRouter()
}
