package auth

import (
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/sirupsen/logrus"

	"game-platform/internal/platform/errors"
	"game-platform/internal/platform/ghttp"
)

// ValidationErrorResponse represents a validation error response
type ValidationErrorResponse struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// Validator defines an interface for validating request data
type Validator interface {
	Validate() error
}

// ValidationContext holds the context for validation
type ValidationContext struct {
	// The gin context
	GinContext *gin.Context

	// The validator to use
	Validator Validator

	// Whether to use custom error handling
	UseCustomErrors bool

	// The error response code to use
	ErrorResponseCode string
}

// ValidationMiddleware returns a middleware that validates request data
func ValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Process request
		c.Next()
	}
}

// ValidateJSON validates a JSON request body against a validator
func ValidateJSON(validator Validator, useCustomErrors bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := logrus.WithContext(c)

		// Bind the request body to the validator
		if err := c.ShouldBindJSON(validator); err != nil {
			logger.Warnf("JSON binding error: %v", err)

			if useCustomErrors {
				ghttp.BadRequest(c, errors.MsgInvalidRequestFormat, gin.H{
					"details": err.Error(),
				})
			} else {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Invalid JSON format: " + err.Error(),
				})
			}

			c.Abort()
			return
		}

		// Validate the request data
		if err := validator.Validate(); err != nil {
			logger.Warnf("Validation error: %v", err)

			if useCustomErrors {
				validationErrors := parseValidationErrors(err)
				ghttp.ValidationError(c, errors.MsgValidationError, gin.H{
					"errors": validationErrors,
				})
			} else {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": err.Error(),
				})
			}

			c.Abort()
			return
		}

		c.Next()
	}
}

// ValidateQuery validates query parameters against a validator
func ValidateQuery(validator Validator, useCustomErrors bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := logrus.WithContext(c)

		// Bind the query parameters to the validator
		if err := c.ShouldBindQuery(validator); err != nil {
			logger.Warnf("Query binding error: %v", err)

			if useCustomErrors {
				ghttp.BadRequest(c, "Invalid query parameters", gin.H{
					"details": err.Error(),
				})
			} else {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Invalid query parameters: " + err.Error(),
				})
			}

			c.Abort()
			return
		}

		// Validate the query parameters
		if err := validator.Validate(); err != nil {
			logger.Warnf("Validation error: %v", err)

			if useCustomErrors {
				validationErrors := parseValidationErrors(err)
				ghttp.ValidationError(c, errors.MsgValidationError, gin.H{
					"errors": validationErrors,
				})
			} else {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": err.Error(),
				})
			}

			c.Abort()
			return
		}

		c.Next()
	}
}

// ValidateParams validates path parameters against a validator
func ValidateParams(validator Validator, useCustomErrors bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := logrus.WithContext(c)

		// Bind the path parameters to the validator
		if err := c.ShouldBindUri(validator); err != nil {
			logger.Warnf("URI binding error: %v", err)

			if useCustomErrors {
				ghttp.BadRequest(c, "Invalid path parameters", gin.H{
					"details": err.Error(),
				})
			} else {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Invalid path parameters: " + err.Error(),
				})
			}

			c.Abort()
			return
		}

		// Validate the path parameters
		if err := validator.Validate(); err != nil {
			logger.Warnf("Validation error: %v", err)

			if useCustomErrors {
				validationErrors := parseValidationErrors(err)
				ghttp.ValidationError(c, errors.MsgValidationError, gin.H{
					"errors": validationErrors,
				})
			} else {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": err.Error(),
				})
			}

			c.Abort()
			return
		}

		c.Next()
	}
}

// parseValidationErrors parses validation errors into a structured format
func parseValidationErrors(err error) []ValidationErrorResponse {
	if fieldErrors, ok := err.(validation.Errors); ok {
		result := make([]ValidationErrorResponse, 0, len(fieldErrors))

		for field, fieldErr := range fieldErrors {
			var code string
			var message string

			if validationErr, ok := fieldErr.(validation.Error); ok {
				code = validationErr.Code()
				message = validationErr.Error()
			} else {
				code = "validation_error"
				message = fieldErr.Error()
			}

			result = append(result, ValidationErrorResponse{
				Field:   field,
				Message: message,
				Code:    code,
			})
		}

		return result
	}

	// If it's not a validation.Errors, return a generic error
	return []ValidationErrorResponse{
		{
			Field:   "unknown",
			Message: err.Error(),
			Code:    "validation_error",
		},
	}
}

// SanitizeString removes potentially dangerous characters from a string
func SanitizeString(s string) string {
	// Remove HTML tags
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")

	// Remove script tags
	s = strings.ReplaceAll(s, "javascript:", "")

	// Remove SQL injection characters
	s = strings.ReplaceAll(s, "'", "''")
	s = strings.ReplaceAll(s, ";", "")

	return s
}

// SanitizeStruct sanitizes all string fields in a struct
func SanitizeStruct(obj interface{}) {
	val := reflect.ValueOf(obj)

	// Check if it's a pointer and get the underlying value
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Only handle structs
	if val.Kind() != reflect.Struct {
		return
	}

	// Iterate over all fields
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)

		// Handle string fields
		if field.Kind() == reflect.String && field.CanSet() {
			sanitized := SanitizeString(field.String())
			field.SetString(sanitized)
		}

		// Recursively handle struct fields
		if (field.Kind() == reflect.Struct ||
			(field.Kind() == reflect.Ptr && field.Elem().Kind() == reflect.Struct)) && field.CanInterface() {
			SanitizeStruct(field.Interface())
		}
	}
}

// ValidationRule defines a validation rule with a code and message
type ValidationRule struct {
	Rule    validation.Rule
	Code    string
	Message string
}

// NewRule creates a new validation rule with a code and message
func NewRule(rule validation.Rule, code, message string) ValidationRule {
	return ValidationRule{
		Rule:    rule,
		Code:    code,
		Message: message,
	}
}

// Validate validates a value against the rule
func (r ValidationRule) Validate(value interface{}) error {
	if err := r.Rule.Validate(value); err != nil {
		return validation.NewError(r.Code, r.Message)
	}
	return nil
}

// ValidationRules provides common validation rules
var ValidationRules = struct {
	Required     func(message string) ValidationRule
	Email        ValidationRule
	Password     ValidationRule
	PhoneNumber  ValidationRule
	URL          ValidationRule
	Username     ValidationRule
	UUID         ValidationRule
	Length       func(min, max int, message string) ValidationRule
	Min          func(min int, message string) ValidationRule
	Max          func(max int, message string) ValidationRule
	Range        func(min, max int, message string) ValidationRule
	In           func(values []interface{}, message string) ValidationRule
	NotIn        func(values []interface{}, message string) ValidationRule
	Match        func(pattern string, message string) ValidationRule
	Alpha        ValidationRule
	Alphanumeric ValidationRule
	Numeric      ValidationRule
	Boolean      ValidationRule
}{
	Required: func(message string) ValidationRule {
		if message == "" {
			message = "This field is required"
		}
		return NewRule(validation.Required, "validation_required", message)
	},
	Email: NewRule(is.Email, "validation_email", "Must be a valid email address"),
	Password: NewRule(
		validation.By(func(value interface{}) error {
			if err := validation.Required.Validate(value); err != nil {
				return err
			}
			if err := validation.Length(8, 100).Validate(value); err != nil {
				return err
			}
			return nil
		}),
		"validation_password",
		"Password must be at least 8 characters long",
	),
	PhoneNumber: NewRule(is.E164, "validation_phone", "Must be a valid phone number"),
	URL:         NewRule(is.URL, "validation_url", "Must be a valid URL"),
	Username: NewRule(
		validation.By(func(value interface{}) error {
			if err := validation.Required.Validate(value); err != nil {
				return err
			}
			if err := validation.Length(3, 50).Validate(value); err != nil {
				return err
			}
			return nil
		}),
		"validation_username",
		"Username must be between 3 and 50 characters",
	),
	UUID: NewRule(is.UUID, "validation_uuid", "Must be a valid UUID"),
	Length: func(min, max int, message string) ValidationRule {
		if message == "" {
			message = fmt.Sprintf("Must be between %d and %d characters", min, max)
		}
		return NewRule(validation.Length(min, max), "validation_length", message)
	},
	Min: func(min int, message string) ValidationRule {
		if message == "" {
			message = fmt.Sprintf("Must be at least %d", min)
		}
		return NewRule(validation.Min(min), "validation_min", message)
	},
	Max: func(max int, message string) ValidationRule {
		if message == "" {
			message = fmt.Sprintf("Must be at most %d", max)
		}
		return NewRule(validation.Max(max), "validation_max", message)
	},
	Range: func(min, max int, message string) ValidationRule {
		if message == "" {
			message = fmt.Sprintf("Must be between %d and %d", min, max)
		}
		return NewRule(validation.By(func(value interface{}) error {
			err := validation.Min(min).Validate(value)
			if err != nil {
				return err
			}
			return validation.Max(max).Validate(value)
		}), "validation_range", message)
	},
	In: func(values []interface{}, message string) ValidationRule {
		if message == "" {
			message = "Must be one of the allowed values"
		}
		return NewRule(validation.In(values...), "validation_in", message)
	},
	NotIn: func(values []interface{}, message string) ValidationRule {
		if message == "" {
			message = "Must not be one of the disallowed values"
		}
		return NewRule(validation.NotIn(values...), "validation_not_in", message)
	},
	Match: func(pattern string, message string) ValidationRule {
		if message == "" {
			message = "Must match the required pattern"
		}
		return NewRule(validation.Match(regexp.MustCompile(pattern)), "validation_match", message)
	},
	Alpha:        NewRule(validation.Match(regexp.MustCompile(`^[a-zA-Z]+$`)), "validation_alpha", "Must contain only letters"),
	Alphanumeric: NewRule(validation.Match(regexp.MustCompile(`^[a-zA-Z0-9]+$`)), "validation_alphanumeric", "Must contain only letters and numbers"),
	Numeric:      NewRule(validation.Match(regexp.MustCompile(`^[0-9]+$`)), "validation_numeric", "Must contain only numbers"),
	Boolean:      NewRule(validation.In(true, false), "validation_boolean", "Must be a boolean value"),
}
