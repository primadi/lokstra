package main

import (
	"errors"
	"time"
)

// Todo represents a todo item
type Todo struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Completed   bool      `json:"completed"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateTodoParams for creating new todos
type CreateTodoParams struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func (p *CreateTodoParams) Validate() error {
	if p.Title == "" {
		return errors.New("title is required")
	}
	if len(p.Title) < 3 {
		return errors.New("title must be at least 3 characters")
	}
	return nil
}

// UpdateTodoParams for updating todos
type UpdateTodoParams struct {
	ID          int     `path:"id"`
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Completed   *bool   `json:"completed,omitempty"`
}

func (p *UpdateTodoParams) Validate() error {
	if p.Title != nil && len(*p.Title) < 3 {
		return errors.New("title must be at least 3 characters")
	}
	return nil
}

// GetTodoParams for getting todo by ID
type GetTodoParams struct {
	ID int `path:"id"`
}

// DeleteTodoParams for deleting todo
type DeleteTodoParams struct {
	ID int `path:"id"`
}
