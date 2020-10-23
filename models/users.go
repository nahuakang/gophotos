package models

import (
	"regexp"
	"strings"

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

type modelError string

// Error ensures that modelError implements error interface
func (e modelError) Error() string {
	return string(e)
}

// Public ensures that modelError implements views.PublicError interface
func (e modelError) Public() string {
	s := strings.Replace(string(e), "models: ", "", 1)
	split := strings.Split(s, " ")
	split[0] = strings.Title(split[0])
	return strings.Join(split, " ")
}

var (
	// ErrNotFound is returned when a resource cannot be found
	// in the database.
	ErrNotFound modelError = "models: resource not found"

	// ErrIDInvalid is returned when an invalid ID is provided
	// to a method such as Delete.
	ErrIDInvalid modelError = "models: ID provided is invalid"

	// ErrPasswordRequired is returned when a create is attempted without password
	ErrPasswordRequired modelError = "models: password is required"

	// ErrPasswordIncorrect is returned when incorrect password is provided.
	ErrPasswordIncorrect modelError = "models: incorrect password provided"

	// ErrEmailRequired is returned when an email address is not provided
	// when creating a user account.
	ErrEmailRequired modelError = "models: email address is required"

	// ErrEmailInvalid is returned when an invalid email address is provided.
	ErrEmailInvalid modelError = "models: email address is not valid"

	// ErrEmailTaken is returned when an update or create is attempted with
	// an email address that is already registered.
	ErrEmailTaken modelError = "models: email address is already taken"

	//ErrPasswordTooShort is returned if the provided password is shorter than
	// 8 characters in length.
	ErrPasswordTooShort modelError = "models: password should be at least 8 characters long"

	// ErrRememberRequired is returned when a create or update is attempted
	// without a user remember token hash.
	ErrRememberRequired modelError = "models: remember token is required"

	// ErrRememberTooShort is returned when a remember token is not at least 32 bytes
	ErrRememberTooShort modelError = "models: remember token must be at least 32 bytes"
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
	// email is returned. Otherwise, ErrNotFound, ErrPasswordIncorrect,
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
	hmac       hash.HMAC
	emailRegex *regexp.Regexp
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
func NewUserService(db *gorm.DB) UserService {
	ug := &userGorm{db}

	hmac := hash.NewHMAC(hmacSecretKey)
	uv := newUserValidator(ug, hmac)

	return &userService{
		UserDB: uv,
	}
}

func newUserValidator(udb UserDB, hmac hash.HMAC) *userValidator {
	return &userValidator{
		UserDB: udb,
		hmac:   hmac,
		emailRegex: regexp.MustCompile(
			`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,16}$`,
		),
	}
}

// bcryptPassword is a validation helper that hashes a user's
// password with an app-wide pepper and bcrypt, which salts the password.
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

// passwordRequired validates if a password is present in userValidator.Create.
func (uv *userValidator) passwordRequired(user *User) error {
	if user.Password == "" {
		return ErrPasswordRequired
	}

	return nil
}

// passwordHashRequired validates if a password hash is present in
// userValidator.Create and userValidator.Update.
// This method comes after the password is generated by userValidator.bcryptPassword.
func (uv *userValidator) passwordHashRequired(user *User) error {
	if user.PasswordHash == "" {
		return ErrPasswordRequired
	}

	return nil
}

func (uv *userValidator) passwordMinLength(user *User) error {
	// If password has not changed, it should be an empty string
	if user.Password == "" {
		return nil
	}

	if len(user.Password) < 8 {
		return ErrPasswordTooShort
	}

	return nil
}

// hmacRemember is a validation helper function to be
// consumed by userValidator.ByRemember.
func (uv *userValidator) hmacRemember(user *User) error {
	if user.Remember == "" {
		return nil
	}
	user.RememberHash = uv.hmac.Hash(user.Remember)
	return nil
}

// setRememberIfUnset is a validation helper that sets the user's remember
// token if it is not set yet.
func (uv *userValidator) setRememberIfUnset(user *User) error {
	if user.Remember != "" {
		return nil
	}

	token, err := rand.RememberToken()
	if err != nil {
		return err
	}

	user.Remember = token
	return nil
}

// rememberHashRequired checks that remember hash token is present.
func (uv *userValidator) rememberHashRequired(user *User) error {
	if user.RememberHash == "" {
		return ErrRememberRequired
	}

	return nil
}

// rememberMinBytes checks that the remember token is at least 32 bytes
func (uv *userValidator) rememberMinBytes(user *User) error {
	if user.Remember == "" {
		return nil
	}

	n, err := rand.NBytes(user.Remember)
	if err != nil {
		return err
	}

	if n < 32 {
		return ErrRememberTooShort
	}

	return nil
}

// requireEmail checks if an email is provided when creating a user.
func (uv *userValidator) requireEmail(user *User) error {
	if user.Email == "" {
		return ErrEmailRequired
	}

	return nil
}

// normalizeEmail normalizes the email address provided by a user.
func (uv *userValidator) normalizeEmail(user *User) error {
	user.Email = strings.ToLower(user.Email)
	user.Email = strings.TrimSpace(user.Email)
	return nil
}

// emailFormat checks if a user's email address complies to the specified format.
func (uv *userValidator) emailFormat(user *User) error {
	if user.Email == "" {
		return nil
	}

	if !uv.emailRegex.MatchString(user.Email) {
		return ErrEmailInvalid
	}

	return nil
}

// emailIsAvail checks if an email address for update or create is already registered.
func (uv *userValidator) emailIsAvail(user *User) error {
	existing, err := uv.ByEmail(user.Email)
	if err == ErrNotFound {
		// Email address is available if no user has that email address
		return nil
	}

	// If another error, return it
	if err != nil {
		return err
	}

	// No email found from query, check if it is the same user or a conflicting user
	if user.ID != existing.ID {
		return ErrEmailTaken
	}

	return nil
}

func (uv *userValidator) idGreaterThan(n uint) userValFn {
	return userValFn(func(user *User) error {
		if user.ID <= n {
			return ErrIDInvalid
		}
		return nil
	})
}

// ByEmail normalizes an email address before passing it onto
// the database layer to perform the query.
func (uv *userValidator) ByEmail(email string) (*User, error) {
	user := User{
		Email: email,
	}

	err := runUserValFns(&user, uv.normalizeEmail)
	if err != nil {
		return nil, err
	}

	return uv.UserDB.ByEmail(user.Email)
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
	err := runUserValFns(
		user,
		uv.passwordRequired,
		uv.passwordMinLength,
		uv.bcryptPassword,
		uv.passwordHashRequired,
		uv.setRememberIfUnset,
		uv.rememberMinBytes,
		uv.hmacRemember,
		uv.rememberHashRequired,
		uv.normalizeEmail,
		uv.requireEmail, // Use after normalizeEmail in case email is whitespace " "
		uv.emailFormat,
		uv.emailIsAvail,
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
		uv.passwordMinLength,
		uv.bcryptPassword,
		uv.passwordHashRequired,
		uv.rememberMinBytes,
		uv.hmacRemember,
		uv.rememberHashRequired,
		uv.normalizeEmail,
		uv.requireEmail,
		uv.emailFormat,
		uv.emailIsAvail,
	)
	if err != nil {
		return err
	}

	return uv.UserDB.Update(user)
}

// Delete deletes the user with the provided ID
func (uv *userValidator) Delete(id uint) error {
	var user User
	user.ID = id

	err := runUserValFns(&user, uv.idGreaterThan(0))
	if err != nil {
		return err
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
// If the password provided is invalid, return nil, ErrPasswordIncorrect
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
		return nil, ErrPasswordIncorrect
	default:
		return nil, err
	}
}
