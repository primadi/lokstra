package main

import (
	"fmt"
	"time"

	"github.com/primadi/lokstra"
)

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Product struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Category string  `json:"category"`
}

var users = []User{
	{ID: 1, Name: "Alice", Email: "alice@example.com"},
	{ID: 2, Name: "Bob", Email: "bob@example.com"},
	{ID: 3, Name: "Charlie", Email: "charlie@example.com"},
}

var products = []Product{
	{ID: 1, Name: "Laptop", Price: 999.99, Category: "electronics"},
	{ID: 2, Name: "Mouse", Price: 29.99, Category: "electronics"},
	{ID: 3, Name: "Desk", Price: 299.99, Category: "furniture"},
	{ID: 4, Name: "Chair", Price: 199.99, Category: "furniture"},
}

func main() {
	r := lokstra.NewRouter("api")

	// Path parameter - Get single user by ID
	r.GET("/users/{id}", getUser)

	// Path parameter - Update user
	r.PUT("/users/{id}", updateUser)

	// Path parameter - Delete user
	r.DELETE("/users/{id}", deleteUser)

	// Query parameters - Search products
	r.GET("/products", searchProducts)

	// Combined - User's products (path + query)
	r.GET("/users/{id}/products", getUserProducts)

	app := lokstra.NewApp("route-parameters", ":3000", r)

	fmt.Println("🚀 Server running on http://localhost:3000")
	fmt.Println("📖 Try:")
	fmt.Println("   curl http://localhost:3000/users/1")
	fmt.Println("   curl http://localhost:3000/products?category=electronics")
	fmt.Println("   curl http://localhost:3000/products?category=furniture&max_price=250")
	fmt.Println("   curl -X PUT http://localhost:3000/users/1 -H 'Content-Type: application/json' -d '{\"name\":\"Alice Smith\"}'")

	app.Run(30 * time.Second)
}

// ============================================================================
// PATH PARAMETERS
// ============================================================================

type GetUserRequest struct {
	ID int `path:"id"` // Extract from URL path
}

func getUser(req *GetUserRequest) (*User, error) {
	for _, u := range users {
		if u.ID == req.ID {
			return &u, nil
		}
	}
	return nil, fmt.Errorf("user not found")
}

type UpdateUserRequest struct {
	ID    int    `path:"id"`    // From path
	Name  string `json:"name"`  // From body
	Email string `json:"email"` // From body
}

func updateUser(req *UpdateUserRequest) (*User, error) {
	for i, u := range users {
		if u.ID == req.ID {
			if req.Name != "" {
				users[i].Name = req.Name
			}
			if req.Email != "" {
				users[i].Email = req.Email
			}
			return &users[i], nil
		}
	}
	return nil, fmt.Errorf("user not found")
}

type DeleteUserRequest struct {
	ID int `path:"id"`
}

func deleteUser(req *DeleteUserRequest) (map[string]string, error) {
	for i, u := range users {
		if u.ID == req.ID {
			users = append(users[:i], users[i+1:]...)
			return map[string]string{
				"message": fmt.Sprintf("User %d deleted", req.ID),
			}, nil
		}
	}
	return nil, fmt.Errorf("user not found")
}

// ============================================================================
// QUERY PARAMETERS
// ============================================================================

type SearchProductsRequest struct {
	Category string  `query:"category"`           // Optional filter
	MinPrice float64 `query:"min_price"`          // Optional min price
	MaxPrice float64 `query:"max_price"`          // Optional max price
	Limit    int     `query:"limit" default:"10"` // With default value
}

func searchProducts(req *SearchProductsRequest) ([]Product, error) {
	result := []Product{}

	for _, p := range products {
		// Filter by category
		if req.Category != "" && p.Category != req.Category {
			continue
		}

		// Filter by min price
		if req.MinPrice > 0 && p.Price < req.MinPrice {
			continue
		}

		// Filter by max price
		if req.MaxPrice > 0 && p.Price > req.MaxPrice {
			continue
		}

		result = append(result, p)

		// Apply limit
		if len(result) >= req.Limit {
			break
		}
	}

	return result, nil
}

// ============================================================================
// COMBINED: PATH + QUERY PARAMETERS
// ============================================================================

type GetUserProductsRequest struct {
	UserID   int     `path:"id"`         // From path
	Category string  `query:"category"`  // From query
	MinPrice float64 `query:"min_price"` // From query
}

func getUserProducts(req *GetUserProductsRequest) (map[string]any, error) {
	// Get user
	var user *User
	for _, u := range users {
		if u.ID == req.UserID {
			user = &u
			break
		}
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Filter products
	filteredProducts := []Product{}
	for _, p := range products {
		if req.Category != "" && p.Category != req.Category {
			continue
		}
		if req.MinPrice > 0 && p.Price < req.MinPrice {
			continue
		}
		filteredProducts = append(filteredProducts, p)
	}

	return map[string]any{
		"user":     user,
		"products": filteredProducts,
		"count":    len(filteredProducts),
	}, nil
}
