package repository

import (
	"database/sql"
	"dongpham/src/model"
)

type GalleryRepository struct {
	db *sql.DB
}

func (up *GalleryRepository) GetAllGallery() []model.GalleryImage {
	//up.db.Exec()
	return nil
}

func NewGalleryRepository(db *sql.DB) (*GalleryRepository, error) {
	return &GalleryRepository{db: db}, nil
}
