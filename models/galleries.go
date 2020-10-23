package models

import "github.com/jinzhu/gorm"

// GalleryService is an interface that represents services to Gallery
type GalleryService interface {
	GalleryDB
}

// GalleryDB represents interface for the gallery database
type GalleryDB interface {
	Create(gallery *Gallery) error
}

type galleryGorm struct {
	db *gorm.DB
}

func (gg *galleryGorm) Create(gallery *Gallery) error {
	// TODO: Implement later
	return nil
}

// Gallery represents the model for a user gallery
type Gallery struct {
	gorm.Model
	UserID uint   `gorm:"not_null;index"`
	Title  string `gorm:"not_null"`
}
