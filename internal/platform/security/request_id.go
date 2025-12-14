package security

import (
	"context"

	"github.com/gin-gonic/gin"
)

const (
	// ContextKeyRequestID is the key used to store the request ID in the context
	ContextKeyRequestID = "request_id"
	// ContextKeyCaller is the key used to store the caller in the context
	ContextKeyCaller = "caller"
)

// GetRequestID extracts the request ID from a gin context
func GetRequestID(c *gin.Context) string {
	// Try to get from header first
	requestID := c.GetHeader(HeaderRequestID)
	c.Set(ContextKeyRequestID, requestID)
	return requestID
}

// GetCaller extracts the caller from a gin context
func GetCaller(c *gin.Context) string {
	// Try to get from header first
	caller := c.GetHeader(HeaderCaller)
	c.Set(ContextKeyCaller, caller)
	return caller
}

// SetRequestIDToContext sets the request ID to a standard Go context
func SetRequestIDToContext(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, ContextKeyRequestID, requestID)
}

// GetRequestIDFromContext gets the request ID from a standard Go context
func GetRequestIDFromContext(ctx context.Context) string {
	if requestID, ok := ctx.Value(ContextKeyRequestID).(string); ok {
		return requestID
	}
	return ""
}

// SetCallerToContext sets the caller to a standard Go context
func SetCallerToContext(ctx context.Context, caller string) context.Context {
	return context.WithValue(ctx, ContextKeyCaller, caller)
}

// GetCallerFromContext gets the caller from a standard Go context
func GetCallerFromContext(ctx context.Context) string {
	if caller, ok := ctx.Value(ContextKeyCaller).(string); ok {
		return caller
	}
	return ""
}

// SetRequestInfoToContext sets both request ID and caller to a standard Go context
func SetRequestInfoToContext(ctx context.Context, requestID, caller string) context.Context {
	ctx = SetRequestIDToContext(ctx, requestID)
	return SetCallerToContext(ctx, caller)
}
