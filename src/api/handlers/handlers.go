package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jamesjscully/un-tie_code/src/api/models"
)

// Handler contains all request handlers and dependencies
type Handler struct {
	projectService models.ProjectService
	authService    models.AuthService
}

// NewHandler creates a new handler with injected dependencies
func NewHandler(projectService models.ProjectService, authService models.AuthService) *Handler {
	return &Handler{
		projectService: projectService,
		authService:    authService,
	}
}

// HomeHandler renders the main dashboard page
func (h *Handler) HomeHandler(c *gin.Context) {
	// Get list of recent projects for the dashboard
	// Using a consistent trace ID for better traceability
	traceID, _ := c.Get("traceID")
	userID := "test-user" // Placeholder, will be from auth session
	
	var recentProjects []map[string]interface{}
	
	// Try to get some projects if they exist, with proper error handling
	projects, err := h.projectService.ListProjects(userID)
	if err != nil {
		// Log the error but continue - fail gracefully
		fmt.Printf("[%s] Error getting projects for homepage: %v\n", traceID, err)
	} else if len(projects) > 0 {
		// Just get up to 3 projects for the dashboard
		count := min(len(projects), 3)
		for i := 0; i < count; i++ {
			recentProjects = append(recentProjects, map[string]interface{}{
				"id":          projects[i].ID,
				"name":        projects[i].Name,
				"description": projects[i].Description,
				"updatedAt":   projects[i].UpdatedAt.Format(time.RFC1123),
			})
		}
	}
	
	// Always use base template, ensuring deterministic rendering
	c.HTML(http.StatusOK, "base", gin.H{
		"title":          "Un-tie.me code",
		"recentProjects": recentProjects,
	})
}

// Helper function to get minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// NewProjectForm renders the form for creating a new project
func (h *Handler) NewProjectForm(c *gin.Context) {
	c.HTML(http.StatusOK, "base", gin.H{
		"title": "Create New Project",
	})
}

// ListProjects returns all projects for the authenticated user
func (h *Handler) ListProjects(c *gin.Context) {
	// Get trace ID for request tracing
	traceID, _ := c.Get("traceID")
	
	// TODO: Get user ID from authenticated session
	userID := "test-user" // Placeholder
	
	projects, err := h.projectService.ListProjects(userID)
	if err != nil {
		fmt.Printf("[%s] Error listing projects: %v\n", traceID, err)
		c.HTML(http.StatusInternalServerError, "base", gin.H{
			"title": "Error",
			"error": "Failed to retrieve projects",
		})
		return
	}
	
	// Convert to a format suitable for templates
	var projectsData []map[string]interface{}
	for _, project := range projects {
		projectsData = append(projectsData, map[string]interface{}{
			"id":          project.ID,
			"name":        project.Name,
			"description": project.Description,
			"updatedAt":   project.UpdatedAt.Format(time.RFC1123),
		})
	}
	
	c.HTML(http.StatusOK, "base", gin.H{
		"title":    "Your Projects",
		"projects": projectsData,
	})
}

// GetProject returns details for a specific project
func (h *Handler) GetProject(c *gin.Context) {
	traceID, _ := c.Get("traceID")
	id := c.Param("id")
	
	project, err := h.projectService.GetProject(id)
	if err != nil {
		fmt.Printf("[%s] Error getting project %s: %v\n", traceID, id, err)
		c.HTML(http.StatusNotFound, "base", gin.H{
			"title": "Project Not Found",
			"error": "The requested project could not be found",
		})
		return
	}
	
	c.HTML(http.StatusOK, "base", gin.H{
		"title":       "Project Details",
		"projectID":   project.ID,
		"projectName": project.Name,
		"description": project.Description,
		"createdAt":   project.CreatedAt.Format(time.RFC1123),
		"updatedAt":   project.UpdatedAt.Format(time.RFC1123),
	})
}

// CreateProject handles the creation of a new project
func (h *Handler) CreateProject(c *gin.Context) {
	traceID, _ := c.Get("traceID")
	
	// Parse form data
	name := c.PostForm("name")
	description := c.PostForm("description")
	
	// TODO: Get user ID from authenticated session
	userID := "test-user" // Placeholder
	
	// Basic validation
	if name == "" {
		c.HTML(http.StatusBadRequest, "base", gin.H{
			"title":       "Create New Project",
			"error":       "Project name is required",
			"description": description, // Return entered data
		})
		return
	}
	
	// Create project
	project := models.NewProject(name, description, userID)
	
	err := h.projectService.CreateProject(project)
	if err != nil {
		fmt.Printf("[%s] Error creating project: %v\n", traceID, err)
		c.HTML(http.StatusInternalServerError, "base", gin.H{
			"title": "Error",
			"error": "Failed to create project",
		})
		return
	}
	
	// Redirect to project page
	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/projects/%s", project.ID))
}

// UpdateProject handles updates to an existing project
func (h *Handler) UpdateProject(c *gin.Context) {
	traceID, _ := c.Get("traceID")
	id := c.Param("id")
	
	// Get existing project
	project, err := h.projectService.GetProject(id)
	if err != nil {
		fmt.Printf("[%s] Error getting project for update %s: %v\n", traceID, id, err)
		c.JSON(http.StatusNotFound, gin.H{
			"status": "error",
			"error":  "Project not found",
		})
		return
	}
	
	// Parse update data
	var updateData struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	
	if err := c.BindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Invalid data format",
		})
		return
	}
	
	// Update fields
	if updateData.Name != "" {
		project.Name = updateData.Name
	}
	if updateData.Description != "" {
		project.Description = updateData.Description
	}
	
	// Save changes
	err = h.projectService.UpdateProject(project)
	if err != nil {
		fmt.Printf("[%s] Error updating project %s: %v\n", traceID, id, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to update project",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status": "updated",
		"id":     id,
	})
}

// DeleteProject handles deletion of a project
func (h *Handler) DeleteProject(c *gin.Context) {
	traceID, _ := c.Get("traceID")
	id := c.Param("id")
	
	err := h.projectService.DeleteProject(id)
	if err != nil {
		fmt.Printf("[%s] Error deleting project %s: %v\n", traceID, id, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to delete project",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status": "deleted",
		"id":     id,
	})
}

// LoginPage renders the login page
func (h *Handler) LoginPage(c *gin.Context) {
	traceID, _ := c.Get("traceID")
	fmt.Printf("[%s] Rendering login page\n", traceID)
	
	// Check if there's an error message to display (e.g., from a failed login)
	errorMsg := c.Query("error")
	
	// Check if user is already authenticated
	if user := h.getCurrentUser(c); user != nil {
		fmt.Printf("[%s] User already authenticated, redirecting to home\n", traceID)
		c.Redirect(http.StatusFound, "/")
		return
	}
	
	// For a more deterministic rendering, use a dedicated auth template
	// This avoids conflicts with the content template logic
	c.HTML(http.StatusOK, "auth.html", gin.H{
		"title": "Login",
		"error": errorMsg,
	})
}

// Login processes the login attempt
func (h *Handler) Login(c *gin.Context) {
	traceID, _ := c.Get("traceID")
	
	email := c.PostForm("email")
	password := c.PostForm("password")
	rememberMe := c.PostForm("remember-me") == "on"
	
	// Basic form validation with traceability
	if email == "" || password == "" {
		fmt.Printf("[%s] Login attempt with missing credentials\n", traceID)
		c.HTML(http.StatusBadRequest, "base", gin.H{
			"title": "Login",
			"error": "Email and password are required",
			"email": email, // Return email for convenience
		})
		return
	}
	
	// Authenticate user
	user, err := h.authService.Authenticate(email, password)
	if err != nil {
		fmt.Printf("[%s] Authentication failed for %s: %v\n", traceID, email, err)
		c.HTML(http.StatusUnauthorized, "base", gin.H{
			"title": "Login",
			"error": "Invalid email or password",
			"email": email, // Return email for convenience
		})
		return
	}
	
	// Generate session token
	token, err := h.authService.GenerateSessionToken(user)
	if err != nil {
		fmt.Printf("[%s] Failed to generate session for %s: %v\n", traceID, user.ID, err)
		c.HTML(http.StatusInternalServerError, "base", gin.H{
			"title": "Error",
			"error": "Failed to create session",
		})
		return
	}
	
	// Set session cookie
	// Calculate expiration time (24 hours or 30 days if remember-me is checked)
	var maxAge int
	if rememberMe {
		maxAge = 60 * 60 * 24 * 30 // 30 days
		fmt.Printf("[%s] Setting long-lived session for %s (remember me)\n", traceID, user.ID)
	} else {
		maxAge = 60 * 60 * 24 // 24 hours
		fmt.Printf("[%s] Setting standard session for %s\n", traceID, user.ID)
	}
	
	c.SetCookie(
		"session_token",
		token,
		maxAge,
		"/",
		"",
		false, // secure (should be true in production with HTTPS)
		true,  // httpOnly
	)
	
	fmt.Printf("[%s] Login successful for user: %s (%s)\n", traceID, user.ID, user.Email)
	
	// Check if there's a return_to cookie for redirection after login
	returnTo, err := c.Cookie("return_to")
	if err == nil && returnTo != "" {
		// Clear the return_to cookie
		c.SetCookie("return_to", "", -1, "/", "", false, true)
		
		// Validate the return URL (security measure to prevent open redirect)
		if strings.HasPrefix(returnTo, "/") && !strings.Contains(returnTo, "://") {
			fmt.Printf("[%s] Redirecting to: %s\n", traceID, returnTo)
			c.Redirect(http.StatusSeeOther, returnTo)
			return
		}
	}
	
	// Default redirect to home page
	c.Redirect(http.StatusSeeOther, "/")
}

// Logout handles user logout
func (h *Handler) Logout(c *gin.Context) {
	traceID, _ := c.Get("traceID")
	
	// Get the session token from cookie
	sessionToken, err := c.Cookie("session_token")
	if err == nil {
		// Invalidate the session in the auth service
		if err := h.authService.InvalidateSession(sessionToken); err != nil {
			fmt.Printf("[%s] Error invalidating session: %v\n", traceID, err)
			// Continue with logout even if invalidation fails
		}
	}
	
	// Clear the session cookie
	c.SetCookie("session_token", "", -1, "/", "", false, true)
	
	fmt.Printf("[%s] User logged out\n", traceID)
	c.Redirect(http.StatusSeeOther, "/auth/login")
}

// Helper function to get current user from context
func (h *Handler) getCurrentUser(c *gin.Context) *models.User {
	user, exists := c.Get("user")
	if !exists {
		return nil
	}
	
	if u, ok := user.(*models.User); ok {
		return u
	}
	
	return nil
}

// Feature-specific handlers

// ArchitectureCanvas renders the architecture canvas page
func (h *Handler) ArchitectureCanvas(c *gin.Context) {
	id := c.Param("id")
	
	project, err := h.projectService.GetProject(id)
	if err != nil {
		c.HTML(http.StatusNotFound, "base", gin.H{
			"title": "Project Not Found",
			"error": "The requested project could not be found",
		})
		return
	}
	
	c.HTML(http.StatusOK, "base", gin.H{
		"title":       "Architecture Canvas",
		"projectID":   project.ID,
		"projectName": project.Name,
	})
}

// StoryFlow renders the story flow board
func (h *Handler) StoryFlow(c *gin.Context) {
	id := c.Param("id")
	
	project, err := h.projectService.GetProject(id)
	if err != nil {
		c.HTML(http.StatusNotFound, "base", gin.H{
			"title": "Project Not Found",
			"error": "The requested project could not be found",
		})
		return
	}
	
	c.HTML(http.StatusOK, "base", gin.H{
		"title":       "Story Flow Board",
		"projectID":   project.ID,
		"projectName": project.Name,
	})
}

// TaskHub renders the task monitoring page
func (h *Handler) TaskHub(c *gin.Context) {
	id := c.Param("id")
	
	project, err := h.projectService.GetProject(id)
	if err != nil {
		c.HTML(http.StatusNotFound, "base", gin.H{
			"title": "Project Not Found",
			"error": "The requested project could not be found",
		})
		return
	}
	
	c.HTML(http.StatusOK, "base", gin.H{
		"title":       "Task Monitoring Hub",
		"projectID":   project.ID,
		"projectName": project.Name,
	})
}

// ReviewQueue renders the review queue page
func (h *Handler) ReviewQueue(c *gin.Context) {
	id := c.Param("id")
	
	project, err := h.projectService.GetProject(id)
	if err != nil {
		c.HTML(http.StatusNotFound, "base", gin.H{
			"title": "Project Not Found",
			"error": "The requested project could not be found",
		})
		return
	}
	
	c.HTML(http.StatusOK, "base", gin.H{
		"title":       "Review Queue",
		"projectID":   project.ID,
		"projectName": project.Name,
	})
}

// DesignAssistant renders the design assistant page
func (h *Handler) DesignAssistant(c *gin.Context) {
	id := c.Param("id")
	
	project, err := h.projectService.GetProject(id)
	if err != nil {
		c.HTML(http.StatusNotFound, "base", gin.H{
			"title": "Project Not Found",
			"error": "The requested project could not be found",
		})
		return
	}
	
	c.HTML(http.StatusOK, "base", gin.H{
		"title":       "Design Assistant",
		"projectID":   project.ID,
		"projectName": project.Name,
	})
}

// API Handlers

// APIStatus is a simple endpoint to verify API functionality
func (h *Handler) APIStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"version": "0.1.0",
	})
}

// APIListProjects returns all projects for a user as JSON
func (h *Handler) APIListProjects(c *gin.Context) {
	// TODO: Get user ID from authenticated session
	userID := "test-user" // Placeholder
	
	projects, err := h.projectService.ListProjects(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to retrieve projects",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"projects": projects,
	})
}

// APIGetProject returns a single project as JSON
func (h *Handler) APIGetProject(c *gin.Context) {
	id := c.Param("id")
	
	project, err := h.projectService.GetProject(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status": "error",
			"error":  "Project not found",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"project": project,
	})
}

// APICreateProject creates a new project via API
func (h *Handler) APICreateProject(c *gin.Context) {
	var projectData struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}
	
	if err := c.BindJSON(&projectData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Invalid data format",
		})
		return
	}
	
	// TODO: Get user ID from authenticated session
	userID := "test-user" // Placeholder
	
	project := models.NewProject(projectData.Name, projectData.Description, userID)
	
	err := h.projectService.CreateProject(project)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to create project",
		})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"project": project,
	})
}

// APIUpdateProject updates a project via API
func (h *Handler) APIUpdateProject(c *gin.Context) {
	id := c.Param("id")
	
	// Get existing project
	project, err := h.projectService.GetProject(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status": "error",
			"error":  "Project not found",
		})
		return
	}
	
	// Parse update data
	var updateData struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	
	if err := c.BindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Invalid data format",
		})
		return
	}
	
	// Update fields
	if updateData.Name != "" {
		project.Name = updateData.Name
	}
	if updateData.Description != "" {
		project.Description = updateData.Description
	}
	
	// Save changes
	err = h.projectService.UpdateProject(project)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to update project",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"project": project,
	})
}

// APIDeleteProject deletes a project via API
func (h *Handler) APIDeleteProject(c *gin.Context) {
	id := c.Param("id")
	
	err := h.projectService.DeleteProject(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  "Failed to delete project",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"id":     id,
	})
}
