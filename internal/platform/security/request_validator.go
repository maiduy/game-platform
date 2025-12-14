package security

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"game-platform/internal/platform/errors"

	http "game-platform/internal/platform/ghttp"
)

const (
	// Request header constants
	HeaderRequestID = "X-Request-ID"
	HeaderCaller    = "X-Caller"
	HeaderSignature = "X-Signature"
	HeaderAPIKey    = "X-API-Key"
)

// RequestValidatorConfig holds configuration for the request validation middleware
type RequestValidatorConfig struct {
	// Whether to skip validation for certain paths
	SkipPaths []string

	// Whether to validate the signature
	ValidateSignature bool

	// Whether to validate the API key without signature validation
	ValidateKey bool

	// Whether to require base64 encoding for the request body
	RequireBase64Encoding bool
}

// DefaultRequestValidatorConfig returns a default request validator configuration
func DefaultRequestValidatorConfig() RequestValidatorConfig {
	return RequestValidatorConfig{
		ValidateSignature:     true,
		ValidateKey:           false,
		RequireBase64Encoding: true,
	}
}

// APIKeyConfig represents the API key configuration structure
type APIKeyConfig struct {
	Environments map[string]Environment `json:"environments"`
}

// Environment represents environment-specific API key configuration
type Environment struct {
	APIKeys []APIKey `json:"apiKeys"`
}

// APIKey represents a single API key configuration
type APIKey struct {
	Caller string `json:"caller"`
	Key    string `json:"key"`
	Active bool   `json:"active"`
}

// loadAuthConfig loads the API key configuration from file
func loadAuthConfig() (*APIKeyConfig, error) {
	// Determine the config file path
	configPath := os.Getenv("AUTH_CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/auth_config.json"
	}

	// Read the config file
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	// Parse the config
	var config APIKeyConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// enhancedVerifyAPIKey verifies the API key against the configuration
func enhancedVerifyAPIKey(apiKey, caller string) bool {
	// Load API key configuration
	config, err := loadAuthConfig()
	if err != nil {
		logrus.WithError(err).Error("Failed to load API key configuration")
		return false
	}

	// Determine the environment
	env := os.Getenv("GO_ENV")
	if env == "" {
		env = "dev" // Default to dev environment
	}

	// Get environment-specific API keys
	envConfig, exists := config.Environments[env]
	if !exists {
		logrus.WithField("environment", env).Warn("Environment not found in API key configuration")
		return false
	}

	// Check if the API key and caller match any configured key
	for _, key := range envConfig.APIKeys {
		if key.Key == apiKey && key.Caller == caller && key.Active {
			return true
		}
	}

	return false
}

// RequestValidatorMiddleware returns a middleware that validates incoming requests
func RequestValidatorMiddleware(config RequestValidatorConfig) gin.HandlerFunc {
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

		// Extract required headers
		requestID := c.GetHeader(HeaderRequestID)
		caller := c.GetHeader(HeaderCaller)
		signature := c.GetHeader(HeaderSignature)
		apiKey := c.GetHeader(HeaderAPIKey)

		// Validate required headers
		if requestID == "" {
			logger.Warn("Missing request ID header")
			http.BadRequest(c, errors.MsgMissingRequiredHeader, nil)
			return
		}

		if caller == "" {
			logger.Warn("Missing caller header")
			http.BadRequest(c, errors.MsgMissingRequiredHeader, nil)
			return
		}

		// Validate API key if required
		if config.ValidateKey {
			if apiKey == "" || caller == "" {
				logger.Warn("Missing API key or caller header")
				http.Unauthorized(c, errors.MsgMissingAPIKey)
				return
			}

			if !enhancedVerifyAPIKey(apiKey, caller) {
				logger.WithFields(logrus.Fields{
					"apiKey": apiKey,
					"caller": caller,
				}).Warn("Invalid API key or caller")
				http.Unauthorized(c, errors.MsgInvalidAPIKey)
				return
			}
		}

		// Additional signature validation if required
		if config.ValidateSignature {
			if signature == "" {
				logger.Warn("Missing signature header")
				http.BadRequest(c, errors.MsgMissingRequiredHeader, nil)
				return
			}
		}

		// Read and store the request body
		var bodyBytes []byte
		var err error

		if c.Request.Body != nil {
			bodyBytes, err = c.GetRawData()
			if err != nil {
				logger.Warnf("Failed to read request body: %v", err)
				http.BadRequest(c, errors.MsgInvalidRequestBody, nil)
				return
			}
		}

		// Store the original body for later use
		c.Set("originalBody", bodyBytes)

		// If body is empty, skip further validation
		if len(bodyBytes) == 0 {
			c.Next()
			return
		}

		// Handle base64 encoded body
		var decodedBody []byte
		dataBodyInBase64 := string(bodyBytes)

		if config.RequireBase64Encoding {
			// Verify that the body is valid base64
			decodedBody, err = base64.StdEncoding.DecodeString(dataBodyInBase64)
			if err != nil {
				logger.Warnf("Invalid base64 encoding in request body: %v", err)
				http.BadRequest(c, errors.MsgInvalidRequestFormat, nil)
				return
			}
		} else {
			decodedBody = bodyBytes
		}

		// Store the decoded body for later use
		c.Set("decodedBody", decodedBody)

		// Validate signature if required
		if config.ValidateSignature {
			// Compute expected signature
			expectedSignature := computeSignature(requestID, caller, dataBodyInBase64, apiKey)

			// Compare with provided signature
			if !strings.EqualFold(signature, expectedSignature) {
				logger.Warn("Invalid signature")
				http.Unauthorized(c, errors.MsgInvalidSignature)
				return
			}
		}

		c.Next()
	}
}

// computeSignature generates a SHA-256 hash using the specified format
func computeSignature(requestID, caller, dataBodyInBase64, key string) string {
	data := fmt.Sprintf("%s|%s|%s|%s", requestID, caller, dataBodyInBase64, key)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// GetDecodedBody retrieves the decoded body from the context
func GetDecodedBody(c *gin.Context) ([]byte, bool) {
	if decodedBody, exists := c.Get("decodedBody"); exists {
		if body, ok := decodedBody.([]byte); ok {
			return body, true
		}
	}
	return nil, false
}

// GetOriginalBody retrieves the original raw body from the context
func GetOriginalBody(c *gin.Context) ([]byte, bool) {
	if originalBody, exists := c.Get("originalBody"); exists {
		if body, ok := originalBody.([]byte); ok {
			return body, true
		}
	}
	return nil, false
}
