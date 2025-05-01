package services

import (
	"testing"
	"time"

	"github.com/jamesjscully/un-tie_code/src/api/models"
	"github.com/jamesjscully/un-tie_code/src/api/repositories"
)

// Test_ProjectService_CreateProject tests project creation
func Test_ProjectService_CreateProject(t *testing.T) {
	// Arrange
	repo := repositories.NewMemoryProjectRepository()
	service := NewProjectService(repo)
	
	// Test data
	testUserID := "user123"
	testName := "Test Project"
	testDesc := "Test project description"
	
	project := &models.Project{
		Name:        testName,
		Description: testDesc,
		UserID:      testUserID,
	}
	
	// Act
	err := service.CreateProject(project)
	
	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if project.ID == "" {
		t.Fatal("Expected project ID to be set, got empty string")
	}
	
	if !project.CreatedAt.Before(time.Now()) || project.CreatedAt.IsZero() {
		t.Fatal("Expected CreatedAt to be set to a valid time")
	}
	
	if !project.UpdatedAt.Before(time.Now()) || project.UpdatedAt.IsZero() {
		t.Fatal("Expected UpdatedAt to be set to a valid time")
	}
	
	// Verify project is stored
	projects, err := service.ListProjects(testUserID)
	if err != nil {
		t.Fatalf("Expected no error listing projects, got %v", err)
	}
	
	if len(projects) != 1 {
		t.Fatalf("Expected 1 project, got %d", len(projects))
	}
	
	if projects[0].Name != testName {
		t.Fatalf("Expected project name %s, got %s", testName, projects[0].Name)
	}
}

// Test_ProjectService_GetProject tests retrieving a project
func Test_ProjectService_GetProject(t *testing.T) {
	// Arrange
	repo := repositories.NewMemoryProjectRepository()
	service := NewProjectService(repo)
	
	// Create test project
	testUserID := "user123"
	testName := "Test Project"
	project := &models.Project{
		ID:          "proj-123",
		Name:        testName,
		Description: "Test Description",
		UserID:      testUserID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	// Use the service to create the project (instead of directly accessing the repo)
	err := service.CreateProject(project)
	if err != nil {
		t.Fatalf("Failed to set up test: %v", err)
	}
	
	// Act
	retrievedProject, err := service.GetProject("proj-123")
	
	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if retrievedProject == nil {
		t.Fatal("Expected to retrieve project, got nil")
	}
	
	if retrievedProject.ID != "proj-123" {
		t.Fatalf("Expected project ID proj-123, got %s", retrievedProject.ID)
	}
	
	if retrievedProject.Name != testName {
		t.Fatalf("Expected project name %s, got %s", testName, retrievedProject.Name)
	}
}

// Test_ProjectService_UpdateProject tests updating a project
func Test_ProjectService_UpdateProject(t *testing.T) {
	// Arrange
	repo := repositories.NewMemoryProjectRepository()
	service := NewProjectService(repo)
	
	// Create test project
	testUserID := "user123"
	originalName := "Original Name"
	project := &models.Project{
		ID:          "proj-123",
		Name:        originalName,
		Description: "Original Description",
		UserID:      testUserID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	// Use the service to create the project (instead of directly accessing the repo)
	err := service.CreateProject(project)
	if err != nil {
		t.Fatalf("Failed to set up test: %v", err)
	}
	
	// Act
	updatedName := "Updated Name"
	updatedProject := &models.Project{
		ID:          "proj-123",
		Name:        updatedName,
		Description: "Updated Description",
		UserID:      testUserID,
	}
	
	err = service.UpdateProject(updatedProject)
	
	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	// Verify project was updated
	retrievedProject, err := service.GetProject("proj-123")
	if err != nil {
		t.Fatalf("Failed to retrieve project: %v", err)
	}
	
	if retrievedProject.Name != updatedName {
		t.Fatalf("Expected project name to be %s, got %s", updatedName, retrievedProject.Name)
	}
	
	if retrievedProject.CreatedAt.IsZero() {
		t.Fatal("Expected CreatedAt to be preserved")
	}
	
	if retrievedProject.UpdatedAt.Before(project.UpdatedAt) {
		t.Fatal("Expected UpdatedAt to be more recent")
	}
}

// Test_ProjectService_DeleteProject tests project deletion
func Test_ProjectService_DeleteProject(t *testing.T) {
	// Arrange
	repo := repositories.NewMemoryProjectRepository()
	service := NewProjectService(repo)
	
	// Create test project
	testUserID := "user123"
	project := &models.Project{
		ID:          "proj-123",
		Name:        "Project to Delete",
		Description: "Test Description",
		UserID:      testUserID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	// Use the service to create the project (instead of directly accessing the repo)
	err := service.CreateProject(project)
	if err != nil {
		t.Fatalf("Failed to set up test: %v", err)
	}
	
	// Act
	err = service.DeleteProject("proj-123")
	
	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	// Verify project was deleted
	projects, err := service.ListProjects(testUserID)
	if err != nil {
		t.Fatalf("Failed to list projects: %v", err)
	}
	
	if len(projects) != 0 {
		t.Fatalf("Expected 0 projects after deletion, got %d", len(projects))
	}
	
	_, err = service.GetProject("proj-123")
	if err == nil {
		t.Fatal("Expected error retrieving deleted project, got nil")
	}
}
