package main

import (
	pkgamodel "github.com/primadi/lokstra/core/annotation/internal/multifile_test/import_alias_test/pkga"
)

// @RouterService name="service-d", prefix="/api/d"
type ServiceD struct {
}

// @Route "GET /data"
// This service uses pkga with alias "pkgamodel" (same path as service-c, different alias)
func (s *ServiceD) GetData() (*pkgamodel.User, error) {
	return &pkgamodel.User{
		ID:   "d1",
		Name: "User from D",
	}, nil
}

// @Route "POST /action"
func (s *ServiceD) PerformAction(req *pkgamodel.Request) error {
	return nil
}
