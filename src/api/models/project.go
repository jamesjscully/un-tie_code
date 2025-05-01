package models

import (
	"time"
)

// Project represents a software project in the Un-tie.me system
type Project struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	UserID      string    `json:"userId"`
	TechStack   TechStack `json:"techStack"`
	Features    []Feature `json:"features"`
}

// TechStack represents the technology choices for a project
type TechStack struct {
	Frontend  []string `json:"frontend"`
	Backend   []string `json:"backend"`
	Database  []string `json:"database"`
	Hosting   []string `json:"hosting"`
	CI        []string `json:"ci"`
	Other     []string `json:"other"`
}

// Feature represents a project feature with versioning
type Feature struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Versions    []FeatureImpl `json:"versions"` // Different implementations across versions
}

// FeatureImpl represents a specific implementation of a feature for a version
type FeatureImpl struct {
	Version     string `json:"version"` // e.g., "MVP", "v1.0", "v2.0"
	Status      string `json:"status"`  // "planned", "in-progress", "completed", "deferred"
	Description string `json:"description"`
}

// ProjectRepository defines the interface for project persistence
// Following interface segregation principle from SOLID
type ProjectRepository interface {
	GetByID(id string) (*Project, error)
	List(userID string) ([]*Project, error)
	Create(project *Project) error
	Update(project *Project) error
	Delete(id string) error
}

// ProjectService defines the interface for project business logic
// Keeping business logic separate from persistence (single responsibility)
type ProjectService interface {
	GetProject(id string) (*Project, error)
	ListProjects(userID string) ([]*Project, error)
	CreateProject(project *Project) error
	UpdateProject(project *Project) error
	DeleteProject(id string) error
	GeneratePRD(project *Project) (string, error)
}

// NewProject creates a new project with proper initialization
// Factory function pattern for consistent object creation
func NewProject(name, description, userID string) *Project {
	now := time.Now()
	return &Project{
		ID:          generateID(), // Implement this function elsewhere
		Name:        name,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
		UserID:      userID,
		TechStack:   TechStack{},
		Features:    []Feature{},
	}
}

// Simple ID generator - would be replaced with UUID or other robust solution
func generateID() string {
	return time.Now().Format("20060102150405")
}
