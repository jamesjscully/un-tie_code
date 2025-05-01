package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/jamesjscully/un-tie_code/src/api/models"
)

// ProjectServiceImpl implements the ProjectService interface
// Following the Dependency Inversion principle with repository injection
type ProjectServiceImpl struct {
	repo models.ProjectRepository
	// Trace ID generator for traceability
	traceIDGenerator func() string
}

// NewProjectService creates a new instance of the project service
func NewProjectService(repo models.ProjectRepository) models.ProjectService {
	return &ProjectServiceImpl{
		repo: repo,
		traceIDGenerator: func() string {
			return fmt.Sprintf("trace-%d", time.Now().UnixNano())
		},
	}
}

// GetProject retrieves a project by ID
func (s *ProjectServiceImpl) GetProject(id string) (*models.Project, error) {
	traceID := s.traceIDGenerator()
	
	// Log the operation for traceability
	fmt.Printf("[%s] Getting project with ID: %s\n", traceID, id)
	
	project, err := s.repo.GetByID(id)
	if err != nil {
		// Log failure with trace ID for debugging
		fmt.Printf("[%s] Failed to get project: %v\n", traceID, err)
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	
	// Log success
	fmt.Printf("[%s] Successfully retrieved project: %s\n", traceID, project.Name)
	return project, nil
}

// ListProjects retrieves all projects for a user
func (s *ProjectServiceImpl) ListProjects(userID string) ([]*models.Project, error) {
	traceID := s.traceIDGenerator()
	
	// Log the operation
	fmt.Printf("[%s] Listing projects for user: %s\n", traceID, userID)
	
	projects, err := s.repo.List(userID)
	if err != nil {
		fmt.Printf("[%s] Failed to list projects: %v\n", traceID, err)
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}
	
	fmt.Printf("[%s] Successfully listed %d projects\n", traceID, len(projects))
	return projects, nil
}

// CreateProject handles project creation with validation
func (s *ProjectServiceImpl) CreateProject(project *models.Project) error {
	traceID := s.traceIDGenerator()
	
	// Log the operation
	fmt.Printf("[%s] Creating new project: %s\n", traceID, project.Name)
	
	// Validate project data
	if project.Name == "" {
		return errors.New("project name cannot be empty")
	}
	
	// Set timestamps
	now := time.Now()
	project.CreatedAt = now
	project.UpdatedAt = now
	
	// Generate ID if empty
	if project.ID == "" {
		project.ID = fmt.Sprintf("proj-%d", now.UnixNano())
	}
	
	err := s.repo.Create(project)
	if err != nil {
		fmt.Printf("[%s] Failed to create project: %v\n", traceID, err)
		return fmt.Errorf("failed to create project: %w", err)
	}
	
	fmt.Printf("[%s] Successfully created project with ID: %s\n", traceID, project.ID)
	return nil
}

// UpdateProject handles project updates with validation
func (s *ProjectServiceImpl) UpdateProject(project *models.Project) error {
	traceID := s.traceIDGenerator()
	
	// Log the operation
	fmt.Printf("[%s] Updating project: %s\n", traceID, project.ID)
	
	// Validate project exists
	existingProject, err := s.repo.GetByID(project.ID)
	if err != nil {
		fmt.Printf("[%s] Failed to get existing project: %v\n", traceID, err)
		return fmt.Errorf("failed to get existing project: %w", err)
	}
	
	// Validate project data
	if project.Name == "" {
		return errors.New("project name cannot be empty")
	}
	
	// Preserve creation time
	project.CreatedAt = existingProject.CreatedAt
	
	// Update timestamp
	project.UpdatedAt = time.Now()
	
	err = s.repo.Update(project)
	if err != nil {
		fmt.Printf("[%s] Failed to update project: %v\n", traceID, err)
		return fmt.Errorf("failed to update project: %w", err)
	}
	
	fmt.Printf("[%s] Successfully updated project: %s\n", traceID, project.ID)
	return nil
}

// DeleteProject handles project deletion
func (s *ProjectServiceImpl) DeleteProject(id string) error {
	traceID := s.traceIDGenerator()
	
	// Log the operation
	fmt.Printf("[%s] Deleting project: %s\n", traceID, id)
	
	// Verify project exists
	_, err := s.repo.GetByID(id)
	if err != nil {
		fmt.Printf("[%s] Failed to get project for deletion: %v\n", traceID, err)
		return fmt.Errorf("failed to get project for deletion: %w", err)
	}
	
	err = s.repo.Delete(id)
	if err != nil {
		fmt.Printf("[%s] Failed to delete project: %v\n", traceID, err)
		return fmt.Errorf("failed to delete project: %w", err)
	}
	
	fmt.Printf("[%s] Successfully deleted project: %s\n", traceID, id)
	return nil
}

// GeneratePRD generates a Product Requirements Document for the project
func (s *ProjectServiceImpl) GeneratePRD(project *models.Project) (string, error) {
	traceID := s.traceIDGenerator()
	
	// Log the operation
	fmt.Printf("[%s] Generating PRD for project: %s\n", traceID, project.ID)
	
	// This would normally integrate with an AI service to generate the PRD
	// For now, we'll just create a simple markdown document
	
	prd := fmt.Sprintf(`# Product Requirements Document — **"%s"**

## 1  Overview  
%s

## 2  Core Features  
`, project.Name, project.Description)
	
	// Add features to PRD
	for i, feature := range project.Features {
		prd += fmt.Sprintf("### 2.%d %s\n%s\n\n", i+1, feature.Name, feature.Description)
	}
	
	// Add tech stack
	prd += "## 3  Technical Stack\n"
	if len(project.TechStack.Frontend) > 0 {
		prd += "### Frontend\n"
		for _, tech := range project.TechStack.Frontend {
			prd += fmt.Sprintf("- %s\n", tech)
		}
	}
	
	if len(project.TechStack.Backend) > 0 {
		prd += "### Backend\n"
		for _, tech := range project.TechStack.Backend {
			prd += fmt.Sprintf("- %s\n", tech)
		}
	}
	
	fmt.Printf("[%s] Successfully generated PRD for project: %s\n", traceID, project.ID)
	return prd, nil
}
