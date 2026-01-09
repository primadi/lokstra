package main

import (
	models "github.com/primadi/lokstra/core/annotation/internal/multifile_test/import_alias_test/pkga"
)

// @RouterService name="service-a", prefix="/api/a"
type ServiceA struct {
}

// @Route "GET /users"
// This service uses pkga with alias "models"
func (s *ServiceA) GetUsers() (*models.User, error) {
	return &models.User{
		ID:   "a1",
		Name: "User from A",
	}, nil
}

// @Route "POST /process"
func (s *ServiceA) Process(req *models.Request) error {
	return nil
}
