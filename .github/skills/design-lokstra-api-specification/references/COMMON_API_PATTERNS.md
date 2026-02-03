# Common API Patterns

**Purpose**: Reusable API design patterns for pagination, filtering, sorting, bulk operations, search, and other common scenarios in REST APIs.

**Context**: These patterns provide consistency across all API endpoints and improve developer experience.

---

## Table of Contents

1. [Pagination Patterns](#pagination-patterns)
2. [Filtering Patterns](#filtering-patterns)
3. [Sorting Patterns](#sorting-patterns)
4. [Search Patterns](#search-patterns)
5. [Bulk Operations](#bulk-operations)
6. [Batch Processing](#batch-processing)
7. [Versioning Patterns](#versioning-patterns)
8. [Response Formats](#response-formats)

---

## Pagination Patterns

### Pattern 1: Cursor-Based Pagination (Recommended)

**When to Use**: Large datasets, real-time data, consistent results

```yaml
Request:
  GET /api/v1/patients?limit=20&cursor=eyJpZCI6InBhdF8xMDAifQ==

Query Parameters:
  limit:
    type: integer
    default: 20
    min: 1
    max: 100
    description: Number of items per page
  
  cursor:
    type: string
    format: base64
    description: Opaque cursor for next page
    example: eyJpZCI6InBhdF8xMDAiLCJ0ZW5hbnRfaWQiOiJjbGluaWNfMDAxIn0=

Response:
  {
    "status": "success",
    "data": [
      { "id": "pat_101", "name": "John Doe", ... },
      { "id": "pat_102", "name": "Jane Smith", ... },
      ...
    ],
    "pagination": {
      "limit": 20,
      "hasNext": true,
      "hasPrev": false,
      "nextCursor": "eyJpZCI6InBhdF8xMjAifQ==",
      "prevCursor": null
    }
  }

Cursor Encoding:
  # Encode
  cursor = base64.encode(json.dumps({
    "id": "pat_120",
    "tenant_id": "clinic_001",
    "created_at": "2024-01-20T10:30:00Z"
  }))
  
  # Decode
  cursor_data = json.loads(base64.decode(cursor))
```

**SQL Implementation**:
```sql
-- Forward pagination
SELECT * FROM patients
WHERE tenant_id = $1 
  AND (created_at, id) > ($2, $3)  -- from cursor
ORDER BY created_at, id
LIMIT $4;

-- Backward pagination
SELECT * FROM patients
WHERE tenant_id = $1 
  AND (created_at, id) < ($2, $3)  -- from cursor
ORDER BY created_at DESC, id DESC
LIMIT $4;
```

**Advantages**:
- Consistent results (no skipped/duplicated items)
- Good performance with indexes
- Works with real-time data
- No deep pagination penalty

**Disadvantages**:
- Can't jump to specific page
- More complex implementation
- Cursor must be opaque

### Pattern 2: Offset-Based Pagination

**When to Use**: Small datasets, need to jump to specific page

```yaml
Request:
  GET /api/v1/patients?page=2&perPage=20

Query Parameters:
  page:
    type: integer
    default: 1
    min: 1
    description: Page number (1-indexed)
  
  perPage:
    type: integer
    default: 20
    min: 1
    max: 100
    description: Items per page

Response:
  {
    "status": "success",
    "data": [...],
    "pagination": {
      "page": 2,
      "perPage": 20,
      "total": 150,
      "totalPages": 8,
      "hasNext": true,
      "hasPrev": true
    }
  }
```

**SQL Implementation**:
```sql
-- Get paginated data
SELECT * FROM patients
WHERE tenant_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;  -- LIMIT perPage OFFSET (page-1)*perPage

-- Get total count
SELECT COUNT(*) FROM patients WHERE tenant_id = $1;
```

**Advantages**:
- Simple implementation
- Can jump to any page
- Total count available

**Disadvantages**:
- Performance degrades with deep pagination
- Inconsistent results if data changes
- COUNT(*) can be slow

---

## Filtering Patterns

### Pattern 1: Query Parameter Filters

**Simple Filters**:
```yaml
Request:
  GET /api/v1/patients?status=active&gender=male

Query Parameters:
  status:
    type: string
    enum: [active, inactive, suspended]
    description: Filter by patient status
  
  gender:
    type: string
    enum: [male, female, other]
    description: Filter by gender

SQL:
  SELECT * FROM patients
  WHERE tenant_id = $1 
    AND status = $2
    AND gender = $3;
```

**Range Filters**:
```yaml
Request:
  GET /api/v1/appointments?dateFrom=2024-01-01&dateTo=2024-01-31

Query Parameters:
  dateFrom:
    type: string
    format: date
    description: Start date (inclusive)
  
  dateTo:
    type: string
    format: date
    description: End date (inclusive)

SQL:
  SELECT * FROM appointments
  WHERE tenant_id = $1
    AND appointment_date BETWEEN $2 AND $3;
```

**Array Filters**:
```yaml
Request:
  GET /api/v1/patients?status=active,pending&roles=doctor,nurse

Query Parameters:
  status:
    type: array
    items:
      type: string
    description: Filter by multiple statuses (comma-separated)
  
  roles:
    type: array
    items:
      type: string
    description: Filter by multiple roles (comma-separated)

SQL:
  SELECT * FROM patients
  WHERE tenant_id = $1
    AND status = ANY($2::text[])
    AND role = ANY($3::text[]);
```

### Pattern 2: Complex Filter Expressions

**JSON Query DSL**:
```yaml
Request:
  POST /api/v1/patients/search
  
Body:
  {
    "filters": {
      "and": [
        { "field": "status", "op": "eq", "value": "active" },
        { "field": "age", "op": "gte", "value": 18 },
        {
          "or": [
            { "field": "city", "op": "eq", "value": "Jakarta" },
            { "field": "city", "op": "eq", "value": "Surabaya" }
          ]
        }
      ]
    }
  }

Supported Operators:
  - eq: Equal
  - ne: Not equal
  - gt: Greater than
  - gte: Greater than or equal
  - lt: Less than
  - lte: Less than or equal
  - in: In array
  - nin: Not in array
  - like: SQL LIKE pattern
  - ilike: Case-insensitive LIKE
  - between: Range
```

**Implementation**:
```go
type Filter struct {
	Field string      `json:"field"`
	Op    string      `json:"op"`
	Value interface{} `json:"value"`
	And   []Filter    `json:"and,omitempty"`
	Or    []Filter    `json:"or,omitempty"`
}

func BuildWhereClause(filter Filter, args *[]interface{}) string {
	if len(filter.And) > 0 {
		clauses := []string{}
		for _, f := range filter.And {
			clauses = append(clauses, BuildWhereClause(f, args))
		}
		return "(" + strings.Join(clauses, " AND ") + ")"
	}
	
	if len(filter.Or) > 0 {
		clauses := []string{}
		for _, f := range filter.Or {
			clauses = append(clauses, BuildWhereClause(f, args))
		}
		return "(" + strings.Join(clauses, " OR ") + ")"
	}
	
	// Build simple condition
	*args = append(*args, filter.Value)
	argNum := len(*args)
	
	switch filter.Op {
	case "eq":
		return fmt.Sprintf("%s = $%d", filter.Field, argNum)
	case "gt":
		return fmt.Sprintf("%s > $%d", filter.Field, argNum)
	case "gte":
		return fmt.Sprintf("%s >= $%d", filter.Field, argNum)
	case "like":
		return fmt.Sprintf("%s LIKE $%d", filter.Field, argNum)
	// ... more operators
	}
}
```

---

## Sorting Patterns

### Pattern 1: Simple Sorting

```yaml
Request:
  GET /api/v1/patients?sortBy=name&sortOrder=asc

Query Parameters:
  sortBy:
    type: string
    enum: [name, email, createdAt, updatedAt]
    default: createdAt
    description: Field to sort by
  
  sortOrder:
    type: string
    enum: [asc, desc]
    default: desc
    description: Sort direction

SQL:
  SELECT * FROM patients
  WHERE tenant_id = $1
  ORDER BY 
    CASE WHEN $2 = 'name' THEN name END ASC,
    CASE WHEN $2 = 'createdAt' THEN created_at END DESC;
```

### Pattern 2: Multi-Field Sorting

```yaml
Request:
  GET /api/v1/patients?sort=status:asc,name:asc,createdAt:desc

Query Parameters:
  sort:
    type: string
    description: Comma-separated sort fields with direction
    format: field:direction,field:direction,...
    example: status:asc,name:asc

Parsing:
  sorts = [
    { field: "status", direction: "asc" },
    { field: "name", direction: "asc" },
    { field: "createdAt", direction: "desc" }
  ]

SQL:
  SELECT * FROM patients
  WHERE tenant_id = $1
  ORDER BY status ASC, name ASC, created_at DESC;
```

---

## Search Patterns

### Pattern 1: Simple Text Search

```yaml
Request:
  GET /api/v1/patients?q=john

Query Parameters:
  q:
    type: string
    minLength: 2
    description: Search query
    example: john

SQL (PostgreSQL):
  SELECT * FROM patients
  WHERE tenant_id = $1
    AND (
      name ILIKE '%' || $2 || '%' 
      OR email ILIKE '%' || $2 || '%'
      OR phone ILIKE '%' || $2 || '%'
    )
  ORDER BY created_at DESC;
```

### Pattern 2: Full-Text Search

```yaml
Request:
  GET /api/v1/patients?search=diabetes+medication&searchFields=diagnosis,notes

Query Parameters:
  search:
    type: string
    description: Full-text search query
    example: diabetes medication
  
  searchFields:
    type: array
    items:
      type: string
    description: Fields to search in

SQL (PostgreSQL with tsvector):
  SELECT * FROM patients
  WHERE tenant_id = $1
    AND search_vector @@ to_tsquery('english', $2)
  ORDER BY ts_rank(search_vector, to_tsquery('english', $2)) DESC;

Index:
  CREATE INDEX idx_patients_fts 
    ON patients USING GIN(search_vector)
    WHERE tenant_id IS NOT NULL;
```

### Pattern 3: Autocomplete/Typeahead

```yaml
Request:
  GET /api/v1/patients/autocomplete?q=joh&limit=10

Query Parameters:
  q:
    type: string
    minLength: 2
    maxLength: 50
    description: Search prefix
  
  limit:
    type: integer
    default: 10
    max: 20

Response:
  {
    "status": "success",
    "data": [
      { "id": "pat_123", "name": "John Doe", "email": "john@example.com" },
      { "id": "pat_456", "name": "John Smith", "email": "jsmith@example.com" }
    ]
  }

SQL (with trigram index):
  SELECT id, name, email
  FROM patients
  WHERE tenant_id = $1
    AND name ILIKE $2 || '%'
  ORDER BY similarity(name, $2) DESC
  LIMIT $3;

Index:
  CREATE INDEX idx_patients_name_trgm 
    ON patients USING gin(name gin_trgm_ops)
    WHERE tenant_id IS NOT NULL;
```

---

## Bulk Operations

### Pattern 1: Bulk Create

```yaml
Request:
  POST /api/v1/patients/bulk

Body:
  {
    "patients": [
      { "name": "John Doe", "email": "john@example.com", ... },
      { "name": "Jane Smith", "email": "jane@example.com", ... },
      ...
    ]
  }

Validation:
  - Max 100 items per request
  - All items validated before processing
  - Atomic operation (all or nothing)

Response (201 Created):
  {
    "status": "success",
    "message": "50 patients created successfully",
    "data": {
      "created": 50,
      "failed": 0,
      "results": [
        { "id": "pat_123", "name": "John Doe", "status": "created" },
        { "id": "pat_124", "name": "Jane Smith", "status": "created" },
        ...
      ]
    }
  }

Error Response (400 Bad Request):
  {
    "status": "error",
    "error": {
      "code": "BULK_VALIDATION_FAILED",
      "message": "2 items failed validation",
      "details": [
        { "index": 5, "field": "email", "message": "Invalid email format" },
        { "index": 12, "field": "phone", "message": "Phone number required" }
      ]
    }
  }
```

**Implementation**:
```go
func BulkCreatePatients(ctx context.Context, tenantID string, patients []Patient) error {
	// Validate all items first
	for i, p := range patients {
		if err := validate.Struct(p); err != nil {
			return fmt.Errorf("validation failed at index %d: %w", i, err)
		}
	}
	
	// Start transaction
	tx, _ := db.BeginTx(ctx, nil)
	defer tx.Rollback()
	
	// Insert in batch
	stmt, _ := tx.Prepare(`
		INSERT INTO patients (id, tenant_id, name, email, phone)
		VALUES ($1, $2, $3, $4, $5)
	`)
	
	for _, p := range patients {
		_, err := stmt.Exec(p.ID, tenantID, p.Name, p.Email, p.Phone)
		if err != nil {
			return err
		}
	}
	
	return tx.Commit()
}
```

### Pattern 2: Bulk Update

```yaml
Request:
  PATCH /api/v1/patients/bulk

Body:
  {
    "updates": [
      { "id": "pat_123", "status": "active" },
      { "id": "pat_456", "phone": "+6281234567890" },
      ...
    ]
  }

Response (200 OK):
  {
    "status": "success",
    "message": "45 patients updated, 5 failed",
    "data": {
      "updated": 45,
      "failed": 5,
      "results": [
        { "id": "pat_123", "status": "updated" },
        { "id": "pat_456", "status": "updated" },
        { "id": "pat_789", "status": "failed", "error": "Patient not found" },
        ...
      ]
    }
  }
```

### Pattern 3: Bulk Delete

```yaml
Request:
  DELETE /api/v1/patients/bulk

Body:
  {
    "ids": ["pat_123", "pat_456", "pat_789"]
  }

Validation:
  - Max 100 IDs per request
  - All IDs must exist in tenant
  - Soft delete (set deleted_at)

Response (200 OK):
  {
    "status": "success",
    "message": "3 patients deleted",
    "data": {
      "deleted": 3,
      "failed": 0
    }
  }
```

---

## Batch Processing

### Pattern: Async Batch with Status Tracking

```yaml
# Step 1: Submit batch job
POST /api/v1/patients/import

Body:
  {
    "fileUrl": "https://s3.amazonaws.com/bucket/patients.csv",
    "options": {
      "skipDuplicates": true,
      "notifyOnComplete": true
    }
  }

Response (202 Accepted):
  {
    "status": "accepted",
    "data": {
      "jobId": "job_abc123",
      "status": "pending",
      "createdAt": "2024-01-20T10:30:00Z",
      "estimatedDuration": 300
    }
  }

# Step 2: Check status
GET /api/v1/jobs/job_abc123

Response (200 OK):
  {
    "status": "success",
    "data": {
      "jobId": "job_abc123",
      "status": "processing",
      "progress": {
        "total": 1000,
        "processed": 650,
        "succeeded": 640,
        "failed": 10,
        "percentage": 65
      },
      "startedAt": "2024-01-20T10:30:00Z",
      "estimatedCompletion": "2024-01-20T10:35:00Z"
    }
  }

# Step 3: Get results when complete
GET /api/v1/jobs/job_abc123/results

Response (200 OK):
  {
    "status": "success",
    "data": {
      "jobId": "job_abc123",
      "status": "completed",
      "results": {
        "total": 1000,
        "succeeded": 990,
        "failed": 10,
        "duration": 295,
        "errors": [
          { "row": 45, "error": "Invalid email format" },
          { "row": 123, "error": "Duplicate entry" },
          ...
        ]
      },
      "completedAt": "2024-01-20T10:35:00Z"
    }
  }
```

---

## Versioning Patterns

### Pattern 1: URL Path Versioning (Recommended)

```yaml
Format: /api/{version}/{resource}

Examples:
  /api/v1/patients
  /api/v2/patients
  /api/v1/appointments

Advantages:
  - Clear and explicit
  - Easy to route
  - Standard practice

Disadvantages:
  - URL changes with version
  - Need to maintain multiple versions
```

### Pattern 2: Header Versioning

```yaml
Request:
  GET /api/patients
  Headers:
    API-Version: 2024-01-20
    Accept: application/vnd.myapi.v2+json

Advantages:
  - Clean URLs
  - Fine-grained control

Disadvantages:
  - Less discoverable
  - More complex routing
```

### Pattern 3: Version Deprecation

```yaml
Deprecation Process:
  1. Announce deprecation 6 months in advance
  2. Add Deprecation header to responses
  3. Document migration guide
  4. Set sunset date
  5. Remove old version

Response Headers:
  Deprecation: true
  Sunset: Wed, 01 Jul 2024 00:00:00 GMT
  Link: <https://api.example.com/docs/v2>; rel="successor-version"

Response Body:
  {
    "status": "success",
    "data": {...},
    "warnings": [
      {
        "code": "API_VERSION_DEPRECATED",
        "message": "API v1 will be sunset on 2024-07-01. Please migrate to v2.",
        "link": "https://api.example.com/docs/migration-v1-v2"
      }
    ]
  }
```

---

## Response Formats

### Pattern 1: Standard Success Response

```json
{
  "status": "success",
  "message": "Operation completed successfully",
  "data": {
    // Response payload
  },
  "meta": {
    "requestId": "req_abc123",
    "timestamp": "2024-01-20T10:30:00Z",
    "version": "1.0.0"
  }
}
```

### Pattern 2: Standard Error Response

```json
{
  "status": "error",
  "error": {
    "code": "RESOURCE_NOT_FOUND",
    "message": "Patient not found",
    "field": "patientId",
    "details": [
      {
        "field": "patientId",
        "message": "Patient with ID pat_123 does not exist",
        "code": "NOT_FOUND"
      }
    ]
  },
  "meta": {
    "requestId": "req_abc123",
    "timestamp": "2024-01-20T10:30:00Z"
  }
}
```

### Pattern 3: Collection Response

```json
{
  "status": "success",
  "data": [
    { "id": "pat_123", "name": "John Doe" },
    { "id": "pat_456", "name": "Jane Smith" }
  ],
  "pagination": {
    "limit": 20,
    "hasNext": true,
    "hasPrev": false,
    "nextCursor": "eyJpZCI6InBhdF80NTYifQ==",
    "prevCursor": null
  },
  "meta": {
    "count": 2,
    "total": 150,
    "requestId": "req_abc123"
  }
}
```

### Pattern 4: Partial Response (Field Selection)

```yaml
Request:
  GET /api/v1/patients/pat_123?fields=id,name,email

Query Parameters:
  fields:
    type: string
    description: Comma-separated fields to include
    example: id,name,email

Response:
  {
    "status": "success",
    "data": {
      "id": "pat_123",
      "name": "John Doe",
      "email": "john@example.com"
      // Other fields excluded
    }
  }
```

---

## Rate Limiting Response

```yaml
Response Headers:
  X-RateLimit-Limit: 100
  X-RateLimit-Remaining: 95
  X-RateLimit-Reset: 1705752000
  Retry-After: 60

Error Response (429 Too Many Requests):
  {
    "status": "error",
    "error": {
      "code": "RATE_LIMIT_EXCEEDED",
      "message": "Rate limit exceeded. Try again in 60 seconds.",
      "retryAfter": 60
    },
    "meta": {
      "requestId": "req_abc123",
      "rateLimit": {
        "limit": 100,
        "remaining": 0,
        "reset": 1705752000
      }
    }
  }
```

---

## Best Practices Summary

### Query Parameters
- Use camelCase for consistency
- Provide sensible defaults
- Document min/max values
- Validate all inputs

### Pagination
- Use cursor-based for large datasets
- Limit max page size (e.g., 100)
- Include hasNext/hasPrev flags
- Provide total count when feasible

### Filtering
- Support common operators (eq, gt, in, like)
- Allow multiple filters
- Document available fields
- Validate filter values

### Sorting
- Support multi-field sorting
- Default to logical sort (e.g., createdAt desc)
- Whitelist sortable fields
- Combine with pagination

### Search
- Require minimum query length (2-3 chars)
- Use full-text search for performance
- Index searchable fields
- Return relevance scores

### Bulk Operations
- Limit batch size (e.g., 100 items)
- Validate all items before processing
- Use transactions for consistency
- Provide detailed error feedback

---

**File Size**: 16 KB  
**Last Updated**: 2024-01-20  
**Related**: API_SPEC_TEMPLATE.md, MULTI_TENANT_API_PATTERNS.md
