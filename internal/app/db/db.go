package db

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/gorm"
	"gorm.io/driver/sqlite"

	"gitlab.com/ttpcodes/prismriver/internal/app/constants"

	"fmt"
	"math"
	"path"
	"sync"
)

var db *gorm.DB
var err error
var once sync.Once

// BeQuiet is built-in media used for the "Be Quiet!" feature.
var BeQuiet = &Media{
	ID:     "bequiet",
	Length: 3710000000,
	Title:  "Please Be Quiet!",
	Type:   "internal",
}

// GetDatabase gets the instance of the database connection used for the application.
func GetDatabase() (*gorm.DB, error) {
	once.Do(func() {
		db, err = gorm.Open(sqlite.Open(path.Join(viper.GetString(constants.DATA), "prismriver.db")), &gorm.Config{})
		if err != nil {
			return
		}
		if err = db.AutoMigrate(&Media{}); err != nil {
			return
		}
		if err = db.FirstOrCreate(BeQuiet).Error; err != nil {
			return
		}
	})
	return db, err
}

// AddMedia adds a new Media to the database.
func AddMedia(media Media) error {
	db, err := GetDatabase()
	if err != nil {
		return err
	}
	db.Create(&media)
	return nil
}

// FindMedia searches the database for Media items matching the title in query and returns the number of results
// specified by limit.
func FindMedia(query string, limit int, page int) ([]Media, uint) {
	db, err := GetDatabase()
	if err != nil {
		logrus.Fatal("Error loading database:", err)
	}
	if page == 0 {
		page = 1
	}
	var media []Media
	db.Limit(limit).Offset((page - 1) * limit).Where("title LIKE ? AND type <> ?", "%"+query+"%", "internal").Find(&media)
	var count float64
	db.Model(&Media{}).Where("title ILIKE ? AND type <> ?", "%" + query + "%", "internal").Count(&count)
	return media, uint(math.Ceil(count / float64(limit)))
}

// GetMedia attempts to return the Media identified by id and kind, and returns an error if not found.
func GetMedia(id string, kind string) (Media, error) {
	db, err := GetDatabase()
	if err != nil {
		logrus.Fatal("Error loading database:", err)
	}
	var media []Media
	db.Where(Media{ID: id, Type: kind}).First(&media)
	if len(media) > 0 {
		return media[0], nil
	}
	return Media{}, errors.New("media not found in DB")
}

// GetMediaByURL attempts to return the Media identified by url, and returns an error if not found.
func GetMediaByURL(url string) (Media, error) {
	db, err := GetDatabase()
	if err != nil {
		logrus.Fatalf("could not load database: %v", err)
	}
	var media []Media
	db.Where(Media{URL: url}).First(&media)
	if len(media) > 0 {
		return media[0], nil
	}
	return Media{}, errors.New(fmt.Sprintf("media with url %v not found in database", url))
}

// GetRandomMedia returns a number of random Media specified by limit.
func GetRandomMedia(limit int) []Media {
	db, err := GetDatabase()
	if err != nil {
		logrus.Fatal("Error loading database:", err)
	}
	var media []Media
	db.Order("random()").Where("type <> ?", "internal").Limit(limit).Find(&media)
	return media
}
