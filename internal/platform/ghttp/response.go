package ghttp

import (
	"encoding/json"
	"game-platform/internal/platform/errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIResponse represents the standardized API response structure
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Success sends a successful response with optional data
func Success(ctx *gin.Context, data interface{}) {
	response := APIResponse{
		Code:    errors.CodeSuccess,
		Message: errors.MsgSuccess,
		Data:    data,
	}
	ctx.JSON(http.StatusOK, response)
}

// SuccessWithMessage sends a successful response with custom message and data
func SuccessWithMessage(ctx *gin.Context, message string, data interface{}) {
	response := APIResponse{
		Code:    errors.CodeSuccess,
		Message: message,
		Data:    data,
	}
	ctx.JSON(http.StatusOK, response)
}

// Created sends a successful creation response with data
func Created(ctx *gin.Context, data interface{}) {
	response := APIResponse{
		Code:    errors.CodeSuccess,
		Message: errors.MsgSuccess,
		Data:    data,
	}
	ctx.JSON(http.StatusCreated, response)
}

// Error sends an error response with code and message
func Error(ctx *gin.Context, code int, message string, data interface{}) {
	response := APIResponse{
		Code:    code,
		Message: message,
		Data:    data,
	}
	httpStatus := errors.GetHTTPStatus(code)
	ctx.JSON(httpStatus, response)
	ctx.Abort()
}

// BadRequest sends a 400 Bad Request response
func BadRequest(ctx *gin.Context, message string, data interface{}) {
	Error(ctx, errors.CodeBadRequest, message, data)
}

// Unauthorized sends a 401 Unauthorized response
func Unauthorized(ctx *gin.Context, message string) {
	Error(ctx, errors.CodeUnauthorized, message, nil)
}

// Forbidden sends a 403 Forbidden response
func Forbidden(ctx *gin.Context, message string) {
	Error(ctx, errors.CodeForbidden, message, nil)
}

// NotFound sends a 404 Not Found response
func NotFound(ctx *gin.Context, message string) {
	Error(ctx, errors.CodeNotFound, message, nil)
}

// Conflict sends a 409 Conflict response
func Conflict(ctx *gin.Context, message string, data interface{}) {
	Error(ctx, errors.CodeConflict, message, data)
}

// ValidationError sends a 422 Validation Error response
func ValidationError(ctx *gin.Context, message string, data interface{}) {
	Error(ctx, errors.CodeValidationError, message, data)
}

// InternalServerError sends a 500 Internal Server Error response
func InternalServerError(ctx *gin.Context, message string) {
	Error(ctx, errors.CodeInternalServerError, message, nil)
}

// SendResponse sends a standardized API response for standard HTTP handlers
func SendResponse(w http.ResponseWriter, r *http.Request, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := APIResponse{
		Code:    errors.CodeSuccess,
		Message: errors.MsgSuccess,
		Data:    data,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{
			Code:    errors.CodeInternalServerError,
			Message: "Failed to encode response",
		})
	}
}

// SendErrorResponse sends a standardized error response for standard HTTP handlers
func SendErrorResponse(w http.ResponseWriter, r *http.Request, code int, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := APIResponse{
		Code:    code,
		Message: message,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{
			Code:    errors.CodeInternalServerError,
			Message: "Failed to encode response",
		})
	}
}
