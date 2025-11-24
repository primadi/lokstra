package main

import "github.com/primadi/lokstra/core/service"

// @RouterService name="profile-service", prefix="/api/profiles"
type ProfileService struct {
	// @Inject "profile-repo"
	ProfileRepo *service.Cached[any]
}

// @Route "GET /profiles/{id}"
func (s *ProfileService) GetByID(id string) (string, error) {
	return "profile", nil
}

// @Route "PUT /profiles/{id}"
func (s *ProfileService) Update(id string, name string) error {
	return nil
}
