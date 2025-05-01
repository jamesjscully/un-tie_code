package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jamesjscully/un-tie_code/src/api/handlers"
	"github.com/jamesjscully/un-tie_code/src/api/middleware"
	"github.com/jamesjscully/un-tie_code/src/api/models"
	"github.com/jamesjscully/un-tie_code/src/api/repositories"
	"github.com/jamesjscully/un-tie_code/src/api/services"
)

// setupTestRouter creates a test router with all the necessary middleware and handlers
func setupTestRouter() (*gin.Engine, *handlers.Handler, models.AuthService, models.UserRepository) {
	// Use test mode to disable logger output
	gin.SetMode(gin.TestMode)

	// Create repositories
	userRepo := repositories.NewMemoryUserRepository()
	projectRepo := repositories.NewMemoryProjectRepository()

	// Create services
	authService := services.NewAuthService(userRepo)
	projectService := services.NewProjectService(projectRepo)

	// Create handlers
	h := handlers.NewHandler(projectService, authService)

	// Setup router
	r := gin.Default()

	// Add trace ID middleware
	r.Use(func(c *gin.Context) {
		c.Set("traceID", "test-trace-id")
		c.Next()
	})

	// Add session middleware
	r.Use(middleware.SessionMiddleware(authService))

	// Configure routes
	r.LoadHTMLGlob("../../web/templates/*")

	// Auth routes (public)
	auth := r.Group("/auth")
	{
		auth.GET("/login", h.LoginPage)
		auth.POST("/login", h.Login)
		auth.GET("/logout", h.Logout)
	}

	// Protected routes
	protected := r.Group("/")
	protected.Use(middleware.RequireAuth())
	{
		protected.GET("/", h.HomeHandler)
		protected.GET("/projects", h.ListProjects)
	}

	return r, h, authService, userRepo
}

// extractSessionCookie extracts the session cookie from a response
func extractSessionCookie(res *httptest.ResponseRecorder) (*http.Cookie, error) {
	cookies := res.Result().Cookies()
	for _, cookie := range cookies {
		if cookie.Name == "session" {
			return cookie, nil
		}
	}
	return nil, http.ErrNoCookie
}

// TestLoginSuccess tests successful login flow
func TestLoginSuccess(t *testing.T) {
	// Setup
	router, _, _, _ := setupTestRouter()

	// Test login with test@untie.me (should work with any password)
	form := url.Values{}
	form.Add("email", "test@untie.me")
	form.Add("password", "any-password")

	req := httptest.NewRequest("POST", "/auth/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(res, req)

	// Assert
	if res.Code != http.StatusSeeOther {
		t.Errorf("Expected status %d; got %d", http.StatusSeeOther, res.Code)
	}

	// Check for redirect to home page
	location := res.Header().Get("Location")
	if location != "/" {
		t.Errorf("Expected redirect to /; got %s", location)
	}

	// Check that a session cookie was set
	cookie, err := extractSessionCookie(res)
	if err != nil {
		t.Errorf("Session cookie not set: %v", err)
	}
	if cookie.Value == "" {
		t.Error("Session cookie value is empty")
	}
}

// TestLoginFailure tests failed login
func TestLoginFailure(t *testing.T) {
	// Setup
	router, _, _, _ := setupTestRouter()

	// Test login with non-existent user
	form := url.Values{}
	form.Add("email", "nonexistent@example.com")
	form.Add("password", "wrong-password")

	req := httptest.NewRequest("POST", "/auth/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(res, req)

	// Assert
	if res.Code != http.StatusSeeOther {
		t.Errorf("Expected status %d; got %d", http.StatusSeeOther, res.Code)
	}

	// Check for redirect to login page with error
	location := res.Header().Get("Location")
	if !strings.HasPrefix(location, "/auth/login?error=") {
		t.Errorf("Expected redirect to /auth/login?error=...; got %s", location)
	}
}

// TestProtectedRouteWithAuth tests accessing a protected route with authentication
func TestProtectedRouteWithAuth(t *testing.T) {
	// Setup
	router, _, authService, userRepo := setupTestRouter()

	// Create a test user
	user := models.NewUser("test@example.com", "Test User")
	// Important: Actually store the user in the repository
	err := userRepo.Create(user)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	
	token, err := authService.GenerateSessionToken(user)
	if err != nil {
		t.Fatalf("Failed to generate session token: %v", err)
	}

	// Make request to protected route
	req := httptest.NewRequest("GET", "/projects", nil)
	res := httptest.NewRecorder()

	// Add session cookie
	req.AddCookie(&http.Cookie{
		Name:  "session",
		Value: token,
	})

	// Execute request
	router.ServeHTTP(res, req)

	// Assert successful access
	if res.Code != http.StatusOK {
		t.Errorf("Expected status %d; got %d", http.StatusOK, res.Code)
	}
}

// TestProtectedRouteWithoutAuth tests accessing a protected route without authentication
func TestProtectedRouteWithoutAuth(t *testing.T) {
	// Setup
	router, _, _, _ := setupTestRouter()

	// Make request to protected route without auth
	req := httptest.NewRequest("GET", "/projects", nil)
	res := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(res, req)

	// Assert redirect to login
	if res.Code != http.StatusFound {
		t.Errorf("Expected status %d; got %d", http.StatusFound, res.Code)
	}

	location := res.Header().Get("Location")
	if location != "/auth/login" {
		t.Errorf("Expected redirect to /auth/login; got %s", location)
	}
}

// TestLogout tests the logout functionality
func TestLogout(t *testing.T) {
	// Setup
	router, _, authService, userRepo := setupTestRouter()

	// Create a test user and session
	user := models.NewUser("test@example.com", "Test User")
	// Important: Store the user in the repository
	err := userRepo.Create(user)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	
	token, err := authService.GenerateSessionToken(user)
	if err != nil {
		t.Fatalf("Failed to generate session token: %v", err)
	}

	// Make logout request
	req := httptest.NewRequest("GET", "/auth/logout", nil)
	res := httptest.NewRecorder()

	// Add session cookie
	req.AddCookie(&http.Cookie{
		Name:  "session",
		Value: token,
	})

	// Execute request
	router.ServeHTTP(res, req)

	// Assert redirect to login page
	if res.Code != http.StatusSeeOther {
		t.Errorf("Expected status %d; got %d", http.StatusSeeOther, res.Code)
	}

	location := res.Header().Get("Location")
	if location != "/auth/login" {
		t.Errorf("Expected redirect to /auth/login; got %s", location)
	}

	// Check that the session cookie was cleared
	cookie, err := extractSessionCookie(res)
	if err != nil {
		t.Errorf("Expected session cookie to be set (for clearing): %v", err)
	}
	if cookie.MaxAge != -1 {
		t.Errorf("Expected cookie to be cleared with MaxAge -1; got %d", cookie.MaxAge)
	}

	// Verify session was invalidated by trying to access protected route
	req = httptest.NewRequest("GET", "/projects", nil)
	res = httptest.NewRecorder()
	req.AddCookie(&http.Cookie{
		Name:  "session",
		Value: token,
	})

	router.ServeHTTP(res, req)

	// Should be redirected to login
	if res.Code != http.StatusFound {
		t.Errorf("Expected status %d after logout; got %d", http.StatusFound, res.Code)
	}
}

// TestSessionExpiry tests behavior when session token is expired/invalid
func TestSessionExpiry(t *testing.T) {
	// Setup
	router, _, _, _ := setupTestRouter()

	// Make request with invalid session
	req := httptest.NewRequest("GET", "/projects", nil)
	res := httptest.NewRecorder()

	// Add invalid session cookie
	req.AddCookie(&http.Cookie{
		Name:  "session",
		Value: "invalid-session-token",
	})

	// Execute request
	router.ServeHTTP(res, req)

	// Assert redirect to login
	if res.Code != http.StatusFound {
		t.Errorf("Expected status %d; got %d", http.StatusFound, res.Code)
	}

	location := res.Header().Get("Location")
	if location != "/auth/login" {
		t.Errorf("Expected redirect to /auth/login; got %s", location)
	}

	// Check that the invalid session cookie was cleared
	cookie, err := extractSessionCookie(res)
	if err != nil {
		t.Errorf("Expected session cookie to be set (for clearing): %v", err)
	}
	if cookie.MaxAge != -1 {
		t.Errorf("Expected cookie to be cleared with MaxAge -1; got %d", cookie.MaxAge)
	}
}
