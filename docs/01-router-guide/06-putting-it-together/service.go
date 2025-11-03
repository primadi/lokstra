package main

import (
	"errors"
	"sync"
	"time"
)

// TodoService manages todo operations
type TodoService struct {
	todos  map[int]*Todo
	mu     sync.RWMutex
	nextID int
}

// NewTodoService creates a new todo service
func NewTodoService() *TodoService {
	return &TodoService{
		todos:  make(map[int]*Todo),
		nextID: 1,
	}
}

// Create a new todo
func (s *TodoService) Create(params *CreateTodoParams) (*Todo, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	todo := &Todo{
		ID:          s.nextID,
		Title:       params.Title,
		Description: params.Description,
		Completed:   false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	s.todos[s.nextID] = todo
	s.nextID++

	return todo, nil
}

// List returns all todos
func (s *TodoService) List() ([]*Todo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	todos := make([]*Todo, 0, len(s.todos))
	for _, todo := range s.todos {
		todos = append(todos, todo)
	}

	return todos, nil
}

// GetByID returns a todo by ID
func (s *TodoService) GetByID(params *GetTodoParams) (*Todo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	todo, exists := s.todos[params.ID]
	if !exists {
		return nil, errors.New("todo not found")
	}

	return todo, nil
}

// Update a todo
func (s *TodoService) Update(params *UpdateTodoParams) (*Todo, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	todo, exists := s.todos[params.ID]
	if !exists {
		return nil, errors.New("todo not found")
	}

	// Update fields if provided
	if params.Title != nil {
		todo.Title = *params.Title
	}
	if params.Description != nil {
		todo.Description = *params.Description
	}
	if params.Completed != nil {
		todo.Completed = *params.Completed
	}

	todo.UpdatedAt = time.Now()

	return todo, nil
}

// Delete a todo
func (s *TodoService) Delete(params *DeleteTodoParams) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.todos[params.ID]; !exists {
		return errors.New("todo not found")
	}

	delete(s.todos, params.ID)
	return nil
}
