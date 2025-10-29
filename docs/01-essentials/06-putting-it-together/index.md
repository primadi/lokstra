# Todo API - Complete Example

A complete REST API built with Lokstra's auto-router feature.

## What You'll Learn

- Auto-generate routes from service methods
- Type-safe request parameters
- Convention-based routing (REST)
- Input validation
- Clean code structure (3 files!)

## Project Structure

```
.
├── models.go   - Data structures and validation
├── service.go  - Business logic
└── main.go     - Auto-router setup
```

## Running

```bash
# Navigate to example directory
cd docs/01-essentials/06-putting-it-together

# Run directly (go.mod already exists in project root)
go run .
```

## Testing

Use the included `test.http` file with VS Code REST Client extension, or use curl:

**Create todo:**
```bash
curl -X POST http://localhost:3000/todos \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Learn Lokstra",
    "description": "Complete all essentials tutorials"
  }'
```

**Get all todos:**
```bash
curl http://localhost:3000/todos
```

**Get single todo:**
```bash
curl http://localhost:3000/todos/1
```

**Update todo (partial update with pointers):**
```bash
curl -X PUT http://localhost:3000/todos/1 \
  -H "Content-Type: application/json" \
  -d '{"completed": true}'
```

**Delete todo:**
```bash
curl -X DELETE http://localhost:3000/todos/1
```

**Test validation (should fail):**
```bash
curl -X POST http://localhost:3000/todos \
  -H "Content-Type: application/json" \
  -d '{"title": "ab", "description": "Too short"}'
```

> 💡 **Tip**: Open `test.http` in VS Code for interactive testing with the REST Client extension!

## How It Works

The auto-router generates routes from service methods:

| Service Method | Generated Route | HTTP Method |
|----------------|----------------|-------------|
| `Create(params)` | `/todos` | POST |
| `List()` | `/todos` | GET |
| `GetByID(params)` | `/todos/{id}` | GET |
| `Update(params)` | `/todos/{id}` | PUT |
| `Delete(params)` | `/todos/{id}` | DELETE |

**No manual route registration needed!**

## Key Features

✅ Auto-generated REST routes  
✅ Type-safe parameters  
✅ Input validation  
✅ Thread-safe service  
✅ Graceful shutdown  
✅ Only 3 files (~200 lines total)
