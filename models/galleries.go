package models

import "github.com/jinzhu/gorm"

// Gallery represents the model for a user gallery
type Gallery struct {
	gorm.Model
	UserID uint   `gorm:"not_null;index"`
	Title  string `gorm:"not_null"`
}

// NewGalleryService returns a GalleryService
func NewGalleryService(db *gorm.DB) GalleryService {
	return &galleryService{
		GalleryDB: &galleryValidator{
			GalleryDB: &galleryGorm{
				db: db,
			},
		},
	}
}

// GalleryService is an interface that represents services to Gallery
type GalleryService interface {
	GalleryDB
}

type galleryService struct {
	GalleryDB
}

// GalleryDB represents interface for the gallery database
type GalleryDB interface {
	Create(gallery *Gallery) error
}

type galleryValidator struct {
	GalleryDB
}

// Ensure galleryGorm implements GalleryDB interface
var _ GalleryDB = &galleryGorm{}

type galleryGorm struct {
	db *gorm.DB
}

func (gg *galleryGorm) Create(gallery *Gallery) error {
	return gg.db.Create(gallery).Error
}
