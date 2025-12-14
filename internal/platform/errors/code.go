package errors

import "net/http"

// Standard HTTP-style response codes
const (
	CodeSuccess              = 0   // Request processed successfully
	CodeBadRequest           = 400 // Request body or parameters are invalid
	CodeUnauthorized         = 401 // Token is missing, invalid, or expired
	CodeForbidden            = 403 // Client does not have permission
	CodeNotFound             = 404 // Requested resource does not exist
	CodeConflict             = 409 // Request conflicts with current resource state
	CodeValidationError      = 422 // Input data fails business or validation rules
	CodeInternalServerError  = 500 // Unexpected server-side error
)

// Standard response messages
const (
	MsgSuccess              = "Success"
	MsgBadRequest           = "Bad Request"
	MsgUnauthorized         = "Unauthorized"
	MsgForbidden            = "Forbidden"
	MsgNotFound             = "Not Found"
	MsgConflict             = "Conflict"
	MsgValidationError      = "Validation Error"
	MsgInternalServerError  = "Internal Server Error"
)

// Token authentication error messages
const (
	MsgTokenMissing           = "Token is missing"
	MsgTokenInvalidFormat     = "Invalid token format"
	MsgTokenInvalidSignature  = "Invalid token signature"
	MsgTokenExpired           = "Token has expired"
	MsgTokenFieldMissing      = "Required token field is missing"
	MsgTokenAppInvalid        = "Invalid or inactive app_id"
	MsgTokenDisabled          = "Token authentication is disabled"
	MsgTokenVerificationFailed = "Token verification failed"
)

// Request validation error messages
const (
	MsgMissingRequiredHeader = "Missing required header"
	MsgInvalidRequestBody    = "Invalid request body"
	MsgInvalidRequestFormat  = "Invalid request format"
	MsgInvalidSignature      = "Invalid signature"
)

// Authentication error messages
const (
	MsgInvalidAPIKey          = "Invalid API key"
	MsgMissingAPIKey          = "Missing API key"
	MsgInvalidCredentials     = "Invalid credentials"
	MsgInvalidAuthFormat      = "Invalid authentication format"
	MsgInsufficientPermissions = "Insufficient permissions"
)

// GetHTTPStatus returns the appropriate HTTP status code for an error code
func GetHTTPStatus(code int) int {
	switch code {
	case CodeSuccess:
		return http.StatusOK
	case CodeBadRequest:
		return http.StatusBadRequest
	case CodeUnauthorized:
		return http.StatusUnauthorized
	case CodeForbidden:
		return http.StatusForbidden
	case CodeNotFound:
		return http.StatusNotFound
	case CodeConflict:
		return http.StatusConflict
	case CodeValidationError:
		return http.StatusUnprocessableEntity
	case CodeInternalServerError:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
