package models

import (
	"errors"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/nahuakang/gophotos/hash"
	"github.com/nahuakang/gophotos/rand"
	"golang.org/x/crypto/bcrypt"
)

const (
	hmacSecretKey      = "secret-hmac-key"
	userPasswordPepper = "secret-random-string"
)

var (
	// ErrNotFound is returned when a resource cannot be found
	// in the database.
	ErrNotFound = errors.New("models: resource not found")

	// ErrInvalidID is returned when an invalid ID is provided
	// to a method such as Delete.
	ErrInvalidID = errors.New("models: ID provided is invalid")

	// ErrInvalidPassword is returned when incorrect password is provided.
	ErrInvalidPassword = errors.New("models: incorrect password provided")
)

// UserDB interacts with the users database.
//
// For single user queries:
// If the user is found, a nil error is returned
// If the user is not found, ErrNotFound is returned
// If another error occurs, error with more information is returned
//
// For single user queries, any error but ErrNotFound results in a
// 500 error until public-facing errors are created.
type UserDB interface {
	ByID(id uint) (*User, error)
	ByEmail(email string) (*User, error)
	ByRemember(token string) (*User, error)

	// Methods for altering users
	Create(user *User) error
	Update(user *User) error
	Delete(id uint) error

	// Close DB connection method
	Close() error

	// Migration helpers
	AutoMigrate() error
	DestructiveReset() error
}

// userGorm represents database interaction layer
// and implements the UserDB interface fully.
type userGorm struct {
	db *gorm.DB
}

// UserService is an abstract layer to interact with gorm.DB
type UserService interface {
	// Authenticate verifies the provided email address and password
	// are correct. If they are correct, the user corresponding to the
	// email is returned. Otherwise, ErrNotFound, ErrInvalidPassword,
	// or another error is returned.
	Authenticate(email, password string) (*User, error)
	UserDB
}

type userService struct {
	UserDB
}

// userValidator is the validation layer that validates
// and normalizes data before passing it on to the next
// UserDB in the interface chain
type userValidator struct {
	UserDB
	hmac hash.HMAC
}

// User represents a user data type
type User struct {
	gorm.Model
	Name         string
	Email        string `gorm:"not null;unique_index"`
	Password     string `gorm:"-"`
	PasswordHash string `gorm:"not null"`
	Remember     string `gorm:"-"`
	RememberHash string `gorm:"not null;unique_index"`
}

// userValFn is the function type for user validation functions
type userValFn func(*User) error

// runUserValFns accepts a pointer to a user and any number of
// validation functions that comply to the type userValFn signature,
// then iterates over all the validation functions.
func runUserValFns(user *User, fns ...userValFn) error {
	for _, fn := range fns {
		if err := fn(user); err != nil {
			return err
		}
	}
	return nil
}

// NewUserService returns a pointer to UserService
func NewUserService(connectionInfo string) (UserService, error) {
	ug, err := newUserGorm(connectionInfo)
	if err != nil {
		return nil, err
	}

	hmac := hash.NewHMAC(hmacSecretKey)
	uv := &userValidator{
		hmac:   hmac,
		UserDB: ug,
	}
	return &userService{
		UserDB: uv,
	}, nil
}

func newUserGorm(connectionInfo string) (*userGorm, error) {
	db, err := gorm.Open("postgres", connectionInfo)
	if err != nil {
		return nil, err
	}

	db.LogMode(true)
	return &userGorm{
		db: db,
	}, nil
}

// bcryptPassword hashes a user's password with an
// app-wide pepper and bcrypt, which salts the password.
func (uv *userValidator) bcryptPassword(user *User) error {
	if user.Password == "" {
		// NO need to run this function if the user's
		// password is not changed.
		return nil
	}

	pwBytes := []byte(user.Password + userPasswordPepper)
	hashedBytes, err := bcrypt.GenerateFromPassword(pwBytes, bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(hashedBytes)
	user.Password = ""

	return nil
}

// hmacRemember is a helper function to be consumed by userValidator.ByRemember
func (uv *userValidator) hmacRemember(user *User) error {
	if user.Remember == "" {
		return nil
	}
	user.RememberHash = uv.hmac.Hash(user.Remember)
	return nil
}

// ByRemember hashes the remember token and calls
// ByRemember on the subsequent UserDB layer.
func (uv *userValidator) ByRemember(token string) (*User, error) {
	user := User{
		Remember: token,
	}
	if err := runUserValFns(&user, uv.hmacRemember); err != nil {
		return nil, err
	}

	return uv.UserDB.ByRemember(user.RememberHash)
}

// Create creates the provided user and backfills data
// such as ID, CreateAt, and UpdateAt fields.
func (uv *userValidator) Create(user *User) error {
	if user.Remember == "" {
		token, err := rand.RememberToken()
		if err != nil {
		}
		user.Remember = token
	}

	err := runUserValFns(
		user,
		uv.bcryptPassword,
		uv.hmacRemember,
	)
	if err != nil {
		return err
	}

	return uv.UserDB.Create(user)
}

// Update hashes a remember token if one is provided
func (uv *userValidator) Update(user *User) error {
	err := runUserValFns(
		user,
		uv.bcryptPassword,
		uv.hmacRemember,
	)
	if err != nil {
		return err
	}

	return uv.UserDB.Update(user)
}

// Delete deletes the user with the provided ID
func (uv *userValidator) Delete(id uint) error {
	if id == 0 {
		return ErrInvalidID
	}
	return uv.UserDB.Delete(id)
}

// Create will create the provided user and backfill data
// like the ID, CreateAt, and UpdateAt fields.
func (ug *userGorm) Create(user *User) error {
	return ug.db.Create(user).Error
}

// Update updates the provided user with the data provided.
func (ug *userGorm) Update(user *User) error {
	return ug.db.Save(user).Error
}

// Delete deletes the user with the provided ID
func (ug *userGorm) Delete(id uint) error {
	user := User{Model: gorm.Model{ID: id}}
	return ug.db.Delete(&user).Error
}

// ByID looks up a user with the provided ID.
// If the user is found, nil error is returned.
// If the user is not found, ErrNotFound is returned.
// If there is another error, we will return an error with
// more information about what went wrong. This may not be
// an error generated by the models package.
//
// As a general rule, any error but ErrNotFound should result
// in a 500 error.
func (ug *userGorm) ByID(id uint) (*User, error) {
	var user User
	db := ug.db.Where("id = ?", id)
	err := first(db, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// ByEmail looks up a user with the given email address and
// returns the user.
// If the user is found, nil is returned as error.
// If the user is not found, ErrNotFound is returned.
// If there is another error, the error is returned with
// more information on what went wrong.
func (ug *userGorm) ByEmail(email string) (*User, error) {
	var user User
	db := ug.db.Where("email = ?", email)
	err := first(db, &user)
	return &user, err
}

// ByRemember looks up a user with a given remember token
// and returns the user from the database. This method handles
// token hashing.
func (ug *userGorm) ByRemember(rememberHash string) (*User, error) {
	var user User
	err := first(ug.db.Where("remember_hash = ?", rememberHash), &user) // Gorm uses snake case
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// DestructiveReset drops the user table and rebuilds it
func (ug *userGorm) DestructiveReset() error {
	err := ug.db.DropTableIfExists(&User{}).Error
	if err != nil {
		return err
	}
	return ug.AutoMigrate()
}

// AutoMigrate will attempt to automatically migrate user table
func (ug *userGorm) AutoMigrate() error {
	if err := ug.db.AutoMigrate(&User{}).Error; err != nil {
		return err
	}

	return nil
}

// Close closes UserService database connection
func (ug *userGorm) Close() error {
	return ug.db.Close()
}

// first will query using the provided gorm.DB and it returns
// the first item returned and place it into dst. If nothing
// is found in the query, the method returns ErrNotFound
func first(db *gorm.DB, dst interface{}) error {
	err := db.First(dst).Error
	if err == gorm.ErrRecordNotFound {
		return ErrNotFound
	}
	return err
}

// Authenticate authenticates a user with the provided email and password.
// If the email address provided is invalid, return nil, ErrNotFound
// If the password provided is invalid, return nil, ErrInvalidPassword
// If the email and the password are both valid, return user, nil
// Otherwise, return nil, error
func (us *userService) Authenticate(email, password string) (*User, error) {
	foundUser, err := us.ByEmail(email)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(foundUser.PasswordHash),
		[]byte(password+userPasswordPepper),
	)

	switch err {
	case nil:
		return foundUser, nil
	case bcrypt.ErrMismatchedHashAndPassword:
		return nil, ErrInvalidPassword
	default:
		return nil, err
	}
}
