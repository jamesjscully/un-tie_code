package repositories

import (
	"errors"
	"fmt"
	"sync"

	"github.com/jamesjscully/un-tie_code/src/api/models"
)

// MemoryProjectRepository implements ProjectRepository using in-memory storage
// This is useful for development and testing
type MemoryProjectRepository struct {
	mutex    sync.RWMutex
	projects map[string]*models.Project
}

// NewMemoryProjectRepository creates a new in-memory project repository
func NewMemoryProjectRepository() models.ProjectRepository {
	return &MemoryProjectRepository{
		projects: make(map[string]*models.Project),
	}
}

// GetByID retrieves a project by ID from memory
func (r *MemoryProjectRepository) GetByID(id string) (*models.Project, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	project, exists := r.projects[id]
	if !exists {
		return nil, errors.New("project not found")
	}
	
	// Return a copy to prevent external modification
	projectCopy := *project
	return &projectCopy, nil
}

// List retrieves all projects for a user from memory
func (r *MemoryProjectRepository) List(userID string) ([]*models.Project, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	var userProjects []*models.Project
	for _, project := range r.projects {
		if project.UserID == userID {
			// Create a copy of each project to prevent external modification
			projectCopy := *project
			userProjects = append(userProjects, &projectCopy)
		}
	}
	
	return userProjects, nil
}

// Create adds a new project to memory
func (r *MemoryProjectRepository) Create(project *models.Project) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	if _, exists := r.projects[project.ID]; exists {
		return errors.New("project with this ID already exists")
	}
	
	// Store a copy to prevent external modification
	projectCopy := *project
	r.projects[project.ID] = &projectCopy
	
	return nil
}

// Update modifies an existing project in memory
func (r *MemoryProjectRepository) Update(project *models.Project) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	if _, exists := r.projects[project.ID]; !exists {
		return errors.New("project not found")
	}
	
	// Store a copy to prevent external modification
	projectCopy := *project
	r.projects[project.ID] = &projectCopy
	
	return nil
}

// Delete removes a project from memory
func (r *MemoryProjectRepository) Delete(id string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	if _, exists := r.projects[id]; !exists {
		return errors.New("project not found")
	}
	
	delete(r.projects, id)
	
	return nil
}

// MemoryUserRepository implements UserRepository using in-memory storage
type MemoryUserRepository struct {
	mutex sync.RWMutex
	users map[string]*models.User  // by ID
	byEmail map[string]*models.User // by email
}

// NewMemoryUserRepository creates a new in-memory user repository
func NewMemoryUserRepository() models.UserRepository {
	repo := &MemoryUserRepository{
		users:   make(map[string]*models.User),
		byEmail: make(map[string]*models.User),
	}
	
	// Create a test user for development/login functionality
	testUser := models.NewUser("test@untie.me", "Test User")
	testUser.Role = "admin"
	
	// Save the test user
	err := repo.Create(testUser)
	if err != nil {
		// Log error but continue (non-critical)
		fmt.Printf("Failed to create test user: %v\n", err)
	} else {
		fmt.Println("Test user created with email: test@untie.me")
	}
	
	return repo
}

// GetByID retrieves a user by ID from memory
func (r *MemoryUserRepository) GetByID(id string) (*models.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	user, exists := r.users[id]
	if !exists {
		return nil, models.ErrUserNotFound
	}
	
	// Return a copy to prevent external modification
	userCopy := *user
	return &userCopy, nil
}

// GetByEmail retrieves a user by email from memory
func (r *MemoryUserRepository) GetByEmail(email string) (*models.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	user, exists := r.byEmail[email]
	if !exists {
		return nil, models.ErrUserNotFound
	}
	
	// Return a copy to prevent external modification
	userCopy := *user
	return &userCopy, nil
}

// Create adds a new user to memory
func (r *MemoryUserRepository) Create(user *models.User) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	if _, exists := r.users[user.ID]; exists {
		return errors.New("user with this ID already exists")
	}
	
	if _, exists := r.byEmail[user.Email]; exists {
		return models.ErrEmailAlreadyExists
	}
	
	// Store a copy to prevent external modification
	userCopy := *user
	r.users[user.ID] = &userCopy
	r.byEmail[user.Email] = &userCopy
	
	return nil
}

// Update modifies an existing user in memory
func (r *MemoryUserRepository) Update(user *models.User) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	existingUser, exists := r.users[user.ID]
	if !exists {
		return models.ErrUserNotFound
	}
	
	// Check if email changed and if new email is already in use
	if existingUser.Email != user.Email {
		if _, emailExists := r.byEmail[user.Email]; emailExists {
			return models.ErrEmailAlreadyExists
		}
		// Remove old email reference
		delete(r.byEmail, existingUser.Email)
	}
	
	// Store a copy to prevent external modification
	userCopy := *user
	r.users[user.ID] = &userCopy
	r.byEmail[user.Email] = &userCopy
	
	return nil
}

// Delete removes a user from memory
func (r *MemoryUserRepository) Delete(id string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	user, exists := r.users[id]
	if !exists {
		return models.ErrUserNotFound
	}
	
	delete(r.users, id)
	delete(r.byEmail, user.Email)
	
	return nil
}
