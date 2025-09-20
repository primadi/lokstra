package apps

import (
	"github.com/primadi/lokstra/core/request"
)

// Dashboard HTMX handlers
func DashboardHome(ctx *request.Context) error {
	html := `
<!DOCTYPE html>
<html>
<head>
    <title>Admin Dashboard</title>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .container { max-width: 1200px; margin: 0 auto; }
        .card { border: 1px solid #ddd; padding: 20px; margin: 10px 0; border-radius: 8px; }
        .btn { background: #007bff; color: white; padding: 10px 20px; border: none; cursor: pointer; border-radius: 4px; }
        .btn:hover { background: #0056b3; }
        table { width: 100%; border-collapse: collapse; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Admin Dashboard</h1>
        
        <div class="card">
            <h2>Users</h2>
            <button class="btn" hx-get="/dashboard/users" hx-target="#users-content">Load Users</button>
            <div id="users-content"></div>
        </div>
        
        <div class="card">
            <h2>Orders</h2>
            <button class="btn" hx-get="/dashboard/orders" hx-target="#orders-content">Load Orders</button>
            <div id="orders-content"></div>
        </div>
    </div>
</body>
</html>`

	return ctx.HTML(200, html)
}

func DashboardUsers(ctx *request.Context) error {
	html := `
<table>
    <thead>
        <tr>
            <th>ID</th>
            <th>Name</th>
            <th>Email</th>
        </tr>
    </thead>
    <tbody>
        <tr>
            <td>1</td>
            <td>John Doe</td>
            <td>john@example.com</td>
        </tr>
        <tr>
            <td>2</td>
            <td>Jane Smith</td>
            <td>jane@example.com</td>
        </tr>
    </tbody>
</table>`

	return ctx.HTML(200, html)
}

func DashboardOrders(ctx *request.Context) error {
	html := `
<table>
    <thead>
        <tr>
            <th>ID</th>
            <th>User ID</th>
            <th>Product</th>
            <th>Amount</th>
            <th>Status</th>
        </tr>
    </thead>
    <tbody>
        <tr>
            <td>1</td>
            <td>1</td>
            <td>Laptop</td>
            <td>$999.99</td>
            <td>pending</td>
        </tr>
        <tr>
            <td>2</td>
            <td>2</td>
            <td>Phone</td>
            <td>$599.99</td>
            <td>completed</td>
        </tr>
    </tbody>
</table>`

	return ctx.HTML(200, html)
}
