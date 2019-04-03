package main

import (
	"github.com/jinzhu/gorm"
)

// Face  Fake face
type Face struct {
	ID string `json:"id" gorm:"primary_key"`
}

func main() {

	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		panic("Database open failed")
	}
	defer db.Close()

	db.AutoMigrate(&Face{})

}
