package services

import (
	"testing"

	"github.com/jamesjscully/un-tie_code/src/api/models"
	"github.com/jamesjscully/un-tie_code/src/api/repositories"
)

// Test_AuthService_RegisterUser tests user registration
func Test_AuthService_RegisterUser(t *testing.T) {
	// Arrange
	userRepo := repositories.NewMemoryUserRepository()
	authService := NewAuthService(userRepo)
	
	// Test data
	testEmail := "test@example.com"
	testName := "Test User"
	testPassword := "secure-password"
	
	// Act
	user, err := authService.RegisterUser(testEmail, testName, testPassword)
	
	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if user == nil {
		t.Fatal("Expected user to be returned, got nil")
	}
	
	if user.Email != testEmail {
		t.Fatalf("Expected email %s, got %s", testEmail, user.Email)
	}
	
	if user.Name != testName {
		t.Fatalf("Expected name %s, got %s", testName, user.Name)
	}
	
	if user.ID == "" {
		t.Fatal("Expected user ID to be set")
	}
	
	// Test duplicate email
	_, err = authService.RegisterUser(testEmail, "Another User", "another-password")
	if err == nil {
		t.Fatal("Expected error for duplicate email, got nil")
	}
}

// Test_AuthService_Authenticate tests user authentication
func Test_AuthService_Authenticate(t *testing.T) {
	// Arrange
	userRepo := repositories.NewMemoryUserRepository()
	authService := NewAuthService(userRepo)
	
	// Create a test user first
	testEmail := "test@example.com"
	testName := "Test User"
	
	// Create user directly in repository
	user := models.NewUser(testEmail, testName)
	err := userRepo.Create(user)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	
	// Act - Test successful authentication (using the test password from auth_service.go)
	authenticatedUser, err := authService.Authenticate(testEmail, "test-password")
	
	// Assert
	if err != nil {
		t.Fatalf("Expected no error for valid credentials, got %v", err)
	}
	
	if authenticatedUser == nil {
		t.Fatal("Expected user to be returned, got nil")
	}
	
	if authenticatedUser.Email != testEmail {
		t.Fatalf("Expected email %s, got %s", testEmail, authenticatedUser.Email)
	}
	
	// Test invalid credentials
	_, err = authService.Authenticate(testEmail, "wrong-password")
	if err == nil {
		t.Fatal("Expected error for invalid credentials, got nil")
	}
	
	// Test non-existent user
	_, err = authService.Authenticate("nonexistent@example.com", "any-password")
	if err == nil {
		t.Fatal("Expected error for non-existent user, got nil")
	}
}

// Test_AuthService_SessionManagement tests the session functionality
func Test_AuthService_SessionManagement(t *testing.T) {
	// Arrange
	userRepo := repositories.NewMemoryUserRepository()
	authService := NewAuthService(userRepo)
	
	// Create a test user
	testEmail := "test@example.com"
	testName := "Test User"
	user := models.NewUser(testEmail, testName)
	err := userRepo.Create(user)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	
	// Act - Generate session token
	token, err := authService.GenerateSessionToken(user)
	
	// Assert
	if err != nil {
		t.Fatalf("Expected no error generating token, got %v", err)
	}
	
	if token == "" {
		t.Fatal("Expected token to be non-empty")
	}
	
	// Verify session
	verifiedUser, err := authService.VerifySession(token)
	if err != nil {
		t.Fatalf("Expected no error verifying session, got %v", err)
	}
	
	if verifiedUser == nil {
		t.Fatal("Expected user to be returned, got nil")
	}
	
	if verifiedUser.ID != user.ID {
		t.Fatalf("Expected user ID %s, got %s", user.ID, verifiedUser.ID)
	}
	
	// Invalidate session
	err = authService.InvalidateSession(token)
	if err != nil {
		t.Fatalf("Expected no error invalidating session, got %v", err)
	}
	
	// Verify session is invalidated
	_, err = authService.VerifySession(token)
	if err == nil {
		t.Fatal("Expected error verifying invalidated session, got nil")
	}
}
