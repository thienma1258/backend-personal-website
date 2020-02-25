package repository

import (
	"database/sql"
	"dongpham/src/config"
	"log"
)

var db *sql.DB
var GalleryRepo *GalleryRepository

func init() {
	var err error
	db, err = sql.Open("mysql", config.DBConnection)
	if err != nil {
		log.Fatal(err)
	}
	initRepository()
}

func initRepository() {
	GalleryRepo, _ = NewGalleryRepository(db)
}
