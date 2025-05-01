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
	traceID, _ := c.Get("traceID")
	user := h.getCurrentUser(c)
	
	var recentProjects []map[string]interface{}
	
	// Try to get some projects if they exist, with proper error handling
	projects, err := h.projectService.ListProjects(user.ID)
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
	// Get the current user for context - following the pattern of tracing and user context
	traceID, _ := c.Get("traceID")
	user := h.getCurrentUser(c)
	
	// Log the action for traceability
	fmt.Printf("[%s] User %s accessing new project form\n", traceID, user.ID)
	
	// Render using deterministic templates
	c.HTML(http.StatusOK, "base", gin.H{
		"title": "Create New Project",
		"page": "new_project",
		// Add any default values or context needed for the form
		"name": "",
		"description": "",
		// Enable tracing of which template is being rendered
		"traceID": traceID,
	})
}

// ListProjects returns all projects for the authenticated user
func (h *Handler) ListProjects(c *gin.Context) {
	// Get trace ID for request tracing
	traceID, _ := c.Get("traceID")
	user := h.getCurrentUser(c)
	
	projects, err := h.projectService.ListProjects(user.ID)
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
	user := h.getCurrentUser(c)
	
	// Parse form data
	name := c.PostForm("name")
	description := c.PostForm("description")
	
	if name == "" {
		fmt.Printf("[%s] Project creation failed: missing name\n", traceID)
		c.HTML(http.StatusBadRequest, "base", gin.H{
			"title": "Create Project - Error",
			"error": "Project name is required",
		})
		return
	}
	
	project := models.NewProject(name, description, user.ID)
	
	err := h.projectService.CreateProject(project)
	if err != nil {
		fmt.Printf("[%s] Project creation failed: %v\n", traceID, err)
		c.HTML(http.StatusInternalServerError, "base", gin.H{
			"title": "Create Project - Error",
			"error": "Failed to create project",
		})
		return
	}
	
	c.Redirect(http.StatusSeeOther, "/projects/"+project.ID)
}

// UpdateProject handles updates to an existing project
func (h *Handler) UpdateProject(c *gin.Context) {
	traceID, _ := c.Get("traceID")
	id := c.Param("id")
	user := h.getCurrentUser(c)
	
	// Parse form data
	name := c.PostForm("name")
	description := c.PostForm("description")
	
	// Get current project
	project, err := h.projectService.GetProject(id)
	if err != nil {
		fmt.Printf("[%s] Error getting project %s for update: %v\n", traceID, id, err)
		c.HTML(http.StatusNotFound, "base", gin.H{
			"title": "Project Not Found",
			"error": "The requested project could not be found",
		})
		return
	}
	
	// Verify ownership
	if project.UserID != user.ID {
		fmt.Printf("[%s] Unauthorized update attempt on project %s by user %s\n", traceID, id, user.ID)
		c.HTML(http.StatusForbidden, "base", gin.H{
			"title": "Unauthorized",
			"error": "You do not have permission to update this project",
		})
		return
	}
	
	// Update fields
	if name != "" {
		project.Name = name
	}
	if description != "" {
		project.Description = description
	}
	
	// Save changes
	err = h.projectService.UpdateProject(project)
	if err != nil {
		fmt.Printf("[%s] Error updating project %s: %v\n", traceID, id, err)
		c.HTML(http.StatusInternalServerError, "base", gin.H{
			"title": "Update Project - Error",
			"error": "Failed to update project",
		})
		return
	}
	
	c.Redirect(http.StatusSeeOther, "/projects/"+id)
}

// DeleteProject handles deletion of a project
func (h *Handler) DeleteProject(c *gin.Context) {
	traceID, _ := c.Get("traceID")
	id := c.Param("id")
	user := h.getCurrentUser(c)
	
	// Get project to verify ownership
	project, err := h.projectService.GetProject(id)
	if err == nil && project.UserID != user.ID {
		fmt.Printf("[%s] Unauthorized delete attempt on project %s by user %s\n", traceID, id, user.ID)
		c.HTML(http.StatusForbidden, "base", gin.H{
			"title": "Unauthorized",
			"error": "You do not have permission to delete this project",
		})
		return
	}
	
	err = h.projectService.DeleteProject(id)
	if err != nil {
		fmt.Printf("[%s] Error deleting project %s: %v\n", traceID, id, err)
		c.HTML(http.StatusInternalServerError, "base", gin.H{
			"title": "Delete Project - Error",
			"error": "Failed to delete project",
		})
		return
	}
	
	c.Redirect(http.StatusSeeOther, "/projects")
}

// LoginPage renders the login page
func (h *Handler) LoginPage(c *gin.Context) {
	// Check if user is already authenticated
	user, _ := c.Get("user")
	if user != nil {
		// Redirect to home if already logged in
		c.Redirect(http.StatusSeeOther, "/")
		return
	}
	
	// Check for error message from failed login
	errorMsg := c.Query("error")
	email := c.Query("email")
	
	// Render the login page with separate auth template for deterministic rendering
	c.HTML(http.StatusOK, "auth.html", gin.H{
		"title": "Login",
		"error": errorMsg,
		"email": email,
	})
}

// Login processes the login attempt
func (h *Handler) Login(c *gin.Context) {
	traceID, _ := c.Get("traceID")
	fmt.Printf("[%s] Processing login attempt\n", traceID)
	
	// Parse form data
	email := c.PostForm("email")
	password := c.PostForm("password")
	rememberMe := c.PostForm("remember-me") != ""
	
	// Validate inputs
	if strings.TrimSpace(email) == "" || strings.TrimSpace(password) == "" {
		redirectURL := "/auth/login?error=Email and password are required"
		if email != "" {
			redirectURL += "&email=" + email
		}
		c.Redirect(http.StatusSeeOther, redirectURL)
		return
	}
	
	// Attempt authentication
	user, err := h.authService.Authenticate(email, password)
	
	if err != nil {
		fmt.Printf("[%s] Authentication failed for %s: %v\n", traceID, email, err)
		var errorMsg string
		
		switch err {
		case models.ErrUserNotFound:
			errorMsg = "Invalid email or password"
		case models.ErrInvalidCredentials:
			errorMsg = "Invalid email or password"
		default:
			errorMsg = "Authentication failed"
		}
		
		redirectURL := fmt.Sprintf("/auth/login?error=%s", errorMsg)
		if email != "" {
			redirectURL += "&email=" + email
		}
		c.Redirect(http.StatusSeeOther, redirectURL)
		return
	}
	
	// Generate session token
	token, err := h.authService.GenerateSessionToken(user)
	if err != nil {
		fmt.Printf("[%s] Failed to generate session token: %v\n", traceID, err)
		c.Redirect(http.StatusSeeOther, "/auth/login?error=Session creation failed")
		return
	}
	
	// Set session cookie
	secure := false // Set to true in production
	httpOnly := true
	
	maxAge := 3600 // 1 hour
	if rememberMe {
		maxAge = 7 * 24 * 3600 // 7 days
	}
	
	c.SetCookie("session", token, maxAge, "/", "", secure, httpOnly)
	
	fmt.Printf("[%s] User %s (%s) authenticated successfully\n", traceID, user.Name, user.Email)
	
	// Redirect to homepage or intended destination
	c.Redirect(http.StatusSeeOther, "/")
}

// Logout handles user logout
func (h *Handler) Logout(c *gin.Context) {
	traceID, _ := c.Get("traceID")
	
	// Get session token from cookie
	token, err := c.Cookie("session")
	if err == nil && token != "" {
		// Invalidate the session in auth service
		err = h.authService.InvalidateSession(token)
		if err != nil {
			fmt.Printf("[%s] Error invalidating session: %v\n", traceID, err)
		}
	}
	
	// Clear the session cookie
	c.SetCookie("session", "", -1, "/", "", false, true)
	
	fmt.Printf("[%s] User logged out\n", traceID)
	
	// Redirect to login page
	c.Redirect(http.StatusSeeOther, "/auth/login")
}

// Helper function to get current user from context
func (h *Handler) getCurrentUser(c *gin.Context) *models.User {
	userValue, exists := c.Get("user")
	if !exists {
		// This should never happen with RequireAuth middleware, but handling it gracefully
		traceID, _ := c.Get("traceID")
		fmt.Printf("[%s] WARNING: getCurrentUser called but no user in context\n", traceID)
		c.AbortWithStatus(http.StatusUnauthorized)
		return nil
	}
	
	return userValue.(*models.User)
}

// Feature-specific handlers

// ArchitectureCanvas renders the architecture canvas page
func (h *Handler) ArchitectureCanvas(c *gin.Context) {
	id := c.Param("id")
	traceID, _ := c.Get("traceID")
	user := h.getCurrentUser(c)
	
	project, err := h.projectService.GetProject(id)
	if err != nil {
		fmt.Printf("[%s] Error getting project %s for architecture canvas: %v\n", traceID, id, err)
		c.HTML(http.StatusNotFound, "base", gin.H{
			"title": "Project Not Found",
			"error": "The requested project could not be found",
		})
		return
	}
	
	// Verify ownership
	if project.UserID != user.ID {
		c.HTML(http.StatusForbidden, "base", gin.H{
			"title": "Unauthorized",
			"error": "You do not have permission to view this project",
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
	traceID, _ := c.Get("traceID")
	user := h.getCurrentUser(c)
	
	project, err := h.projectService.GetProject(id)
	if err != nil {
		fmt.Printf("[%s] Error getting project %s for story flow: %v\n", traceID, id, err)
		c.HTML(http.StatusNotFound, "base", gin.H{
			"title": "Project Not Found",
			"error": "The requested project could not be found",
		})
		return
	}
	
	// Verify ownership
	if project.UserID != user.ID {
		c.HTML(http.StatusForbidden, "base", gin.H{
			"title": "Unauthorized",
			"error": "You do not have permission to view this project",
		})
		return
	}
	
	c.HTML(http.StatusOK, "base", gin.H{
		"title":       "Story Flow",
		"projectID":   project.ID,
		"projectName": project.Name,
	})
}

// TaskHub renders the task monitoring page
func (h *Handler) TaskHub(c *gin.Context) {
	id := c.Param("id")
	traceID, _ := c.Get("traceID")
	user := h.getCurrentUser(c)
	
	project, err := h.projectService.GetProject(id)
	if err != nil {
		fmt.Printf("[%s] Error getting project %s for task hub: %v\n", traceID, id, err)
		c.HTML(http.StatusNotFound, "base", gin.H{
			"title": "Project Not Found",
			"error": "The requested project could not be found",
		})
		return
	}
	
	// Verify ownership
	if project.UserID != user.ID {
		c.HTML(http.StatusForbidden, "base", gin.H{
			"title": "Unauthorized",
			"error": "You do not have permission to view this project",
		})
		return
	}
	
	c.HTML(http.StatusOK, "base", gin.H{
		"title":       "Task Hub",
		"projectID":   project.ID,
		"projectName": project.Name,
	})
}

// ReviewQueue renders the review queue page
func (h *Handler) ReviewQueue(c *gin.Context) {
	id := c.Param("id")
	traceID, _ := c.Get("traceID")
	user := h.getCurrentUser(c)
	
	project, err := h.projectService.GetProject(id)
	if err != nil {
		fmt.Printf("[%s] Error getting project %s for review queue: %v\n", traceID, id, err)
		c.HTML(http.StatusNotFound, "base", gin.H{
			"title": "Project Not Found",
			"error": "The requested project could not be found",
		})
		return
	}
	
	// Verify ownership
	if project.UserID != user.ID {
		c.HTML(http.StatusForbidden, "base", gin.H{
			"title": "Unauthorized",
			"error": "You do not have permission to view this project",
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
	traceID, _ := c.Get("traceID")
	user := h.getCurrentUser(c)
	
	project, err := h.projectService.GetProject(id)
	if err != nil {
		fmt.Printf("[%s] Error getting project %s for design assistant: %v\n", traceID, id, err)
		c.HTML(http.StatusNotFound, "base", gin.H{
			"title": "Project Not Found",
			"error": "The requested project could not be found",
		})
		return
	}
	
	// Verify ownership
	if project.UserID != user.ID {
		c.HTML(http.StatusForbidden, "base", gin.H{
			"title": "Unauthorized",
			"error": "You do not have permission to view this project",
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
	user := h.getCurrentUser(c)
	traceID, _ := c.Get("traceID")
	
	projects, err := h.projectService.ListProjects(user.ID)
	if err != nil {
		fmt.Printf("[%s] Error listing projects for API: %v\n", traceID, err)
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
	user := h.getCurrentUser(c)
	traceID, _ := c.Get("traceID")
	
	project, err := h.projectService.GetProject(id)
	if err != nil {
		fmt.Printf("[%s] Error getting project %s for API: %v\n", traceID, id, err)
		c.JSON(http.StatusNotFound, gin.H{
			"status": "error",
			"error":  "Project not found",
		})
		return
	}
	
	// Verify ownership
	if project.UserID != user.ID {
		fmt.Printf("[%s] Unauthorized API access attempt for project %s by user %s\n", traceID, id, user.ID)
		c.JSON(http.StatusForbidden, gin.H{
			"status": "error",
			"error":  "You do not have permission to access this project",
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
	user := h.getCurrentUser(c)
	traceID, _ := c.Get("traceID")
	
	var projectData struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}
	
	if err := c.BindJSON(&projectData); err != nil {
		fmt.Printf("[%s] Invalid project data format: %v\n", traceID, err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  "Invalid data format",
		})
		return
	}
	
	project := models.NewProject(projectData.Name, projectData.Description, user.ID)
	
	err := h.projectService.CreateProject(project)
	if err != nil {
		fmt.Printf("[%s] Error creating project via API: %v\n", traceID, err)
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
	user := h.getCurrentUser(c)
	traceID, _ := c.Get("traceID")
	
	// Get existing project
	project, err := h.projectService.GetProject(id)
	if err != nil {
		fmt.Printf("[%s] Error getting project %s for API update: %v\n", traceID, id, err)
		c.JSON(http.StatusNotFound, gin.H{
			"status": "error",
			"error":  "Project not found",
		})
		return
	}
	
	// Verify ownership
	if project.UserID != user.ID {
		fmt.Printf("[%s] Unauthorized API update attempt for project %s by user %s\n", traceID, id, user.ID)
		c.JSON(http.StatusForbidden, gin.H{
			"status": "error",
			"error":  "You do not have permission to update this project",
		})
		return
	}
	
	// Parse update data
	var updateData struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	
	if err := c.BindJSON(&updateData); err != nil {
		fmt.Printf("[%s] Invalid project update data: %v\n", traceID, err)
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
		fmt.Printf("[%s] Error updating project %s via API: %v\n", traceID, id, err)
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
	user := h.getCurrentUser(c)
	traceID, _ := c.Get("traceID")
	
	// Get project to verify ownership
	project, err := h.projectService.GetProject(id)
	if err != nil {
		fmt.Printf("[%s] Error getting project %s for API delete: %v\n", traceID, id, err)
		c.JSON(http.StatusNotFound, gin.H{
			"status": "error",
			"error":  "Project not found",
		})
		return
	}
	
	// Verify ownership
	if project.UserID != user.ID {
		fmt.Printf("[%s] Unauthorized API delete attempt for project %s by user %s\n", traceID, id, user.ID)
		c.JSON(http.StatusForbidden, gin.H{
			"status": "error",
			"error":  "You do not have permission to delete this project",
		})
		return
	}
	
	err = h.projectService.DeleteProject(id)
	if err != nil {
		fmt.Printf("[%s] Error deleting project %s via API: %v\n", traceID, id, err)
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
