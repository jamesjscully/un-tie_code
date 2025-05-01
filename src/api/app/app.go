package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jamesjscully/un-tie_code/src/api/config"
	"github.com/jamesjscully/un-tie_code/src/api/handlers"
	"github.com/jamesjscully/un-tie_code/src/api/middleware"
	"github.com/jamesjscully/un-tie_code/src/api/models"
	"github.com/jamesjscully/un-tie_code/src/api/repositories"
	"github.com/jamesjscully/un-tie_code/src/api/services"
)

// Application represents the main application with all its dependencies
type Application struct {
	Config         *config.Config
	Router         *gin.Engine
	ProjectService models.ProjectService
	AuthService    models.AuthService
	Server         *http.Server
}

// NewApplication creates and initializes a new application instance
func NewApplication() *Application {
	// Load configuration
	cfg := config.LoadFromEnv()
	
	// Setup repository layer based on config
	var projectRepo models.ProjectRepository
	var userRepo models.UserRepository
	
	// For now, always use memory repositories
	// This will be extended to support database repositories based on config
	projectRepo = repositories.NewMemoryProjectRepository()
	userRepo = repositories.NewMemoryUserRepository()
	
	// Create services with repositories (dependency injection)
	projectService := services.NewProjectService(projectRepo)
	authService := services.NewAuthService(userRepo)
	
	// Create Gin router with appropriate mode
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.Default()
	
	// Set up the HTTP server
	server := &http.Server{
		Addr:         cfg.GetAddress(),
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}
	
	// Create the application
	app := &Application{
		Config:         cfg,
		Router:         router,
		ProjectService: projectService,
		AuthService:    authService,
		Server:         server,
	}
	
	// Initialize routes and middleware
	app.setupMiddleware()
	app.setupRoutes()
	
	return app
}

// setupMiddleware configures middleware for the application
func (a *Application) setupMiddleware() {
	// Add global middleware
	a.Router.Use(middleware.ErrorHandler())
	
	// Add request tracing for observability
	a.Router.Use(func(c *gin.Context) {
		traceID := fmt.Sprintf("request-%d", time.Now().UnixNano())
		c.Set("traceID", traceID)
		c.Next()
	})
	
	// Add session middleware for authentication
	a.Router.Use(middleware.SessionMiddleware(a.AuthService))
	
	// Set up static file serving
	a.Router.Static("/static", "./src/web/static")
	a.Router.LoadHTMLGlob("./src/web/templates/*")
}

// setupRoutes configures all routes for the application
func (a *Application) setupRoutes() {
	// Create handler with injected services
	h := handlers.NewHandler(a.ProjectService, a.AuthService)
	
	// Health check
	a.Router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	
	// Auth routes (public)
	auth := a.Router.Group("/auth")
	{
		auth.GET("/login", h.LoginPage)
		auth.POST("/login", h.Login)
		auth.GET("/logout", h.Logout)
	}
	
	// Protected routes requiring authentication
	authenticated := a.Router.Group("/")
	authenticated.Use(middleware.RequireAuth())
	{
		// Home route - protected
		authenticated.GET("/", h.HomeHandler)
		
		// Project routes
		projects := authenticated.Group("/projects")
		{
			projects.GET("/", h.ListProjects)
			projects.GET("/new", h.NewProjectForm)
			projects.POST("/", h.CreateProject)
			projects.GET("/:id", h.GetProject)
			projects.PUT("/:id", h.UpdateProject)
			projects.DELETE("/:id", h.DeleteProject)
			
			// Feature-specific project routes
			if a.Config.IsFeatureEnabled("FEATURE_ARCHITECTURE_CANVAS") {
				projects.GET("/:id/architecture", h.ArchitectureCanvas)
			}
			
			if a.Config.IsFeatureEnabled("FEATURE_STORY_FLOW") {
				projects.GET("/:id/stories", h.StoryFlow)
			}
			
			if a.Config.IsFeatureEnabled("FEATURE_TASK_HUB") {
				projects.GET("/:id/tasks", h.TaskHub)
			}
			
			if a.Config.IsFeatureEnabled("FEATURE_REVIEW_QUEUE") {
				projects.GET("/:id/review", h.ReviewQueue)
			}
			
			if a.Config.IsFeatureEnabled("FEATURE_DESIGN_ASSISTANT") {
				projects.GET("/:id/assistant", h.DesignAssistant)
			}
		}
	}
	
	// API routes
	api := a.Router.Group("/api/v1")
	{
		// Public API endpoints
		api.GET("/status", h.APIStatus)
		
		// Protected API routes
		apiAuth := api.Group("/")
		apiAuth.Use(middleware.RequireAuth())
		{
			apiProjects := apiAuth.Group("/projects")
			{
				apiProjects.GET("/", h.APIListProjects)
				apiProjects.POST("/", h.APICreateProject)
				apiProjects.GET("/:id", h.APIGetProject)
				apiProjects.PUT("/:id", h.APIUpdateProject)
				apiProjects.DELETE("/:id", h.APIDeleteProject)
			}
		}
	}
}

// Start begins the server and handles graceful shutdown
func (a *Application) Start() {
	// Start the server in a goroutine
	go func() {
		log.Printf("Server starting on %s", a.Server.Addr)
		if err := a.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()
	
	// Set up graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	
	// Block until we receive a shutdown signal
	<-quit
	log.Println("Server shutting down...")
	
	// Create a context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	// Attempt graceful shutdown
	if err := a.Server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	
	log.Println("Server exited properly")
}
