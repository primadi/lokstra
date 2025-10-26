package contract

// ============================================================================
// User Service Contracts
// ============================================================================

type GetUserParams struct {
	ID int `path:"id"`
}

type ListUsersParams struct {
	Role string `query:"role"`
}

// ============================================================================
// Product Service Contracts
// ============================================================================

type GetProductParams struct {
	ID int `path:"id"`
}

type ListProductsParams struct {
	Category string `query:"category"`
}
