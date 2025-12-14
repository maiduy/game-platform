# Coding Standards and Rules

## General Principles

### 1. DELETE MORE THAN YOU ADD
- Complexity compounds into disasters
- Favor simplicity over cleverness
- Remove dead code immediately
- Consolidate duplicate logic

### 2. Follow Existing Patterns
- Don't invent new approaches
- Study the codebase before coding
- Match the style of surrounding code
- Use established libraries

### 3. Read First, Code Second
- Read at least 1500 lines to understand context
- Understand the full scope before making changes
- Look for existing solutions before creating new ones

## Go-Specific Standards

### Package Organization
```go
// ✅ Good: Clear, focused package
package character

import (
    "context"
    "game-platform/internal/platform/errors"
)

// ❌ Bad: Mixed concerns
package stuff
```

### Error Handling
```go
// ✅ Good: Wrap errors with context
if err := repo.Save(ctx, character); err != nil {
    return nil, fmt.Errorf("failed to save character %s: %w", character.ID, err)
}

// ❌ Bad: Swallow or return raw errors
if err := repo.Save(ctx, character); err != nil {
    log.Println(err)  // Lost error
    return nil, err   // No context
}
```

### Use Platform Errors
```go
// ✅ Good: Use platform error types
import "game-platform/internal/platform/errors"

func GetCharacter(id string) (*Character, error) {
    char, err := repo.FindByID(id)
    if err != nil {
        return nil, errors.NewNotFound("CHAR_NOT_FOUND", "Character not found", err)
    }
    return char, nil
}

// ❌ Bad: Return generic errors
func GetCharacter(id string) (*Character, error) {
    char, err := repo.FindByID(id)
    if err != nil {
        return nil, errors.New("not found")
    }
    return char, nil
}
```

### Struct Design
```go
// ✅ Good: Clear field tags, validation
type Character struct {
    ID          string    `json:"id" bson:"_id"`
    Name        string    `json:"name" bson:"name" binding:"required"`
    Level       int       `json:"level" bson:"level" binding:"required,min=1,max=100"`
    CreatedAt   int64     `json:"created_at" bson:"created_at"`
}

// ❌ Bad: No tags, unclear structure
type Character struct {
    ID string
    Name string
    Level int
    CreatedAt int64
}
```

### Interface Design
```go
// ✅ Good: Small, focused interfaces
type CharacterRepository interface {
    Create(ctx context.Context, char *Character) error
    FindByID(ctx context.Context, id string) (*Character, error)
    Update(ctx context.Context, char *Character) error
    Delete(ctx context.Context, id string) error
}

// ❌ Bad: God interface with too many methods
type CharacterRepository interface {
    Create(ctx context.Context, char *Character) error
    FindByID(ctx context.Context, id string) (*Character, error)
    Update(ctx context.Context, char *Character) error
    Delete(ctx context.Context, id string) error
    FindByName(ctx context.Context, name string) (*Character, error)
    FindByLevel(ctx context.Context, level int) ([]*Character, error)
    FindByRarity(ctx context.Context, rarity string) ([]*Character, error)
    // ... 20 more methods
}
```

### Context Usage
```go
// ✅ Good: Pass context as first parameter
func (s *service) GetCharacter(ctx context.Context, id string) (*Character, error) {
    return s.repo.FindByID(ctx, id)
}

// ❌ Bad: No context
func (s *service) GetCharacter(id string) (*Character, error) {
    return s.repo.FindByID(id)
}
```

### Logging
```go
// ✅ Good: Structured logging with context
logger.WithFields(logrus.Fields{
    "character_id": char.ID,
    "user_id": userID,
    "action": "create",
}).Info("Character created successfully")

// ❌ Bad: Unstructured logs
log.Println("Character created:", char.ID, userID)
```

## Clean Architecture Rules

### Layer Dependencies
```
api (handlers) → application (services) → domain (entities)
     ↓                  ↓
repository (data access)
```

**Rules**:
- Domain NEVER imports from application/api/repository
- Application can import domain
- API can import application and domain
- Repository implements interfaces defined in application

### Domain Layer
```go
// ✅ Good: Pure domain logic, no external deps
package domain

type Character struct {
    ID    string
    Name  string
    Level int
}

func (c *Character) CanLevelUp() bool {
    return c.Level < 100
}

// ❌ Bad: Domain depends on external packages
package domain

import "github.com/gin-gonic/gin"  // ❌ NO!

type Character struct {
    ID string
    // ...
}
```

### Application Layer
```go
// ✅ Good: Define repository interface in application
package application

type CharacterRepository interface {
    Create(ctx context.Context, char *domain.Character) error
    FindByID(ctx context.Context, id string) (*domain.Character, error)
}

type CharacterService struct {
    repo CharacterRepository
}

func NewCharacterService(repo CharacterRepository) *CharacterService {
    return &CharacterService{repo: repo}
}

// ❌ Bad: Tight coupling to concrete implementation
package application

import "game-platform/internal/masterconfig/repository"

type CharacterService struct {
    repo *repository.MongoCharacterRepository  // ❌ Concrete type
}
```

### API Layer
```go
// ✅ Good: Thin handlers, delegate to service
func (h *Handler) CreateCharacter(c *gin.Context) {
    var req CreateCharacterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    char, err := h.service.CreateCharacter(c.Request.Context(), &req)
    if err != nil {
        handleError(c, err)
        return
    }

    c.JSON(200, gin.H{"data": char})
}

// ❌ Bad: Business logic in handler
func (h *Handler) CreateCharacter(c *gin.Context) {
    var req CreateCharacterRequest
    c.BindJSON(&req)

    // ❌ Validation in handler
    if req.Level < 1 || req.Level > 100 {
        c.JSON(400, gin.H{"error": "invalid level"})
        return
    }

    // ❌ Direct DB access
    h.db.Insert(&req)
    c.JSON(200, gin.H{"data": req})
}
```

## Testing Standards

### Unit Tests
```go
// ✅ Good: Test business logic with mocks
func TestCreateCharacter_Success(t *testing.T) {
    mockRepo := &MockCharacterRepository{}
    mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

    service := NewCharacterService(mockRepo)

    char, err := service.CreateCharacter(context.Background(), &CreateCharacterRequest{
        Name: "TestChar",
        Level: 1,
    })

    assert.NoError(t, err)
    assert.NotNil(t, char)
    mockRepo.AssertExpectations(t)
}

// ❌ Bad: Test with real database
func TestCreateCharacter_Success(t *testing.T) {
    db := connectToRealDB()  // ❌ Slow, fragile
    service := NewCharacterService(db)
    // ...
}
```

### Table-Driven Tests
```go
// ✅ Good: Cover multiple cases
func TestCharacter_CanLevelUp(t *testing.T) {
    tests := []struct {
        name     string
        level    int
        expected bool
    }{
        {"Level 1 can level up", 1, true},
        {"Level 99 can level up", 99, true},
        {"Level 100 cannot level up", 100, false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            char := &Character{Level: tt.level}
            assert.Equal(t, tt.expected, char.CanLevelUp())
        })
    }
}
```

## API Design Standards

### RESTful Conventions
```
GET    /api/v1/characters          # List characters
POST   /api/v1/characters          # Create character
GET    /api/v1/characters/:id      # Get character
PUT    /api/v1/characters/:id      # Update character (full)
PATCH  /api/v1/characters/:id      # Update character (partial)
DELETE /api/v1/characters/:id      # Delete character
```

### Request Validation
```go
// ✅ Good: Use struct tags for validation
type CreateCharacterRequest struct {
    Name      string   `json:"name" binding:"required,min=3,max=50"`
    Level     int      `json:"level" binding:"required,min=1,max=100"`
    Element   string   `json:"element" binding:"required,oneof=Fire Ice Water Earth"`
    SkillIDs  []string `json:"skill_ids" binding:"required,min=1"`
}

// ❌ Bad: Manual validation in handler
func (h *Handler) CreateCharacter(c *gin.Context) {
    var req CreateCharacterRequest
    c.BindJSON(&req)

    if len(req.Name) < 3 {  // ❌ Manual validation
        c.JSON(400, gin.H{"error": "name too short"})
        return
    }
    // ...
}
```

### Response Format
```go
// ✅ Good: Consistent response structure
func respondSuccess(c *gin.Context, data interface{}) {
    c.JSON(200, gin.H{
        "success": true,
        "data": data,
        "error": nil,
    })
}

func respondError(c *gin.Context, status int, code string, message string) {
    c.JSON(status, gin.H{
        "success": false,
        "data": nil,
        "error": gin.H{
            "code": code,
            "message": message,
        },
    })
}

// ❌ Bad: Inconsistent responses
func (h *Handler) GetCharacter(c *gin.Context) {
    char, err := h.service.GetCharacter(c.Param("id"))
    if err != nil {
        c.JSON(404, gin.H{"error": "not found"})  // Different format
        return
    }
    c.JSON(200, char)  // Different format
}
```

## Configuration Management

### Environment Variables
```go
// ✅ Good: Use platform utilities with defaults
import "game-platform/internal/utils"

port := util.GetEnv("HTTP_PORT", "8080")
dbURL := util.GetEnv("DATABASE_URL", "postgres://localhost/db")
logLevel := util.GetEnv("LOG_LEVEL", "info")

// ❌ Bad: Direct os.Getenv without defaults
port := os.Getenv("HTTP_PORT")  // ❌ Could be empty
```

### Configuration Structs
```go
// ✅ Good: Centralized config struct
type Config struct {
    HTTPPort    string
    GRPCPort    string
    DatabaseURL string
    RedisURL    string
    LogLevel    string
}

func LoadConfig() *Config {
    return &Config{
        HTTPPort:    util.GetEnv("HTTP_PORT", "8080"),
        GRPCPort:    util.GetEnv("GRPC_PORT", "50051"),
        DatabaseURL: util.GetEnv("DATABASE_URL", ""),
        RedisURL:    util.GetEnv("REDIS_URL", "redis://localhost:6379"),
        LogLevel:    util.GetEnv("LOG_LEVEL", "info"),
    }
}
```

## Security Best Practices

### JWT Authentication
```go
// ✅ Good: Use platform auth middleware
import "game-platform/internal/platform/auth"

router.Use(auth.AuthMiddleware())

func (h *Handler) CreateCharacter(c *gin.Context) {
    userID := auth.GetUserID(c)  // From JWT claims
    // ...
}

// ❌ Bad: Manual JWT parsing in every handler
func (h *Handler) CreateCharacter(c *gin.Context) {
    tokenString := c.GetHeader("Authorization")
    // ❌ Parse JWT manually every time
}
```

### Input Sanitization
```go
// ✅ Good: Sanitize user input
import "html"

func (s *service) CreateCharacter(req *CreateCharacterRequest) (*Character, error) {
    char := &Character{
        Name: html.EscapeString(req.Name),  // Prevent XSS
        // ...
    }
    return char, nil
}
```

## Performance Best Practices

### Database Queries
```go
// ✅ Good: Use indexes, limit results
func (r *repository) ListCharacters(ctx context.Context, limit int) ([]*Character, error) {
    var characters []*Character
    err := r.db.WithContext(ctx).
        Limit(limit).
        Order("created_at DESC").
        Find(&characters).Error
    return characters, err
}

// ❌ Bad: No limits, no indexes
func (r *repository) ListCharacters(ctx context.Context) ([]*Character, error) {
    var characters []*Character
    r.db.Find(&characters)  // ❌ Could fetch millions
    return characters, nil
}
```

### Caching
```go
// ✅ Good: Cache frequently accessed data
func (s *service) GetCharacter(ctx context.Context, id string) (*Character, error) {
    // Check cache first
    if cached, err := s.cache.Get(ctx, "character:"+id); err == nil {
        return cached, nil
    }

    // Fetch from DB
    char, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, err
    }

    // Cache for future requests
    s.cache.Set(ctx, "character:"+id, char, 5*time.Minute)
    return char, nil
}
```

### Goroutines
```go
// ✅ Good: Use context for cancellation
func (s *service) ProcessBatch(ctx context.Context, ids []string) error {
    var wg sync.WaitGroup
    errChan := make(chan error, len(ids))

    for _, id := range ids {
        wg.Add(1)
        go func(id string) {
            defer wg.Done()
            select {
            case <-ctx.Done():
                return
            default:
                if err := s.processOne(ctx, id); err != nil {
                    errChan <- err
                }
            }
        }(id)
    }

    wg.Wait()
    close(errChan)

    for err := range errChan {
        if err != nil {
            return err
        }
    }
    return nil
}

// ❌ Bad: No cancellation, goroutine leak
func (s *service) ProcessBatch(ids []string) error {
    for _, id := range ids {
        go func(id string) {
            s.processOne(id)  // ❌ Could run forever
        }(id)
    }
    return nil
}
```

## Git Commit Standards

### Commit Messages
```
Format: <type>(<scope>): <subject>

Types:
- feat: New feature
- fix: Bug fix
- refactor: Code refactoring
- docs: Documentation changes
- test: Test changes
- chore: Build/tooling changes

Examples:
✅ feat(masterconfig): add character CRUD endpoints
✅ fix(leaderboard): correct ranking calculation for ties
✅ refactor(platform): consolidate error handling
❌ updated stuff
❌ fix bug
```

### Commit Frequency
- Commit every 5-10 minutes for meaningful progress
- One logical change per commit
- Don't commit broken code

## Summary

1. **Delete more than you add** - Simplicity wins
2. **Follow existing patterns** - Don't reinvent
3. **Read first, code second** - Understand before changing
4. **Use platform utilities** - Don't duplicate
5. **Test your code** - Unit tests required
6. **Log properly** - Structured logs only
7. **Handle errors** - Wrap with context
8. **Commit frequently** - Small, logical changes
