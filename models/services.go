package models

import "github.com/jinzhu/gorm"

// NewServices returns a single copy of services needed for the web app
func NewServices(connectionInfo string) (*Services, error) {
	db, err := gorm.Open("postgres", connectionInfo)
	if err != nil {
		return nil, err
	}

	db.LogMode(true)

	return &Services{
		User:    NewUserService(db),
		Gallery: &galleryGorm{},
	}, nil
}

// Services represents all the services, e.g. GalleryService, UserService
type Services struct {
	Gallery GalleryService
	User    UserService
}
