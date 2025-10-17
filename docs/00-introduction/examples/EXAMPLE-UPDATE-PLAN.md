# Example Update Plan - New Paradigm

## Strategy: Selective Updates

**Philosophy:** Only update examples where new paradigm provides clear benefits.

---

## ❌ NO CHANGES (Keep Simple)

### 01-hello-world
**Reason:** Too simple - no services, no config needs

```go
// Perfect as-is!
r := lokstra.NewRouter("api")
r.GET("/", func() string {
    return "Hello, Lokstra!"
})
app := lokstra.NewApp("hello", ":3000", r)
app.Run(30 * time.Second)
```

**Status:** ✅ **DONE** - No changes made

---

### 02-handler-forms  
**Reason:** Demonstrates framework features (handler forms), not configuration

**What it shows:**
- 29 different handler signatures
- Request binding (JSON, path, query, header)
- Response patterns
- Context access

**Status:** ✅ **DONE** - No changes made

---

## ✅ UPDATED (Clear Benefits)

### 03-crud-api
**Benefit:** Service dependencies → Lazy DI eliminates manual wiring

**Changes:**
1. Add `config.yaml` with service definitions
2. Convert `NewDatabase()` → database factory
3. Convert `NewUserService(db)` → factory with lazy DB injection
4. Update `main.go` to use loader

**Files to modify:**
- ✅ `main.go` - Add factory registration + loader
- ✅ Add `config.yaml` - Service configuration
- ✅ Add `MIGRATION.md` - Explain changes
- ✅ Update `README.md` - Document new approach

**Status:** ⏳ **TODO**

---

### 04-multi-deployment
**Benefit:** Multiple deployments → One YAML config replaces flag parsing

**Changes:**
1. Replace flag-based mode selection with deployment config
2. Create single `config.yaml` with 3 deployments
3. Simplify main.go - remove conditional logic
4. Service definitions shared across deployments

**Files to modify:**
- ✅ `main.go` - Replace flags with loader
- ✅ Add `config.yaml` - Multi-deployment config
- ✅ Add `MIGRATION.md` - Before/after comparison
- ✅ Update `README.md` - New usage

**Status:** ⏳ **TODO**

---

## Implementation Order

### Phase 1: 03-crud-api ✅
**Why first:** Simpler, demonstrates lazy DI basics

**Steps:**
1. Create `config.yaml`
2. Create factory functions for Database and UserService
3. Update main.go to use loader
4. Create MIGRATION.md
5. Update README.md
6. Test all endpoints

**Estimated:** 30 minutes

---

### Phase 2: 04-multi-deployment ✅
**Why second:** More complex, demonstrates full deployment power

**Steps:**
1. Create `config.yaml` with 3 deployments
2. Convert service registration to factories
3. Simplify main.go (remove flags/conditions)
4. Create MIGRATION.md
5. Update README.md
6. Test all 3 deployment modes

**Estimated:** 45 minutes

---

## Files Structure After Update

```
docs/00-introduction/
├── examples/                    # NEW PARADIGM
│   ├── README.md               # ✅ Overview + comparison
│   ├── 01-hello-world/         # ❌ NO CHANGES
│   │   ├── main.go
│   │   ├── README.md
│   │   └── test.http
│   ├── 02-handler-forms/       # ❌ NO CHANGES
│   │   ├── main.go
│   │   ├── README.md
│   │   └── test.http
│   ├── 03-crud-api/            # ✅ UPDATED
│   │   ├── main.go            # Modified
│   │   ├── config.yaml        # NEW
│   │   ├── MIGRATION.md       # NEW
│   │   ├── README.md          # Updated
│   │   └── test.http          # Unchanged
│   └── 04-multi-deployment/   # ✅ UPDATED
│       ├── main.go            # Modified
│       ├── config.yaml        # NEW
│       ├── MIGRATION.md       # NEW
│       ├── README.md          # Updated
│       └── test.http          # Unchanged
│
└── examples_old/               # OLD PARADIGM (backup)
    ├── 01-hello-world/
    ├── 02-handler-forms/
    ├── 03-crud-api/
    └── 04-multi-deployment/
```

---

## Testing Checklist

### For Each Updated Example:

- [ ] **Code compiles** - `go build`
- [ ] **Example runs** - `go run main.go`
- [ ] **All endpoints work** - Test with curl/test.http
- [ ] **Config validates** - YAML schema passes
- [ ] **README updated** - Documents new approach
- [ ] **MIGRATION.md created** - Explains changes

### For 03-crud-api:
- [ ] GET /users returns users
- [ ] GET /users/{id} returns specific user
- [ ] POST /users creates new user
- [ ] PUT /users/{id} updates user
- [ ] DELETE /users/{id} deletes user
- [ ] Database service lazy-loaded
- [ ] UserService gets DB via lazy DI

### For 04-multi-deployment:
- [ ] Monolith mode runs on :3003
- [ ] User service mode runs on :3004
- [ ] Order service mode runs on :3005
- [ ] All services work in monolith
- [ ] User endpoints work in user service
- [ ] Order endpoints work in order service
- [ ] Config shared across deployments

---

## Success Criteria

### Code Quality:
- ✅ All examples compile and run
- ✅ No breaking changes to public APIs
- ✅ Old examples preserved in examples_old/
- ✅ Clear documentation of changes

### Documentation:
- ✅ README.md explains new approach
- ✅ MIGRATION.md shows before/after
- ✅ Examples demonstrate benefits
- ✅ Clear guidance on when to use each approach

### User Experience:
- ✅ Easy to understand which examples changed
- ✅ Clear rationale for changes
- ✅ Old approach still accessible
- ✅ Migration path documented

---

## Timeline

**Total Estimate:** 2-3 hours

| Task | Time | Status |
|------|------|--------|
| Setup & Planning | 30min | ✅ Done |
| Update 03-crud-api | 30min | ⏳ Next |
| Update 04-multi-deployment | 45min | ⏳ After |
| Testing & Documentation | 30min | ⏳ Final |
| Review & Polish | 15min | ⏳ Final |

---

## Notes

1. **Keep it simple** - Don't over-engineer examples
2. **Show benefits** - Make improvements obvious
3. **Preserve old** - examples_old/ as reference
4. **Clear docs** - Explain why and how
5. **Test everything** - All endpoints must work

---

*This plan ensures examples demonstrate new paradigm where it provides clear value, while keeping simple examples simple.*
