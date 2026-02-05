package main

import (
	userentity "github.com/primadi/lokstra/core/annotation/internal/multifile_test/import_alias_test/pkga"
)

// @Handler name="service-c", prefix="/api/c"
type ServiceC struct {
}

// @Route "GET /entity"
// This service uses pkga with alias "userentity"
func (s *ServiceC) GetEntity() (*userentity.User, error) {
	return &userentity.User{
		ID:   "c1",
		Name: "User from C",
	}, nil
}

// @Route "POST /request"
func (s *ServiceC) HandleRequest(req *userentity.Request) error {
	return nil
}
