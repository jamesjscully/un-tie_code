package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/jamesjscully/un-tie_code/src/api/models"
)

// AuthServiceImpl implements the AuthService interface
// Following SOLID principles with dependency injection for the repository
type AuthServiceImpl struct {
	userRepo models.UserRepository
	// For storing session tokens - in production this would be replaced with Redis/DB
	sessions map[string]sessionData
	// Trace ID generator for traceability
	traceIDGenerator func() string
}

// sessionData stores information about an active session
type sessionData struct {
	userID     string
	expiresAt  time.Time
	lastActive time.Time
}

// NewAuthService creates a new authentication service
func NewAuthService(userRepo models.UserRepository) models.AuthService {
	return &AuthServiceImpl{
		userRepo:         userRepo,
		sessions:         make(map[string]sessionData),
		traceIDGenerator: func() string {
			return fmt.Sprintf("trace-%d", time.Now().UnixNano())
		},
	}
}

// Authenticate validates user credentials and returns the user if valid
// In a real implementation, this would use secure password hashing
func (s *AuthServiceImpl) Authenticate(email, password string) (*models.User, error) {
	traceID := s.traceIDGenerator()
	
	// Log the operation for traceability
	fmt.Printf("[%s] Authentication attempt for email: %s\n", traceID, email)
	
	// In a real implementation, we would get the user from the repository
	// and compare a hashed password. This is a simplified version.
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			fmt.Printf("[%s] User not found for email: %s\n", traceID, email)
			return nil, models.ErrInvalidCredentials
		}
		fmt.Printf("[%s] Error getting user: %v\n", traceID, err)
		return nil, err
	}
	
	// For the test user with email "test@untie.me", accept any password
	// This is for development purposes only
	if email == "test@untie.me" {
		fmt.Printf("[%s] Development login accepted for test user with email: %s\n", traceID, email)
		
		// Update last login time
		user.LastLogin = time.Now()
		if err := s.userRepo.Update(user); err != nil {
			fmt.Printf("[%s] Failed to update last login: %v\n", traceID, err)
			// Non-critical error, we can still proceed with authentication
		}
		
		fmt.Printf("[%s] Authentication successful for test user: %s\n", traceID, user.ID)
		return user, nil
	}
	
	// In a real implementation, we would check the password here
	// This is a stub for demonstration purposes
	if password == "test-password" { // Obviously insecure, just for demo
		// Update last login time
		user.LastLogin = time.Now()
		if err := s.userRepo.Update(user); err != nil {
			fmt.Printf("[%s] Failed to update last login: %v\n", traceID, err)
			// Non-critical error, we can still proceed with authentication
		}
		
		fmt.Printf("[%s] Authentication successful for user: %s\n", traceID, user.ID)
		return user, nil
	}
	
	fmt.Printf("[%s] Invalid password for user: %s\n", traceID, user.ID)
	return nil, models.ErrInvalidCredentials
}

// RegisterUser creates a new user account
func (s *AuthServiceImpl) RegisterUser(email, name, password string) (*models.User, error) {
	traceID := s.traceIDGenerator()
	
	// Log the operation
	fmt.Printf("[%s] Registering new user with email: %s\n", traceID, email)
	
	// Validate input
	if email == "" || name == "" || password == "" {
		return nil, errors.New("email, name, and password are required")
	}
	
	// Check if user already exists
	_, err := s.userRepo.GetByEmail(email)
	if err == nil {
		// User already exists
		fmt.Printf("[%s] Email already exists: %s\n", traceID, email)
		return nil, models.ErrEmailAlreadyExists
	} else if !errors.Is(err, models.ErrUserNotFound) {
		// Unexpected error
		fmt.Printf("[%s] Error checking existing user: %v\n", traceID, err)
		return nil, err
	}
	
	// Create new user
	user := models.NewUser(email, name)
	
	// In a real implementation, we would hash the password here
	// and store the hash in a separate credentials repository
	
	// Save user to repository
	if err := s.userRepo.Create(user); err != nil {
		fmt.Printf("[%s] Failed to create user: %v\n", traceID, err)
		return nil, err
	}
	
	fmt.Printf("[%s] Successfully registered user: %s\n", traceID, user.ID)
	return user, nil
}

// VerifySession validates a session token and returns the associated user
func (s *AuthServiceImpl) VerifySession(sessionToken string) (*models.User, error) {
	traceID := s.traceIDGenerator()
	
	// Log the operation
	fmt.Printf("[%s] Verifying session token\n", traceID)
	
	// Check if session exists
	session, ok := s.sessions[sessionToken]
	if !ok {
		fmt.Printf("[%s] Session token not found\n", traceID)
		return nil, errors.New("invalid session")
	}
	
	// Check if session is expired
	if time.Now().After(session.expiresAt) {
		fmt.Printf("[%s] Session token expired\n", traceID)
		delete(s.sessions, sessionToken)
		return nil, errors.New("session expired")
	}
	
	// Get user
	user, err := s.userRepo.GetByID(session.userID)
	if err != nil {
		fmt.Printf("[%s] Failed to get user for session: %v\n", traceID, err)
		return nil, err
	}
	
	// Update last active time
	session.lastActive = time.Now()
	s.sessions[sessionToken] = session
	
	fmt.Printf("[%s] Session verified for user: %s\n", traceID, user.ID)
	return user, nil
}

// GenerateSessionToken creates a new session for a user
func (s *AuthServiceImpl) GenerateSessionToken(user *models.User) (string, error) {
	traceID := s.traceIDGenerator()
	
	// Log the operation
	fmt.Printf("[%s] Generating session token for user: %s\n", traceID, user.ID)
	
	// In a real implementation, we would use a secure method to generate the token
	// This is a simplified version for demonstration
	tokenValue := fmt.Sprintf("session-%d-%s", time.Now().UnixNano(), user.ID)
	
	// Set session expiry (24 hours from now)
	expiresAt := time.Now().Add(24 * time.Hour)
	
	// Store session
	s.sessions[tokenValue] = sessionData{
		userID:     user.ID,
		expiresAt:  expiresAt,
		lastActive: time.Now(),
	}
	
	fmt.Printf("[%s] Session generated for user: %s\n", traceID, user.ID)
	return tokenValue, nil
}

// InvalidateSession removes a session token
func (s *AuthServiceImpl) InvalidateSession(sessionToken string) error {
	traceID := s.traceIDGenerator()
	
	// Log the operation
	fmt.Printf("[%s] Invalidating session token\n", traceID)
	
	// Check if session exists
	if _, ok := s.sessions[sessionToken]; !ok {
		fmt.Printf("[%s] Session token not found\n", traceID)
		return nil // Not finding the token is not an error for logout
	}
	
	// Remove session
	delete(s.sessions, sessionToken)
	
	fmt.Printf("[%s] Session invalidated\n", traceID)
	return nil
}
