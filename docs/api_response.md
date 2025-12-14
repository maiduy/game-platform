# API Response Format Standard

## Overview

All API responses across the game platform follow a consistent JSON structure to ensure predictability and ease of integration for client developers.

## Standard Response Structure

###Success Response
```json
{
  "success": true,
  "data": { ... },
  "error": null,
  "meta": {
    "request_id": "550e8400-e29b-41d4-a716-446655440000",
    "timestamp": 1703001600,
    "version": "v1"
  }
}
```

### Error Response
```json
{
  "success": false,
  "data": null,
  "error": {
    "code": "CHAR_NOT_FOUND",
    "message": "Character with ID 'char_123' not found",
    "details": {
      "character_id": "char_123"
    }
  },
  "meta": {
    "request_id": "550e8400-e29b-41d4-a716-446655440000",
    "timestamp": 1703001600,
    "version": "v1"
  }
}
```

## Response Fields

### Top-Level Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `success` | boolean | Yes | Indicates if the request was successful |
| `data` | object/array/null | Yes | Response payload (null on error) |
| `error` | object/null | Yes | Error information (null on success) |
| `meta` | object | No | Metadata about the response |

### Meta Object

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `request_id` | string | Yes | Unique identifier for request tracking |
| `timestamp` | integer | Yes | Unix timestamp of response |
| `version` | string | Yes | API version (e.g., "v1") |
| `pagination` | object | No | Pagination info (for list endpoints) |

### Error Object

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `code` | string | Yes | Machine-readable error code |
| `message` | string | Yes | Human-readable error message |
| `details` | object | No | Additional context about the error |

### Pagination Object (for list endpoints)

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `page` | integer | Yes | Current page number (1-indexed) |
| `page_size` | integer | Yes | Number of items per page |
| `total_count` | integer | Yes | Total number of items |
| `total_pages` | integer | Yes | Total number of pages |

## HTTP Status Codes

### Success Codes

| Code | Status | Usage |
|------|--------|-------|
| 200 | OK | Successful GET, PUT, PATCH, DELETE |
| 201 | Created | Successful POST (resource created) |
| 204 | No Content | Successful DELETE with no response body |

### Client Error Codes

| Code | Status | Usage |
|------|--------|-------|
| 400 | Bad Request | Invalid request format or parameters |
| 401 | Unauthorized | Missing or invalid authentication |
| 403 | Forbidden | Authenticated but lacks permissions |
| 404 | Not Found | Resource does not exist |
| 409 | Conflict | Resource already exists or conflict |
| 422 | Unprocessable Entity | Validation error |
| 429 | Too Many Requests | Rate limit exceeded |

### Server Error Codes

| Code | Status | Usage |
|------|--------|-------|
| 500 | Internal Server Error | Unexpected server error |
| 502 | Bad Gateway | Upstream service error |
| 503 | Service Unavailable | Service temporarily unavailable |
| 504 | Gateway Timeout | Upstream service timeout |

## Error Codes

### General Errors

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `BAD_REQUEST` | 400 | Invalid request format |
| `UNAUTHORIZED` | 401 | Authentication required |
| `FORBIDDEN` | 403 | Insufficient permissions |
| `NOT_FOUND` | 404 | Resource not found |
| `CONFLICT` | 409 | Resource conflict |
| `VALIDATION_ERROR` | 422 | Request validation failed |
| `RATE_LIMIT_EXCEEDED` | 429 | Too many requests |
| `INTERNAL_ERROR` | 500 | Internal server error |
| `SERVICE_UNAVAILABLE` | 503 | Service temporarily unavailable |

### MasterConfig Service Errors

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `CHAR_NOT_FOUND` | 404 | Character not found |
| `CHAR_ALREADY_EXISTS` | 409 | Character with ID already exists |
| `INVALID_RARITY` | 422 | Invalid rarity value |
| `INVALID_ELEMENT` | 422 | Invalid element value |
| `SKILL_NOT_FOUND` | 404 | Skill not found |
| `ITEM_NOT_FOUND` | 404 | Item not found |

### Leaderboard Service Errors

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `LEADERBOARD_NOT_FOUND` | 404 | Leaderboard not found |
| `INVALID_SCORE` | 422 | Invalid score value |
| `PLAYER_NOT_FOUND` | 404 | Player not found in leaderboard |

### Economy Service Errors

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `INSUFFICIENT_CURRENCY` | 422 | Not enough currency |
| `ITEM_NOT_OWNED` | 404 | Item not in inventory |
| `TRANSACTION_FAILED` | 500 | Transaction processing failed |
| `IAP_VERIFICATION_FAILED` | 422 | In-app purchase verification failed |

### Match Service Errors

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `MATCH_NOT_FOUND` | 404 | Match not found |
| `QUEUE_FULL` | 503 | Matchmaking queue is full |
| `ALREADY_IN_MATCH` | 409 | Player already in a match |

## Examples

### 1. Create Character (Success)

**Request**:
```http
POST /api/v1/characters
Content-Type: application/json
Authorization: Bearer <token>

{
  "character_id": "char_001",
  "meta": {
    "nameKey": "character.warrior.name",
    "rarity": "SSR",
    "element": "Fire"
  },
  "base_stats": {
    "level": 1,
    "hp": 1000,
    "atk": 150
  }
}
```

**Response** (201 Created):
```json
{
  "success": true,
  "data": {
    "character_id": "char_001",
    "meta": {
      "nameKey": "character.warrior.name",
      "rarity": "SSR",
      "element": "Fire"
    },
    "base_stats": {
      "level": 1,
      "hp": 1000,
      "atk": 150
    },
    "created_at": 1703001600,
    "updated_at": 1703001600
  },
  "error": null,
  "meta": {
    "request_id": "550e8400-e29b-41d4-a716-446655440000",
    "timestamp": 1703001600,
    "version": "v1"
  }
}
```

### 2. Get Character (Not Found)

**Request**:
```http
GET /api/v1/characters/char_999
Authorization: Bearer <token>
```

**Response** (404 Not Found):
```json
{
  "success": false,
  "data": null,
  "error": {
    "code": "CHAR_NOT_FOUND",
    "message": "Character with ID 'char_999' not found",
    "details": {
      "character_id": "char_999"
    }
  },
  "meta": {
    "request_id": "660e8400-e29b-41d4-a716-446655440000",
    "timestamp": 1703001650,
    "version": "v1"
  }
}
```

### 3. List Characters (Paginated)

**Request**:
```http
GET /api/v1/characters?page=1&page_size=20&rarity=SSR
Authorization: Bearer <token>
```

**Response** (200 OK):
```json
{
  "success": true,
  "data": {
    "characters": [
      {
        "character_id": "char_001",
        "meta": {
          "nameKey": "character.warrior.name",
          "rarity": "SSR",
          "element": "Fire"
        }
      },
      {
        "character_id": "char_002",
        "meta": {
          "nameKey": "character.mage.name",
          "rarity": "SSR",
          "element": "Ice"
        }
      }
    ]
  },
  "error": null,
  "meta": {
    "request_id": "770e8400-e29b-41d4-a716-446655440000",
    "timestamp": 1703001700,
    "version": "v1",
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total_count": 42,
      "total_pages": 3
    }
  }
}
```

### 4. Validation Error

**Request**:
```http
POST /api/v1/characters
Content-Type: application/json
Authorization: Bearer <token>

{
  "character_id": "char_001",
  "meta": {
    "nameKey": "",
    "rarity": "INVALID",
    "element": "Fire"
  }
}
```

**Response** (422 Unprocessable Entity):
```json
{
  "success": false,
  "data": null,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Request validation failed",
    "details": {
      "fields": [
        {
          "field": "meta.nameKey",
          "message": "Field is required"
        },
        {
          "field": "meta.rarity",
          "message": "Must be one of: R, SR, SSR"
        }
      ]
    }
  },
  "meta": {
    "request_id": "880e8400-e29b-41d4-a716-446655440000",
    "timestamp": 1703001750,
    "version": "v1"
  }
}
```

### 5. Rate Limit Exceeded

**Request**:
```http
GET /api/v1/characters/char_001
Authorization: Bearer <token>
```

**Response** (429 Too Many Requests):
```json
{
  "success": false,
  "data": null,
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Rate limit exceeded. Please try again later.",
    "details": {
      "limit": 100,
      "window": "60s",
      "retry_after": 45
    }
  },
  "meta": {
    "request_id": "990e8400-e29b-41d4-a716-446655440000",
    "timestamp": 1703001800,
    "version": "v1"
  }
}
```

### 6. Internal Server Error

**Request**:
```http
POST /api/v1/leaderboards/lb_001/scores
Content-Type: application/json
Authorization: Bearer <token>

{
  "player_id": "player_123",
  "score": 9999
}
```

**Response** (500 Internal Server Error):
```json
{
  "success": false,
  "data": null,
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "An unexpected error occurred. Please try again later.",
    "details": {
      "request_id": "aa0e8400-e29b-41d4-a716-446655440000"
    }
  },
  "meta": {
    "request_id": "aa0e8400-e29b-41d4-a716-446655440000",
    "timestamp": 1703001850,
    "version": "v1"
  }
}
```

## Implementation Guidelines

### Go Implementation

```go
// Standard response helpers
package response

import (
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "time"
)

type Response struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data"`
    Error   *ErrorInfo  `json:"error"`
    Meta    *Meta       `json:"meta"`
}

type ErrorInfo struct {
    Code    string      `json:"code"`
    Message string      `json:"message"`
    Details interface{} `json:"details,omitempty"`
}

type Meta struct {
    RequestID  string      `json:"request_id"`
    Timestamp  int64       `json:"timestamp"`
    Version    string      `json:"version"`
    Pagination *Pagination `json:"pagination,omitempty"`
}

type Pagination struct {
    Page       int   `json:"page"`
    PageSize   int   `json:"page_size"`
    TotalCount int64 `json:"total_count"`
    TotalPages int   `json:"total_pages"`
}

func Success(c *gin.Context, data interface{}) {
    c.JSON(200, Response{
        Success: true,
        Data:    data,
        Error:   nil,
        Meta: &Meta{
            RequestID: uuid.New().String(),
            Timestamp: time.Now().Unix(),
            Version:   "v1",
        },
    })
}

func Error(c *gin.Context, status int, code string, message string, details interface{}) {
    c.JSON(status, Response{
        Success: false,
        Data:    nil,
        Error: &ErrorInfo{
            Code:    code,
            Message: message,
            Details: details,
        },
        Meta: &Meta{
            RequestID: uuid.New().String(),
            Timestamp: time.Now().Unix(),
            Version:   "v1",
        },
    })
}

func SuccessWithPagination(c *gin.Context, data interface{}, page, pageSize int, totalCount int64) {
    totalPages := int(totalCount / int64(pageSize))
    if totalCount%int64(pageSize) != 0 {
        totalPages++
    }

    c.JSON(200, Response{
        Success: true,
        Data:    data,
        Error:   nil,
        Meta: &Meta{
            RequestID: uuid.New().String(),
            Timestamp: time.Now().Unix(),
            Version:   "v1",
            Pagination: &Pagination{
                Page:       page,
                PageSize:   pageSize,
                TotalCount: totalCount,
                TotalPages: totalPages,
            },
        },
    })
}
```

### Usage in Handlers

```go
package handler

import (
    "game-platform/internal/platform/http/response"
    "github.com/gin-gonic/gin"
)

func (h *Handler) GetCharacter(c *gin.Context) {
    id := c.Param("id")

    char, err := h.service.GetCharacter(c.Request.Context(), id)
    if err != nil {
        if errors.Is(err, errors.ErrNotFound) {
            response.Error(c, 404, "CHAR_NOT_FOUND", "Character not found", gin.H{
                "character_id": id,
            })
            return
        }
        response.Error(c, 500, "INTERNAL_ERROR", "Internal server error", nil)
        return
    }

    response.Success(c, char)
}

func (h *Handler) ListCharacters(c *gin.Context) {
    page := c.GetInt("page")
    pageSize := c.GetInt("page_size")

    chars, totalCount, err := h.service.ListCharacters(c.Request.Context(), page, pageSize)
    if err != nil {
        response.Error(c, 500, "INTERNAL_ERROR", "Internal server error", nil)
        return
    }

    response.SuccessWithPagination(c, gin.H{
        "characters": chars,
    }, page, pageSize, totalCount)
}
```

## Best Practices

1. **Always include meta**: Include request_id for debugging and tracking
2. **Consistent error codes**: Use predefined error codes, don't make up new ones
3. **Meaningful messages**: Error messages should be actionable for clients
4. **Use details sparingly**: Only include details that help debugging
5. **Pagination**: Always paginate list endpoints
6. **HTTP Status**: Match HTTP status codes to success/error states
7. **Null vs Empty**: Use `null` for missing data, `[]` for empty arrays, `{}` for empty objects
