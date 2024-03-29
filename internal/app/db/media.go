package db

import (
	"github.com/sirupsen/logrus"

	"time"
)

// Media represents a Media item in the database.
type Media struct {
	ID        string `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time

	Length uint64 `gorm:"not null"`
	Title  string `gorm:"not null"`
	Type   string `gorm:"primary_key"`
	Video  bool   `gorm:"not null"`
	URL    string `gorm:"not null"`
}

func (m Media) Save() {
	db, err := GetDatabase()
	if err != nil {
		logrus.Fatalf("could not load database: %v", err)
	}
	db.Save(&m)
}
