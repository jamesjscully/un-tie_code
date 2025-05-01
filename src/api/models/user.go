package models

import (
	"errors"
	"time"

	"github.com/jamesjscully/un-tie_code/src/api/utils"
)

// User represents a user of the Un-tie.me code system
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	LastLogin time.Time `json:"lastLogin"`
}

// Clone creates a deep copy of the User
func (u *User) Clone() *User {
	if u == nil {
		return nil
	}
	
	return &User{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
		LastLogin: u.LastLogin,
	}
}

// Authentication errors
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
)

// UserRepository defines the data access interface for users
type UserRepository interface {
	GetByID(id string) (*User, error)
	GetByEmail(email string) (*User, error)
	Create(user *User) error
	Update(user *User) error
	Delete(id string) error
}

// AuthService defines the authentication operations
// Using interface for dependency inversion
type AuthService interface {
	Authenticate(email, password string) (*User, error)
	RegisterUser(email, name, password string) (*User, error)
	VerifySession(sessionToken string) (*User, error)
	GenerateSessionToken(user *User) (string, error)
	InvalidateSession(sessionToken string) error
}

// NewUser creates a new user with proper initialization
func NewUser(email, name string) *User {
	now := time.Now()
	return &User{
		ID:        utils.GenerateID(),
		Email:     email,
		Name:      name,
		Role:      "user", // Default role
		CreatedAt: now,
		UpdatedAt: now,
	}
}
