package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"game-platform/internal/platform/errors"
	"game-platform/internal/platform/ghttp"
	"game-platform/internal/platform/security"
)

// AuthStrategy defines the authentication strategy type
type AuthStrategy string

const (
	// Authentication strategies
	StrategyJWT         AuthStrategy = "jwt"
	StrategyBasic       AuthStrategy = "basic"
	StrategyOAuth2      AuthStrategy = "oauth2"
	StrategyAPIKey      AuthStrategy = "apikey"
	StrategyToken       AuthStrategy = "token"
	StrategyMultiFactor AuthStrategy = "mfa"
)

// AuthConfig holds configuration for the authentication middleware
type AuthConfig struct {
	Strategy       AuthStrategy
	SecretEnvName  string
	SkipPaths      []string
	RequiredScopes []string
	RequiredRoles  []string
}

// DefaultAuthConfig returns a default authentication configuration
func DefaultAuthConfig() AuthConfig {
	return AuthConfig{
		Strategy:      StrategyJWT,
		SecretEnvName: "JWT_SECRET",
	}
}

// AuthMiddleware returns a middleware that authenticates requests
func AuthMiddleware(config AuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := logrus.WithContext(c)

		// Check if path should be skipped
		path := c.Request.URL.Path
		for _, skipPath := range config.SkipPaths {
			if strings.HasPrefix(path, skipPath) {
				c.Next()
				return
			}
		}

		// Handle different authentication strategies
		switch config.Strategy {
		case StrategyJWT:
			handleJWTAuth(c, config.SecretEnvName, config.RequiredScopes, config.RequiredRoles)
		case StrategyBasic:
			handleBasicAuth(c)
		case StrategyAPIKey:
			handleAPIKeyAuth(c)
		case StrategyToken:
			handleTokenAuth(c)
		case StrategyOAuth2:
			handleOAuth2Auth(c, config.RequiredScopes)
		default:
			logger.Warnf("Unsupported authentication strategy: %s", config.Strategy)
			ghttp.InternalServerError(c, errors.MsgInternalServerError)
			return
		}
	}
}

// handleJWTAuth authenticates using JWT tokens
func handleJWTAuth(c *gin.Context, secretEnvName string, requiredScopes, requiredRoles []string) {
	logger := logrus.WithContext(c)

	token, err := security.VerifyTokenHeader(c, secretEnvName)
	if err != nil {
		logger.Warnf("JWT verification failed: %v", err)
		ghttp.Unauthorized(c, errors.MsgTokenInvalidFormat)
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		logger.Warn("Invalid JWT claims")
		ghttp.Unauthorized(c, errors.MsgTokenInvalidFormat)
		return
	}

	// Extract user ID
	userID, exists := claims["id"].(string)
	if !exists {
		logger.Warn("Missing user ID in JWT")
		ghttp.Unauthorized(c, errors.MsgTokenFieldMissing)
		return
	}

	// Set user ID in context
	c.Set("userId", userID)

	// Check for required scopes
	if len(requiredScopes) > 0 {
		if !hasRequiredScopes(claims, requiredScopes) {
			logger.Warnf("User %s lacks required scopes", userID)
			ghttp.Forbidden(c, errors.MsgInsufficientPermissions)
			return
		}
	}

	// Check for required roles
	if len(requiredRoles) > 0 {
		if !hasRequiredRoles(claims, requiredRoles) {
			logger.Warnf("User %s lacks required roles", userID)
			ghttp.Forbidden(c, errors.MsgInsufficientPermissions)
			return
		}
	}

	// Store roles and scopes in context if available
	if roles, exists := claims["roles"]; exists {
		c.Set("roles", roles)
	}
	if scopes, exists := claims["scopes"]; exists {
		c.Set("scopes", scopes)
	}

	c.Next()
}

// handleBasicAuth authenticates using Basic Authentication
func handleBasicAuth(c *gin.Context) {
	logger := logrus.WithContext(c)
	authHeader := c.GetHeader("Authorization")

	if !strings.HasPrefix(authHeader, "Basic ") {
		logger.Warn("Invalid Basic Auth format")
		ghttp.Unauthorized(c, errors.MsgInvalidAuthFormat)
		return
	}

	// Extract credentials
	encodedCredentials := strings.TrimPrefix(authHeader, "Basic ")
	decodedBytes, err := base64.StdEncoding.DecodeString(encodedCredentials)
	if err != nil {
		logger.Warnf("Failed to decode Basic Auth credentials: %v", err)
		ghttp.Unauthorized(c, errors.MsgInvalidCredentials)
		return
	}

	credentials := string(decodedBytes)
	parts := strings.SplitN(credentials, ":", 2)
	if len(parts) != 2 {
		logger.Warn("Malformed Basic Auth credentials")
		ghttp.Unauthorized(c, errors.MsgInvalidCredentials)
		return
	}

	username := parts[0]
	password := parts[1]

	// TODO: Implement actual credential verification against a secure store
	// This is a placeholder - replace with actual authentication logic
	if !verifyCredentials(username, password) {
		logger.Warnf("Invalid credentials for user: %s", username)
		ghttp.Unauthorized(c, errors.MsgInvalidCredentials)
		return
	}

	// Set user ID in context
	c.Set("userId", username)
	c.Next()
}

// handleAPIKeyAuth authenticates using API keys
func handleAPIKeyAuth(c *gin.Context) {
	logger := logrus.WithContext(c)

	// Try to get API key from header first
	apiKey := c.GetHeader(security.HeaderAPIKey)

	// If not in header, try query parameter
	if apiKey == "" {
		apiKey = c.Query("api_key")
	}

	// Get caller from header
	caller := c.GetHeader(security.HeaderCaller)

	// Check if API key and caller are provided
	if apiKey == "" || caller == "" {
		logger.Warn("Missing API key or caller")
		ghttp.Unauthorized(c, errors.MsgMissingAPIKey)
		return
	}

	// // Use the enhanced API key verification from request_validator.go
	// valid := enhancedVerifyAPIKey(apiKey, caller)
	// if !valid {
	// 	logger.WithFields(logrus.Fields{
	// 		"apiKey": apiKey,
	// 		"caller": caller,
	// 	}).Warn("Invalid API key or caller")
	// 	util.APIResponse(c, errors.INVALID_API_KEY, http.StatusUnauthorized, nil)
	// 	c.Abort()
	// 	return
	// }

	// Set client ID in context (using caller as client ID)
	c.Set("clientId", caller)
	c.Next()
}

// handleOAuth2Auth authenticates using OAuth2 tokens
func handleOAuth2Auth(c *gin.Context, requiredScopes []string) {
	logger := logrus.WithContext(c)
	authHeader := c.GetHeader("Authorization")

	if !strings.HasPrefix(authHeader, "Bearer ") {
		logger.Warn("Invalid OAuth2 token format")
		ghttp.Unauthorized(c, errors.MsgTokenInvalidFormat)
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")

	// TODO: Implement actual OAuth2 token verification against an authorization server
	// This is a placeholder - replace with actual verification logic
	tokenInfo, valid := verifyOAuth2Token(token)
	if !valid {
		logger.Warn("Invalid OAuth2 token")
		ghttp.Unauthorized(c, errors.MsgTokenInvalidFormat)
		return
	}

	// Check scopes if required
	if len(requiredScopes) > 0 {
		hasScopes := true
		for _, requiredScope := range requiredScopes {
			found := false
			for _, scope := range tokenInfo.Scopes {
				if scope == requiredScope {
					found = true
					break
				}
			}
			if !found {
				hasScopes = false
				break
			}
		}

		if !hasScopes {
			logger.Warnf("Token lacks required scopes: %v", requiredScopes)
			ghttp.Forbidden(c, errors.MsgInsufficientPermissions)
			return
		}
	}

	// Set user ID in context
	c.Set("userId", tokenInfo.UserID)
	c.Set("scopes", tokenInfo.Scopes)
	c.Next()
}

// Helper functions

// hasRequiredScopes checks if the token has all required scopes
func hasRequiredScopes(claims jwt.MapClaims, requiredScopes []string) bool {
	scopesInterface, exists := claims["scopes"]
	if !exists {
		return false
	}

	var tokenScopes []string

	// Handle different formats of scopes in JWT
	switch scopes := scopesInterface.(type) {
	case []interface{}:
		for _, scope := range scopes {
			if scopeStr, ok := scope.(string); ok {
				tokenScopes = append(tokenScopes, scopeStr)
			}
		}
	case string:
		tokenScopes = strings.Split(scopes, " ")
	default:
		return false
	}

	// Check if all required scopes are present
	for _, requiredScope := range requiredScopes {
		found := false
		for _, tokenScope := range tokenScopes {
			if tokenScope == requiredScope {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// hasRequiredRoles checks if the token has any of the required roles
func hasRequiredRoles(claims jwt.MapClaims, requiredRoles []string) bool {
	rolesInterface, exists := claims["roles"]
	if !exists {
		return false
	}

	roles, ok := rolesInterface.([]interface{})
	if !ok {
		return false
	}

	// Check if any required role is present
	for _, role := range roles {
		roleStr, ok := role.(string)
		if !ok {
			continue
		}

		for _, requiredRole := range requiredRoles {
			if roleStr == requiredRole {
				return true
			}
		}
	}

	return false
}

// handleTokenAuth authenticates using signed, time-bound tokens
func handleTokenAuth(c *gin.Context) {
	logger := logrus.WithContext(c)
	authHeader := c.GetHeader("Authorization")

	if !strings.HasPrefix(authHeader, "Bearer ") {
		logger.Warn("Invalid token format")
		ghttp.Unauthorized(c, errors.MsgTokenInvalidFormat)
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	tokenInfo, err := verifyGameToken(tokenString)
	if err != nil {
		logger.Warnf("Token verification failed: %v", err)
		ghttp.Unauthorized(c, errors.MsgTokenVerificationFailed)
		return
	}

	// Set user context from token
	c.Set("userId", tokenInfo.UID)
	c.Set("roleId", tokenInfo.RID)
	c.Set("appId", tokenInfo.AppID)
	c.Set("serverId", tokenInfo.ServerID)
	c.Next()
}

// Placeholder functions - replace with actual implementations

// verifyCredentials verifies username and password
func verifyCredentials(username, password string) bool {
	// TODO: Implement actual verification against a secure credential store
	// This is just a placeholder
	return username == "admin" && password == "password"
}

// verifyGameToken verifies a signed game token with HMAC-SHA256
func verifyGameToken(tokenString string) (*GameTokenPayload, error) {
	// Load token configuration
	config, err := loadTokenConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load token config: %w", err)
	}

	if !config.Enabled {
		return nil, fmt.Errorf("token authentication is disabled")
	}

	// Split token into payload and signature
	parts := strings.Split(tokenString, ".")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid token format: expected 2 parts, got %d", len(parts))
	}

	payloadEncoded := parts[0]
	signatureHex := parts[1]

	// Decode base64url payload
	payloadJSON, err := base64.RawURLEncoding.DecodeString(payloadEncoded)
	if err != nil {
		return nil, fmt.Errorf("failed to decode payload: %w", err)
	}

	// Parse payload JSON
	var payload GameTokenPayload
	if err := json.Unmarshal(payloadJSON, &payload); err != nil {
		return nil, fmt.Errorf("failed to parse payload JSON: %w", err)
	}

	// Validate required fields
	if err := validateRequiredFields(&payload, config.RequiredFields); err != nil {
		return nil, err
	}

	// Validate timestamp
	if err := validateTimestamp(payload.TS, config.MaxAgeSec, config.ClockSkewSec, config.AllowFutureSec); err != nil {
		return nil, err
	}

	// Get secret for app_id
	secret, err := getSecretForApp(payload.AppID, config.Secrets)
	if err != nil {
		return nil, err
	}

	// Verify HMAC-SHA256 signature
	expectedSignature := computeHMAC(payloadJSON, secret)
	if !strings.EqualFold(signatureHex, expectedSignature) {
		return nil, fmt.Errorf("invalid signature")
	}

	return &payload, nil
}

// loadTokenConfig loads token configuration from auth_config.json
func loadTokenConfig() (*TokenConfig, error) {
	configPath := os.Getenv("AUTH_CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/auth_config.json"
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var fullConfig map[string]interface{}
	if err := json.Unmarshal(data, &fullConfig); err != nil {
		return nil, err
	}

	// Determine environment
	env := os.Getenv("GO_ENV")
	if env == "" {
		env = "dev" // Default to dev environment
	}

	// Navigate to environments.<env>.token
	environments, ok := fullConfig["environments"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("environments section not found in config")
	}

	envConfig, ok := environments[env].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("environment '%s' not found in config", env)
	}

	tokenData, ok := envConfig["token"]
	if !ok {
		return nil, fmt.Errorf("token configuration not found for environment '%s'", env)
	}

	tokenJSON, err := json.Marshal(tokenData)
	if err != nil {
		return nil, err
	}

	var config TokenConfig
	if err := json.Unmarshal(tokenJSON, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// validateRequiredFields checks all required fields are present
func validateRequiredFields(payload *GameTokenPayload, required []string) error {
	fieldMap := map[string]string{
		"uid":       payload.UID,
		"rid":       payload.RID,
		"app_id":    payload.AppID,
		"server_id": payload.ServerID,
	}

	for _, field := range required {
		if val, exists := fieldMap[field]; !exists || val == "" {
			if field == "ts" && payload.TS == 0 {
				return fmt.Errorf("required field missing: %s", field)
			}
			if field != "ts" && val == "" {
				return fmt.Errorf("required field missing: %s", field)
			}
		}
	}
	return nil
}

// validateTimestamp validates token timestamp with age and skew limits
func validateTimestamp(ts int64, maxAgeSec, clockSkewSec, allowFutureSec int) error {
	now := time.Now().Unix()
	age := now - ts

	// Check if token is too old
	if age > int64(maxAgeSec+clockSkewSec) {
		return fmt.Errorf("token expired: age %ds exceeds max %ds", age, maxAgeSec)
	}

	// Check if token is too far in the future
	if ts > now+int64(allowFutureSec) {
		return fmt.Errorf("token timestamp too far in future: %ds ahead", ts-now)
	}

	// Allow slight clock skew for past tokens
	if age < 0 && age < -int64(clockSkewSec) {
		return fmt.Errorf("token timestamp in past beyond clock skew: %ds", age)
	}

	return nil
}

// getSecretForApp retrieves the active secret for a given app_id
func getSecretForApp(appID string, secrets []TokenSecret) (string, error) {
	for _, s := range secrets {
		if s.AppID == appID && s.Active {
			return s.Secret, nil
		}
	}
	return "", fmt.Errorf("no active secret found for app_id: %s", appID)
}

// computeHMAC computes HMAC-SHA256 signature
func computeHMAC(data []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

// GameTokenPayload represents the decoded token payload
type GameTokenPayload struct {
	UID      string `json:"uid"`
	RID      string `json:"rid"`
	AppID    string `json:"app_id"`
	ServerID string `json:"server_id"`
	TS       int64  `json:"ts"`
}

// TokenConfig represents the token authentication configuration
type TokenConfig struct {
	Enabled        bool          `json:"enabled"`
	MaxAgeSec      int           `json:"maxAgeSec"`
	ClockSkewSec   int           `json:"clockSkewSec"`
	AllowFutureSec int           `json:"allowFutureSec"`
	RequiredFields []string      `json:"requiredFields"`
	Secrets        []TokenSecret `json:"secrets"`
}

// TokenSecret represents an app-specific secret
type TokenSecret struct {
	AppID  string `json:"appId"`
	Secret string `json:"secret"`
	Active bool   `json:"active"`
}

// OAuth2TokenInfo contains information about an OAuth2 token
type OAuth2TokenInfo struct {
	UserID string
	Scopes []string
	Exp    int64
}

// verifyOAuth2Token verifies an OAuth2 token and returns token information
func verifyOAuth2Token(token string) (OAuth2TokenInfo, bool) {
	// TODO: Implement actual OAuth2 token verification against an authorization server
	// This is just a placeholder
	if token == "test-oauth-token" {
		return OAuth2TokenInfo{
			UserID: "user123",
			Scopes: []string{"read", "write"},
			Exp:    time.Now().Add(1 * time.Hour).Unix(),
		}, true
	}
	return OAuth2TokenInfo{}, false
}
