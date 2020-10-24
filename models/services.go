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
		db:      db,
	}, nil
}

// Services represents all the services, e.g. GalleryService, UserService
type Services struct {
	Gallery GalleryService
	User    UserService
	db      *gorm.DB
}

// Close closes the database connection from Services layer
func (s *Services) Close() error {
	return s.db.Close()
}

// AutoMigrate automigrates all tables for Services.db
func (s *Services) AutoMigrate() error {
	return s.db.AutoMigrate(&User{}, &Gallery{}).Error
}

// DestructiveReset drops all tables and rebuilds them
func (s *Services) DestructiveReset() error {
	err := s.db.DropTableIfExists(&User{}, &Gallery{}).Error
	if err != nil {
		return err
	}

	return s.AutoMigrate()
}
