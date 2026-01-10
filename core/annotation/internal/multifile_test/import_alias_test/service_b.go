package main

import (
	models "github.com/primadi/lokstra/core/annotation/internal/multifile_test/import_alias_test/pkgb"
)

// @EndpointService name="service-b", prefix="/api/b"
type ServiceB struct {
}

// @Route "GET /users"
// This service uses pkgb with alias "models" (CONFLICT: same alias, different path)
func (s *ServiceB) GetUsers() (*models.User, error) {
	return &models.User{
		UserID:   "b1",
		FullName: "User from B",
		Email:    "user@b.com",
	}, nil
}

// @Route "POST /respond"
func (s *ServiceB) Respond() (*models.Response, error) {
	return &models.Response{
		Status:  "ok",
		Message: "Response from B",
	}, nil
}
