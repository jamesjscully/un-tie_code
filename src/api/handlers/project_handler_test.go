package handlers_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jamesjscully/un-tie_code/src/api/handlers"
	"github.com/jamesjscully/un-tie_code/src/api/middleware"
	"github.com/jamesjscully/un-tie_code/src/api/models"
	"github.com/jamesjscully/un-tie_code/src/api/repositories"
	"github.com/jamesjscully/un-tie_code/src/api/services"
	"github.com/jamesjscully/un-tie_code/src/api/utils"
)

// setupProjectTestRouter creates a test router with all necessary middleware and handlers
// specifically for project-related tests
func setupProjectTestRouter() (*gin.Engine, *handlers.Handler, models.ProjectService, models.AuthService, models.UserRepository) {
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

	// Configure templates for HTML responses
	r.LoadHTMLGlob("../../web/templates/*")

	// API routes for projects
	api := r.Group("/api/v1")
	{
		// Public endpoints
		api.GET("/status", h.APIStatus)
		
		// Protected endpoints
		apiAuth := api.Group("/")
		apiAuth.Use(middleware.RequireAuth())
		{
			projects := apiAuth.Group("/projects")
			{
				projects.GET("/", h.APIListProjects)
				projects.POST("/", h.APICreateProject)
				projects.GET("/:id", h.APIGetProject)
				projects.PUT("/:id", h.APIUpdateProject)
				projects.DELETE("/:id", h.APIDeleteProject)
			}
		}
	}

	// HTML routes for projects
	protected := r.Group("/")
	protected.Use(middleware.RequireAuth())
	{
		projects := protected.Group("/projects")
		{
			projects.GET("/", h.ListProjects)
			projects.POST("/", h.CreateProject)
			projects.GET("/:id", h.GetProject)
			// Other project routes would be here
		}
	}

	return r, h, projectService, authService, userRepo
}

// createTestUser creates a test user and logs them in
func createTestUser(t *testing.T, authService models.AuthService, userRepo models.UserRepository, email string) (string, *models.User) {
	user := models.NewUser(email, "Test User")
	
	// Store the user in the repository so it can be found when verifying the session
	err := userRepo.Create(user)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	
	token, err := authService.GenerateSessionToken(user)
	if err != nil {
		t.Fatalf("Failed to generate session token: %v", err)
	}
	return token, user
}

// createTestProject creates a test project for a user with a unique ID
func createTestProject(t *testing.T, projectService models.ProjectService, userID string, name string) *models.Project {
	// Create a project with a custom ID to avoid collisions in tests
	uniqueID := fmt.Sprintf("test-proj-%s-%d", utils.GenerateID(), time.Now().UnixNano())
	project := &models.Project{
		ID:          uniqueID,
		Name:        name,
		Description: "Test Description",
		UserID:      userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	err := projectService.CreateProject(project)
	if err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}
	return project
}

// TestProjectOwnershipVerification tests that users can only access their own projects
func TestProjectOwnershipVerification(t *testing.T) {
	router, _, projectService, authService, userRepo := setupProjectTestRouter()

	// Create two test users
	token1, user1 := createTestUser(t, authService, userRepo, "user1@example.com")
	token2, _ := createTestUser(t, authService, userRepo, "user2@example.com")

	// Create a project for user1
	project := createTestProject(t, projectService, user1.ID, "User1's Project")

	// Test scenarios:
	// 1. User1 can access their own project
	// 2. User2 cannot access User1's project

	// Test 1: User1 can access their project
	req := httptest.NewRequest("GET", "/api/v1/projects/"+project.ID, nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token1})
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Errorf("Expected status %d for owner; got %d", http.StatusOK, res.Code)
	}

	// Test 2: User2 cannot access User1's project
	req = httptest.NewRequest("GET", "/api/v1/projects/"+project.ID, nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token2})
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code != http.StatusForbidden {
		t.Errorf("Expected status %d for non-owner; got %d", http.StatusForbidden, res.Code)
	}
}

// TestProjectCreation tests project creation flow
func TestProjectCreation(t *testing.T) {
	router, _, _, authService, userRepo := setupProjectTestRouter()

	// Create test user
	token, user := createTestUser(t, authService, userRepo, "creator@example.com")

	// Test project creation via API
	projectJSON := `{"name": "New Project", "description": "API-created project"}`
	
	req := httptest.NewRequest("POST", "/api/v1/projects/", strings.NewReader(projectJSON))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	res := httptest.NewRecorder()
	
	router.ServeHTTP(res, req)

	// Assert
	if res.Code != http.StatusCreated {
		t.Errorf("Expected status %d; got %d", http.StatusCreated, res.Code)
		t.Logf("Response body: %s", res.Body.String())
		return
	}

	// Parse response to get project ID
	var response struct {
		Status  string         `json:"status"`
		Project *models.Project `json:"project"`
	}
	
	err := json.Unmarshal(res.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v - Body: %s", err, res.Body.String())
	}

	if response.Status != "success" {
		t.Errorf("Expected status 'success'; got '%s'", response.Status)
	}

	if response.Project == nil {
		t.Fatalf("Project data missing from response")
	}

	if response.Project.UserID != user.ID {
		t.Errorf("Expected project to be owned by user ID %s; got %s", 
			user.ID, response.Project.UserID)
	}
}

// TestProjectModificationSecurity tests update and delete security
func TestProjectModificationSecurity(t *testing.T) {
	router, _, projectService, authService, userRepo := setupProjectTestRouter()

	// Create two test users
	_, user1 := createTestUser(t, authService, userRepo, "owner@example.com")
	token2, _ := createTestUser(t, authService, userRepo, "attacker@example.com")

	// Create a project for user1
	project := createTestProject(t, projectService, user1.ID, "Protected Project")

	// Test update security: User2 tries to update User1's project
	updateJSON := `{"name": "Hacked Project", "description": "This shouldn't work"}`
	
	req := httptest.NewRequest("PUT", "/api/v1/projects/"+project.ID, strings.NewReader(updateJSON))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "session", Value: token2})
	res := httptest.NewRecorder()
	
	router.ServeHTTP(res, req)

	// Assert forbidden access for update
	if res.Code != http.StatusForbidden {
		t.Errorf("Expected status %d for unauthorized update; got %d", 
			http.StatusForbidden, res.Code)
	}

	// Test delete security: User2 tries to delete User1's project
	req = httptest.NewRequest("DELETE", "/api/v1/projects/"+project.ID, nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token2})
	res = httptest.NewRecorder()
	
	router.ServeHTTP(res, req)

	// Assert forbidden access for delete
	if res.Code != http.StatusForbidden {
		t.Errorf("Expected status %d for unauthorized deletion; got %d", 
			http.StatusForbidden, res.Code)
	}

	// Verify project still exists and is unchanged
	retrievedProject, err := projectService.GetProject(project.ID)
	if err != nil {
		t.Fatalf("Project doesn't exist after failed delete attempt: %v", err)
	}

	if retrievedProject.Name != "Protected Project" {
		t.Errorf("Project was modified despite access being forbidden")
	}
}

// TestProjectListFiltering tests that users only see their own projects
func TestProjectListFiltering(t *testing.T) {
	router, _, projectService, authService, userRepo := setupProjectTestRouter()

	// Create two test users
	token1, user1 := createTestUser(t, authService, userRepo, "alice@example.com")
	token2, user2 := createTestUser(t, authService, userRepo, "bob@example.com")

	// Create projects for both users with unique IDs
	project1 := createTestProject(t, projectService, user1.ID, "Alice Project 1")
	project2 := createTestProject(t, projectService, user1.ID, "Alice Project 2")
	project3 := createTestProject(t, projectService, user2.ID, "Bob Project")

	// Print all projects for debugging
	t.Logf("Created projects - User1: %s has projects %s, %s; User2: %s has project %s", 
		user1.ID, project1.ID, project2.ID, user2.ID, project3.ID)

	// Get user1's projects
	req := httptest.NewRequest("GET", "/api/v1/projects/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token1})
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	// Assert success
	if res.Code != http.StatusOK {
		t.Errorf("Expected status %d; got %d", http.StatusOK, res.Code)
	}

	// Parse response
	var response struct {
		Status   string          `json:"status"`
		Projects []*models.Project `json:"projects"`
	}
	
	err := json.Unmarshal(res.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v - Body: %s", err, res.Body.String())
	}

	// Verify user1 only sees their own projects
	if len(response.Projects) != 2 {
		t.Errorf("Expected 2 projects for user1; got %d", len(response.Projects))
	}

	// Verify that each project belongs to user1
	for _, p := range response.Projects {
		if p == nil {
			t.Errorf("Got nil project in response")
			continue
		}
		
		t.Logf("User1 sees project: %+v", p)
		
		if p.UserID != user1.ID {
			t.Errorf("User1 can see project with ID %s belonging to %s (expected %s)",
				p.ID, p.UserID, user1.ID)
		}
	}

	// Get user2's projects
	req = httptest.NewRequest("GET", "/api/v1/projects/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token2})
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)

	// Assert success
	if res.Code != http.StatusOK {
		t.Errorf("Expected status %d for user2; got %d", http.StatusOK, res.Code)
	}

	// Parse response for user2
	err = json.Unmarshal(res.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response for user2: %v - Body: %s", err, res.Body.String())
	}

	// Verify user2 only sees their own project
	if len(response.Projects) != 1 {
		t.Errorf("Expected 1 project for user2; got %d", len(response.Projects))
	}

	// Check that the single project belongs to user2
	if len(response.Projects) > 0 {
		p := response.Projects[0]
		if p == nil {
			t.Errorf("Got nil project in user2 response")
		} else {
			t.Logf("User2 sees project: %+v", p)
			
			if p.UserID != user2.ID {
				t.Errorf("User2 can see project with ID %s belonging to %s (expected %s)",
					p.ID, p.UserID, user2.ID)
			}
		}
	}
}
