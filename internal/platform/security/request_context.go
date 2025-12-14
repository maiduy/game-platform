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
	// ContextKeyUserID is the key used to store the user ID in the context
	ContextKeyUserID = "userId"
	// ContextKeyRoleID is the key used to store the role ID in the context
	ContextKeyRoleID = "roleId"
	// ContextKeyAppID is the key used to store the application ID in the context
	ContextKeyAppID = "appId"
	// ContextKeyServerID is the key used to store the server ID in the context
	ContextKeyServerID = "serverId"
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

// GetUserID extracts the user ID from a gin context
func GetUserID(c *gin.Context) string {
	if userID, exists := c.Get(ContextKeyUserID); exists {
		if uid, ok := userID.(string); ok {
			return uid
		}
	}
	return ""
}

// GetRoleID extracts the role ID from a gin context
func GetRoleID(c *gin.Context) string {
	if roleID, exists := c.Get(ContextKeyRoleID); exists {
		if rid, ok := roleID.(string); ok {
			return rid
		}
	}
	return ""
}

// GetAppID extracts the application ID from a gin context
func GetAppID(c *gin.Context) string {
	if appID, exists := c.Get(ContextKeyAppID); exists {
		if aid, ok := appID.(string); ok {
			return aid
		}
	}
	return ""
}

// GetServerID extracts the server ID from a gin context
func GetServerID(c *gin.Context) string {
	if serverID, exists := c.Get(ContextKeyServerID); exists {
		if sid, ok := serverID.(string); ok {
			return sid
		}
	}
	return ""
}

// SetUserIDToContext sets the user ID to a standard Go context
func SetUserIDToContext(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, ContextKeyUserID, userID)
}

// GetUserIDFromContext gets the user ID from a standard Go context
func GetUserIDFromContext(ctx context.Context) string {
	if userID, ok := ctx.Value(ContextKeyUserID).(string); ok {
		return userID
	}
	return ""
}

// SetRoleIDToContext sets the role ID to a standard Go context
func SetRoleIDToContext(ctx context.Context, roleID string) context.Context {
	return context.WithValue(ctx, ContextKeyRoleID, roleID)
}

// GetRoleIDFromContext gets the role ID from a standard Go context
func GetRoleIDFromContext(ctx context.Context) string {
	if roleID, ok := ctx.Value(ContextKeyRoleID).(string); ok {
		return roleID
	}
	return ""
}

// SetAppIDToContext sets the application ID to a standard Go context
func SetAppIDToContext(ctx context.Context, appID string) context.Context {
	return context.WithValue(ctx, ContextKeyAppID, appID)
}

// GetAppIDFromContext gets the application ID from a standard Go context
func GetAppIDFromContext(ctx context.Context) string {
	if appID, ok := ctx.Value(ContextKeyAppID).(string); ok {
		return appID
	}
	return ""
}

// SetServerIDToContext sets the server ID to a standard Go context
func SetServerIDToContext(ctx context.Context, serverID string) context.Context {
	return context.WithValue(ctx, ContextKeyServerID, serverID)
}

// GetServerIDFromContext gets the server ID from a standard Go context
func GetServerIDFromContext(ctx context.Context) string {
	if serverID, ok := ctx.Value(ContextKeyServerID).(string); ok {
		return serverID
	}
	return ""
}

// SetAuthInfoToContext sets all authentication-related fields to a standard Go context
func SetAuthInfoToContext(ctx context.Context, userID, roleID, appID, serverID string) context.Context {
	ctx = SetUserIDToContext(ctx, userID)
	ctx = SetRoleIDToContext(ctx, roleID)
	ctx = SetAppIDToContext(ctx, appID)
	return SetServerIDToContext(ctx, serverID)
}

// SetAllContextInfo sets all context fields (request info + auth info) to a standard Go context
func SetAllContextInfo(ctx context.Context, requestID, caller, userID, roleID, appID, serverID string) context.Context {
	ctx = SetRequestInfoToContext(ctx, requestID, caller)
	return SetAuthInfoToContext(ctx, userID, roleID, appID, serverID)
}
