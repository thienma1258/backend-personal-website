package services

import (
	"dongpham/src/model"
	"dongpham/src/repository"
)

type GalleryServices struct {
	Repo *repository.GalleryRepository
}

func (gs *GalleryServices) GetAllGallery() ([]model.GalleryImage, error) {
	ats := gs.Repo.GetAllGallery()
	return ats, nil

}

func NewGalleryServices(repo *repository.GalleryRepository) *GalleryServices {
	return &GalleryServices{Repo: repo}
}
