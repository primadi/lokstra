package main

import "fmt"

// ============================================================================
// Mock Service Implementations
// ============================================================================

// UserService - Mock implementation for testing
type UserService struct{}

type getUsersParams struct{}

func (s *UserService) GetUsers(c *getUsersParams) (map[string]any, error) {
	return map[string]any{
		"users": []map[string]any{
			{"id": 1, "name": "Alice", "email": "alice@example.com"},
			{"id": 2, "name": "Bob", "email": "bob@example.com"},
			{"id": 3, "name": "Charlie", "email": "charlie@example.com"},
		},
		"source": "App1 (port 9090)",
	}, nil
}

type getUserParams struct {
	ID int `path:"id"`
}

func (s *UserService) GetUser(c *getUserParams) (map[string]any, error) {
	users := map[int]map[string]any{
		1: {"id": 1, "name": "Alice", "email": "alice@example.com"},
		2: {"id": 2, "name": "Bob", "email": "bob@example.com"},
		3: {"id": 3, "name": "Charlie", "email": "charlie@example.com"},
	}

	if user, ok := users[c.ID]; ok {
		user["source"] = "App1 (port 9090)"
		return user, nil
	}

	return nil, fmt.Errorf("user not found")
}

type createUserParams struct {
	Data map[string]any `json:"*"`
}

func (s *UserService) CreateUser(c *createUserParams) (map[string]any, error) {
	return map[string]any{
		"message": "User created successfully",
		"data":    c.Data,
		"source":  "App1 (port 9090)",
	}, nil
}

type updateUserParams struct {
	ID   int            `path:"id"`
	Data map[string]any `json:"*"`
}

func (s *UserService) UpdateUser(c *updateUserParams) (map[string]any, error) {
	return map[string]any{
		"message": "User updated successfully",
		"id":      c.ID,
		"data":    c.Data,
		"source":  "App1 (port 9090)",
	}, nil
}

type deleteUserParams struct {
	ID int `path:"id"`
}

func (s *UserService) DeleteUser(c *deleteUserParams) (map[string]any, error) {
	return map[string]any{
		"message": "User deleted successfully",
		"id":      c.ID,
		"source":  "App1 (port 9090)",
	}, nil
}

// ProductService - Mock implementation for testing
type ProductService struct{}

type getProductsParams struct{}

func (s *ProductService) GetProducts(c *getProductsParams) (map[string]any, error) {
	return map[string]any{
		"products": []map[string]any{
			{"id": 101, "name": "Laptop", "price": 1200, "stock": 50},
			{"id": 102, "name": "Mouse", "price": 25, "stock": 200},
			{"id": 103, "name": "Keyboard", "price": 75, "stock": 150},
		},
		"source": "App2 (port 9091)",
	}, nil
}

type getProductParams struct {
	ID int `path:"id"`
}

func (s *ProductService) GetProduct(c *getProductParams) map[string]any {
	products := map[int]map[string]any{
		101: {"id": 101, "name": "Laptop", "price": 1200, "stock": 50},
		102: {"id": 102, "name": "Mouse", "price": 25, "stock": 200},
		103: {"id": 103, "name": "Keyboard", "price": 75, "stock": 150},
	}

	if product, ok := products[c.ID]; ok {
		product["source"] = "App2 (port 9091)"
		return product
	}

	return map[string]any{"error": "Product not found", "source": "App2 (port 9091)"}
}

type createProductParams struct {
	Data map[string]any `json:"*"`
}

func (s *ProductService) CreateProduct(c *createProductParams) map[string]any {
	return map[string]any{
		"message": "Product created successfully",
		"data":    c.Data,
		"source":  "App2 (port 9091)",
	}
}

type updateProductParams struct {
	ID   int            `path:"id"`
	Data map[string]any `json:"*"`
}

func (s *ProductService) UpdateProduct(c *updateProductParams) map[string]any {
	return map[string]any{
		"message": "Product updated successfully",
		"id":      c.ID,
		"data":    c.Data,
		"source":  "App2 (port 9091)",
	}
}

type deleteProductParams struct {
	ID int `path:"id"`
}

func (s *ProductService) DeleteProduct(c *deleteProductParams) map[string]any {
	return map[string]any{
		"message": "Product deleted successfully",
		"id":      c.ID,
		"source":  "App2 (port 9091)",
	}
}

func UserServiceFactory(_ map[string]any) any {
	return &UserService{}
}

func ProductServiceFactory(_ map[string]any) any {
	return &ProductService{}
}
