// user management and authentication

package main

import (
	"errors"
	"github.com/alexedwards/argon2id"
	"github.com/nbutton23/zxcvbn-go"
	"log"
	"regexp"
	"sync"
)

// TODO: Add tests
// TODO: Password hashing security

var (
	ErrPasswordLength   = errors.New("password must contain between 8 and 256 characters")
	ErrPasswordStrength = errors.New("password is too weak")
	ErrExist            = errors.New("user already exists")
	ErrNotExist         = errors.New("user does not exist")
	ErrInvalidUsername  = errors.New("invalid username")
	ErrAuthFailed       = errors.New("invalid username or password")
)

// Username must start with a lowercase letter and contain only lowercase letters, numbers, hyphens and underscores.
var usernameRules *regexp.Regexp = regexp.MustCompile("^[a-z][a-z0-9_-]+$")

type User struct {
	Username     string
	PasswordHash string
}

func (u *User) CheckPassword(password string) bool {
	if u.PasswordHash == "" {
		log.Printf("Tried to check the password of an unitialized user \"%s\"", u.Username)
		return false
	}

	match, err := argon2id.ComparePasswordAndHash(password, u.PasswordHash)
	if err != nil {
		log.Printf("Error checking password for user \"%s\": %v", u.Username, err)
		return false
	}
	return match
}

func (u *User) SetPassword(password string) error {
	if len(password) < 8 || len(password) > 256 {
		return ErrPasswordLength
	} else if zxcvbn.PasswordStrength(password, []string{u.Username}).Score < 3 {
		return ErrPasswordStrength
	}

	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		log.Printf("Error hashing password for user \"%s\": %v", u.Username, err)
		return err
	}
	u.PasswordHash = hash
	log.Printf("Password set for user \"%s\"", u.Username)
	return nil
}

// NewUser returns an error if the username is invalid. The password has to be set using User.SetPassword.
func NewUser(username string) (User, error) {
	if username == "" || len(username) < 2 || len(username) > 30 || !usernameRules.MatchString(username) {
		return User{}, ErrInvalidUsername
	}
	return User{Username: username}, nil
}

type Users struct {
	List []User
	mu   sync.RWMutex
}

// GetUser returns an error if there is no user with such username.
func (users *Users) GetUser(username string) (User, error) {
	users.mu.RLock()
	defer users.mu.RUnlock()

	for _, u := range users.List {
		if u.Username == username {
			return u, nil
		}
	}
	return User{}, ErrNotExist
}

func (users *Users) GetAllUsernames() (names []string) {
	users.mu.RLock()
	defer users.mu.RUnlock()

	for _, u := range users.List {
		names = append(names, u.Username)
	}
	return
}

func (users *Users) CheckPassword(username string, password string) bool {
	users.mu.RLock()
	defer users.mu.RUnlock()

	var user *User = nil
	for _, u := range users.List {
		if u.Username == username {
			user = &u
		}
	}

	if user != nil {
		return user.CheckPassword(password)
	}
	return false
}

// TODO: Caller has to initialize storage for the new user.
func (users *Users) AddUser(username string, password string) error {
	users.mu.Lock()
	defer users.mu.Unlock()

	for _, u := range users.List {
		if u.Username == username {
			return ErrExist
		}
	}

	u, err := NewUser(username)
	if err != nil {
		return err
	}
	err = u.SetPassword(password)
	if err != nil {
		return err
	}

	users.List = append(users.List, u)
	log.Printf("Added user \"%s\"", u.Username)
	return nil
}

func (users *Users) ChangePassword(username string, old string, new string) error {
	users.mu.Lock()
	defer users.mu.Unlock()

	i := -1
	for k, u := range users.List {
		if u.Username == username {
			i = k
			break
		}
	}
	if i == -1 {
		return ErrNotExist
	}

	if !users.List[i].CheckPassword(old) {
		return ErrAuthFailed
	} else {
		return users.List[i].SetPassword(new)
	}
}

func (users *Users) DeleteUser(username string) error {
	users.mu.Lock()
	defer users.mu.Unlock()

	i := -1
	for k, u := range users.List {
		if u.Username == username {
			i = k
			break
		}
	}
	if i == -1 {
		return ErrNotExist
	}

	last := len(users.List) - 1
	users.List[i] = users.List[last]
	users.List = users.List[:last]

	// TODO: Cleanup everything, maybe leaving some backup
	
	return nil
}
