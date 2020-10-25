package models

import "github.com/jinzhu/gorm"

const (
	// ErrUserIDRequired is returned if user ID is not present
	ErrUserIDRequired modelError = "models: user ID is required"
	// ErrTitleRequired is returned if gallery title is not present
	ErrTitleRequired modelError = "models: title is required"
)

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

// Create creates gallery
func (gv *galleryValidator) Create(gallery *Gallery) error {
	err := runGalleryValFns(
		gallery,
		gv.userIDRequired,
		gv.titleRequired,
	)
	if err != nil {
		return err
	}

	return gv.GalleryDB.Create(gallery)
}

func (gv *galleryValidator) userIDRequired(g *Gallery) error {
	if g.UserID <= 0 {
		return ErrUserIDRequired
	}
	return nil
}

func (gv *galleryValidator) titleRequired(g *Gallery) error {
	if g.Title == "" {
		return ErrTitleRequired
	}
	return nil
}

// Ensure galleryGorm implements GalleryDB interface
var _ GalleryDB = &galleryGorm{}

type galleryGorm struct {
	db *gorm.DB
}

func (gg *galleryGorm) Create(gallery *Gallery) error {
	return gg.db.Create(gallery).Error
}

type galleryValFn func(*Gallery) error

func runGalleryValFns(gallery *Gallery, fns ...galleryValFn) error {
	for _, fn := range fns {
		if err := fn(gallery); err != nil {
			return err
		}
	}

	return nil
}
