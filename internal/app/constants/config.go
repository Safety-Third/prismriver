package constants

// Environment variables used to configure the application.
// these should very much be camelCased but viper requires a string replacer in order to parse environment variable
// names, which would be a pain to get working with anything that isn't snake_cased.
const (
	// ALLOWED_TYPES specifies the media types allowed to be downloaded.
	ALLOWED_TYPES = "allowed_types"
	// DATA specifies the data storage directory.
	DATA = "data_dir"
	// DB_HOST specifies the database connection host.
	DB_HOST = "db_host"
	// DB_NAME specifies the database name.
	DB_NAME = "db_name"
	// DB_PASSWORD specifies the database connection password.
	DB_PASSWORD = "db_password"
	// DB_PORT specifies the database connection port.
	DB_PORT = "db_port"
	// DB_USER specifies the database connection user.
	DB_USER = "db_user"
	// DOWNLOAD_FORMAT specifies which format to use for downloading media.
	DOWNLOAD_FORMAT = "download_format"
	// ORIGIN specifies an optional origin to accept cross-origin requests from.
	ORIGIN = "origin"
	// VERBOSITY specifies the logging verbosity.
	VERBOSITY = "verbosity"
	// VIDEO_TRANSCODING specifies whether or not to enable video transcoding.
	VIDEO_TRANSCODING = "video_transcoding"

	// CONFIGPATH denotes the expected location of the Prismriver config file.
	CONFIG_PATH = "/etc/prismriver/prismriver.yml"
)
