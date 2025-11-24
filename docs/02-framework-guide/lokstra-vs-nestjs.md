---
layout: docs
title: Lokstra vs NestJS - Framework Comparison
---

# Lokstra vs NestJS - Framework Comparison

> **Detailed comparison between Lokstra (Go) and NestJS (TypeScript/Node.js)**

Both Lokstra and NestJS are enterprise-grade frameworks that emphasize **dependency injection**, **modular architecture**, and **convention over configuration**. Here's how they compare:

---

## 🎯 Quick Overview

| Aspect | Lokstra (Go) | NestJS (TypeScript) |
|--------|--------------|-------------------|
| **Language** | Go | TypeScript/Node.js |
| **Architecture** | Service-oriented with DI | Module-based with DI |
| **DI Pattern** | Lazy, type-safe generics | Decorator-based reflection |
| **Router Generation** | ✅ Auto from service methods | ✅ Auto from controller decorators |
| **Configuration** | YAML + Code (flexible) | TypeScript + Environment |
| **Deployment** | ✅ Zero-code topology change | Requires code changes |
| **Performance** | Compiled binary, fast startup | Runtime compilation, slower startup |

---

## 🏗️ Architecture Comparison

### Lokstra: Service-Oriented Architecture

```go
// 1. Define Service
type UserService struct {
    db *Database
}

func (s *UserService) GetAll() ([]User, error) {
    return s.db.Query("SELECT * FROM users")
}

func (s *UserService) GetByID(id string) (*User, error) {
    return s.db.QueryOne("SELECT * FROM users WHERE id = ?", id)
}

// 2. Register Service Factory
lokstra_registry.RegisterServiceType(
    "user-service-factory",
    NewUserService,
    nil,
    deploy.WithResource("user", "users"),
    deploy.WithConvention("rest"),
)

// 3. Auto-generate Router
userRouter := lokstra_registry.NewRouterFromServiceType("user-service-factory")
// Creates: GET /users, GET /users/{id}, etc.
```

### NestJS: Module + Controller Architecture

```typescript
// 1. Define Service
@Injectable()
export class UserService {
  constructor(private db: Database) {}

  async getAll(): Promise<User[]> {
    return this.db.query('SELECT * FROM users');
  }

  async getById(id: string): Promise<User> {
    return this.db.queryOne('SELECT * FROM users WHERE id = ?', [id]);
  }
}

// 2. Define Controller
@Controller('users')
export class UserController {
  constructor(private userService: UserService) {}

  @Get()
  async getAll() {
    return this.userService.getAll();
  }

  @Get(':id')
  async getById(@Param('id') id: string) {
    return this.userService.getById(id);
  }
}

// 3. Module Registration
@Module({
  controllers: [UserController],
  providers: [UserService],
})
export class UserModule {}
```

---

## 🔌 Dependency Injection Comparison

### Lokstra: Lazy Loading with Generics

```go
// Type-safe, lazy loading
var userService = service.LazyLoad[*UserService]("user-service")
var database = service.LazyLoad[*Database]("database")

func handler() {
    // Loaded on first access, cached forever
    users := userService.MustGet().GetAll()
}

// Factory with dependencies
func NewUserService() *UserService {
    return &UserService{
        db: service.LazyLoad[*Database]("database"),
    }
}

// Register with dependencies
lokstra_registry.RegisterServiceFactory("user-service", NewUserService)
lokstra_registry.RegisterServiceFactory("database", NewDatabase)
```

**Lokstra DI Advantages:**
- ✅ **Type-safe**: Compile-time type checking with generics
- ✅ **Lazy**: Services created only when needed
- ✅ **Performance**: No reflection overhead
- ✅ **Simple**: No decorators, just functions

### NestJS: Decorator-based DI

```typescript
// Constructor injection
@Injectable()
export class UserService {
  constructor(
    private database: Database,
    private logger: Logger,
  ) {}
}

// Provider registration
@Module({
  providers: [
    UserService,
    Database,
    Logger,
  ],
})
export class AppModule {}
```

**NestJS DI Advantages:**
- ✅ **Familiar**: Similar to Angular/Spring
- ✅ **Automatic**: Dependency resolution via reflection
- ✅ **Rich ecosystem**: Many built-in providers
- ⚠️ **Runtime overhead**: Reflection-based

---

## 🚦 Router Generation Comparison

### Lokstra: Convention-based from Service Methods

```go
// Service method signatures determine routes
func (s *UserService) GetAll(p *GetAllParams) ([]User, error)     // GET /users
func (s *UserService) GetByID(p *GetByIDParams) (*User, error)    // GET /users/{id}
func (s *UserService) Create(p *CreateParams) (*User, error)      // POST /users
func (s *UserService) Update(p *UpdateParams) (*User, error)      // PUT /users/{id}
func (s *UserService) Delete(p *DeleteParams) error              // DELETE /users/{id}

// Auto-router generation
router := lokstra_registry.NewRouterFromServiceType("user-service-factory")
```

**Lokstra Approach:**
- ✅ **Zero boilerplate**: No controller layer needed
- ✅ **Convention over configuration**: Method names → HTTP routes
- ✅ **Type-safe parameters**: Struct-based parameter binding
- ✅ **Flexible**: Can override routes if needed

### NestJS: Decorator-driven Routes

```typescript
@Controller('users')
export class UserController {
  @Get()                    // GET /users
  getAll() { ... }

  @Get(':id')              // GET /users/:id
  getById(@Param('id') id: string) { ... }

  @Post()                  // POST /users
  create(@Body() data: CreateUserDto) { ... }

  @Put(':id')              // PUT /users/:id
  update(@Param('id') id: string, @Body() data: UpdateUserDto) { ... }

  @Delete(':id')           // DELETE /users/:id
  delete(@Param('id') id: string) { ... }
}
```

**NestJS Approach:**
- ✅ **Explicit**: Clear route definitions
- ✅ **Flexible**: Rich decorator options
- ✅ **Validation**: Built-in DTO validation
- ⚠️ **Boilerplate**: Need controller layer + service layer

---

## 📝 Configuration & Deployment

### Lokstra: YAML + Code Configuration

```yaml
# config.yaml - Deployment topology
service-definitions:
  user-service:
    type: user-service-factory
    depends-on: [database]

deployments:
  monolith:
    servers:
      api-server:
        addr: ":8080"
        published-services: [user-service, order-service]
  
  microservices:
    servers:
      user-service:
        addr: ":8001"
        published-services: [user-service]
      order-service:
        addr: ":8002"
        published-services: [order-service]
```

```bash
# Same binary, different topologies!
./app -server=monolith.api-server      # Monolith
./app -server=microservices.user-service  # Microservice
```

**Lokstra Deployment:**
- ✅ **Zero-code deployment changes**: Same binary, different config
- ✅ **Flexible topology**: Monolith ↔ Microservices without code changes
- ✅ **Environment overrides**: Command-line params > ENV > YAML defaults

### NestJS: Code + Environment Configuration

```typescript
// Different apps for different deployments
@Module({
  imports: [UserModule, OrderModule, PaymentModule],
})
export class MonolithApp {}

@Module({
  imports: [UserModule],
})
export class UserMicroservice {}

@Module({
  imports: [OrderModule],
})
export class OrderMicroservice {}

// Different main.ts files or conditional imports
async function bootstrap() {
  const app = await NestFactory.create(
    process.env.SERVICE_TYPE === 'user' ? UserMicroservice : MonolithApp
  );
  await app.listen(process.env.PORT || 3000);
}
```

**NestJS Deployment:**
- ⚠️ **Code changes required**: Different modules for different deployments
- ✅ **Rich configuration**: ConfigModule, environment validation
- ✅ **Good tooling**: CLI, testing utilities

---

## ⚡ Performance Comparison

### Lokstra (Go)
- ✅ **Fast startup**: Compiled binary, instant startup
- ✅ **Low memory**: Efficient memory usage
- ✅ **High throughput**: Go's goroutines for concurrency
- ✅ **No GC pauses**: Predictable performance
- ✅ **Small binaries**: Single executable file

### NestJS (Node.js)
- ⚠️ **Slower startup**: Runtime compilation, module resolution
- ⚠️ **Higher memory**: V8 engine overhead
- ✅ **Good throughput**: Event loop for I/O-bound tasks
- ⚠️ **GC pauses**: V8 garbage collection
- ⚠️ **Larger deployments**: node_modules + runtime

---

## 🧪 Testing Comparison

### Lokstra Testing

```go
func TestUserService(t *testing.T) {
    // Mock dependencies
    mockDB := &MockDatabase{}
    
    // Create service with mocked deps
    service := &UserService{db: mockDB}
    
    // Test service methods
    users, err := service.GetAll(&GetAllParams{})
    assert.NoError(t, err)
    assert.Len(t, users, 2)
}

// Integration testing with registry
func TestUserServiceWithRegistry(t *testing.T) {
    lokstra_registry.RegisterServiceFactory("database", NewMockDatabase)
    lokstra_registry.RegisterServiceFactory("user-service", NewUserService)
    
    userService := lokstra_registry.GetService[*UserService]("user-service")
    // Test with real DI container
}
```

### NestJS Testing

```typescript
describe('UserService', () => {
  let service: UserService;
  let database: Database;

  beforeEach(async () => {
    const module: TestingModule = await Test.createTestingModule({
      providers: [
        UserService,
        {
          provide: Database,
          useValue: mockDatabase,
        },
      ],
    }).compile();

    service = module.get<UserService>(UserService);
    database = module.get<Database>(Database);
  });

  it('should return users', async () => {
    const users = await service.getAll();
    expect(users).toHaveLength(2);
  });
});
```

---

## 🎯 When to Choose Which?

### Choose Lokstra When:
- ✅ **Performance is critical**: Need fast startup and low latency
- ✅ **Deployment flexibility**: Want monolith ↔ microservices flexibility
- ✅ **Type safety**: Prefer compile-time safety over runtime flexibility  
- ✅ **Simple deployment**: Want single binary deployment
- ✅ **Go ecosystem**: Team familiar with Go
- ✅ **Zero boilerplate**: Want auto-router without controller layer

### Choose NestJS When:
- ✅ **Rich ecosystem**: Need extensive package ecosystem
- ✅ **Team expertise**: Team familiar with TypeScript/Angular
- ✅ **Rapid development**: Need quick prototyping with decorators
- ✅ **Frontend integration**: Building full-stack TypeScript apps
- ✅ **Mature tooling**: Need CLI, testing, and debugging tools
- ✅ **Enterprise patterns**: Want familiar Spring Boot-like patterns

---

## 🏆 Summary Comparison

| Criteria | Lokstra | NestJS | Winner |
|----------|---------|--------|--------|
| **Performance** | Compiled, fast startup | Runtime, slower startup | 🏆 Lokstra |
| **Type Safety** | Compile-time with generics | Runtime with decorators | 🏆 Lokstra |
| **Ecosystem** | Growing Go ecosystem | Mature Node.js ecosystem | 🏆 NestJS |
| **Learning Curve** | Simple, less magic | More concepts, more magic | 🏆 Lokstra |
| **Deployment** | Zero-code topology change | Requires code changes | 🏆 Lokstra |
| **Development Speed** | Good with auto-router | Very fast with decorators | 🏆 NestJS |
| **Enterprise Features** | Service-oriented, DI, config | Modules, guards, pipes, interceptors | 🤝 Tie |
| **Community** | Growing | Very mature | 🏆 NestJS |

---

## 🚀 Migration Path

### From NestJS to Lokstra:

```typescript
// NestJS Controller + Service
@Controller('users')
export class UserController {
  constructor(private userService: UserService) {}
  
  @Get()
  getAll() { return this.userService.getAll(); }
}

@Injectable()
export class UserService {
  getAll() { /* logic */ }
}
```

```go
// Lokstra Service (no controller needed!)
type UserService struct {}

func (s *UserService) GetAll(p *GetAllParams) ([]User, error) {
    // Same logic, auto-generates routes
}

// Register and auto-route
lokstra_registry.RegisterServiceType("user-service-factory", NewUserService, nil,
    deploy.WithResource("user", "users"))
```

**Migration benefits:**
- ✅ **Remove controller layer**: Direct service → HTTP mapping
- ✅ **Better performance**: Compiled Go vs runtime TypeScript
- ✅ **Flexible deployment**: One binary, multiple topologies

---

**Both frameworks are excellent for enterprise applications. Choose based on your team's expertise, performance requirements, and ecosystem preferences!**