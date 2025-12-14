package accesslog

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"game-platform/internal/platform/security"
)

// Config holds configuration for the access log middleware
type Config struct {
	Module              string // e.g., "game"
	Version             string // e.g., "v1"
	IncludeRequestBody  bool   // Whether to include request body in logs
	IncludeResponseBody bool   // Whether to include response body in logs
	MaxBodySize         int    // Maximum body size to log (bytes), 0 = unlimited
}

// DefaultConfig returns the default configuration
func DefaultConfig(module string) *Config {
	return &Config{
		Module:              module,
		Version:             GetVersionFromEnv(),
		IncludeRequestBody:  true,
		IncludeResponseBody: true,
		MaxBodySize:         GetMaxBodySizeFromEnv(),
	}
}

// ResponseRecorder wraps gin.ResponseWriter to capture response body and status
type ResponseRecorder struct {
	gin.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

func (r *ResponseRecorder) Write(b []byte) (int, error) {
	r.body.Write(b) // capture body
	return r.ResponseWriter.Write(b)
}

func (r *ResponseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *ResponseRecorder) Status() int {
	if r.statusCode == 0 {
		return 200 // Default status if not explicitly set
	}
	return r.statusCode
}

// decompressGzip decompresses gzip-encoded data
func decompressGzip(data []byte) (string, error) {
	// Check if data is actually gzip-compressed
	if len(data) < 2 || data[0] != 0x1f || data[1] != 0x8b {
		return "", io.ErrUnexpectedEOF
	}

	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	defer reader.Close()

	var result bytes.Buffer
	_, err = io.Copy(&result, reader)
	if err != nil {
		return "", err
	}
	return result.String(), nil
}

// getLocalIP retrieves the local IP address of the machine
func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "unknown"
	}

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String()
			}
		}
	}

	return "unknown"
}

// extractResponseCode extracts the response code from the response body
// Attempts to parse JSON and extract "code" or "respCode" field (including nested in "data")
func extractResponseCode(responseBody string) string {
	if responseBody == "" {
		return "OK"
	}

	// Try to parse as JSON
	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(responseBody), &jsonData); err == nil {
		// Check for common response code fields
		if code, ok := jsonData["code"]; ok {
			return toString(code)
		}
		// Check for nested respCode in data field
		if data, ok := jsonData["data"]; ok {
			if dataMap, ok := data.(map[string]interface{}); ok {
				if respCode, ok := dataMap["respCode"]; ok {
					return toString(respCode)
				}
			}
		}
	}

	return "OK"
}

// toString converts interface{} to string
func toString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case int:
		return fmt.Sprintf("%d", val)
	case int64:
		return fmt.Sprintf("%d", val)
	case float64:
		return fmt.Sprintf("%.0f", val)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// LoggerMiddleware creates a middleware for logging API requests with structured format
// Deprecated: Use StructuredLoggerMiddleware instead
func LoggerMiddleware(logger *logrus.Logger) gin.HandlerFunc {
	config := DefaultConfig("ds.HubMicroService")
	return StructuredLoggerMiddleware(logger, config)
}

// StructuredLoggerMiddleware creates an enhanced middleware for structured logging
// with consistent JSON format including module, operation, version, requestId, etc.
//
// Example log output:
//
//	{
//	  "module": "ds.HubMicroService",
//	  "operation": "/sme-segment",
//	  "opVer": "v1",
//	  "requestId": "886757577vde07kh",
//	  "localIp": "192.168.105.231",
//	  "httpCode": 200,
//	  "respCode": "OK",
//	  "startTime": 1765162304324,
//	  "endTime": 1765162304431,
//	  "duration_ms": 107,
//	  "level": "info",
//	  "message": "HTTP Request",
//	  "data": {
//	    "method": "POST",
//	    "path": "/sme-segment",
//	    "clientIp": "10.0.0.15",
//	    "userAgent": "...",
//	    "request": {...},
//	    "response": {...}
//	  }
//	}
func StructuredLoggerMiddleware(logger *logrus.Logger, config *Config) gin.HandlerFunc {
	if config == nil {
		config = DefaultConfig("ds.HubMicroService")
	}

	localIP := getLocalIP()

	return func(c *gin.Context) {
		startTime := time.Now()
		startTimeMillis := startTime.UnixMilli()

		// Capture request body
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			// Restore the request body so it can be read again by handlers
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Wrap the response writer
		recorder := &ResponseRecorder{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
			statusCode:     0,
		}
		c.Writer = recorder

		// Process request
		c.Next()

		// Calculate end time and duration
		endTime := time.Now()
		endTimeMillis := endTime.UnixMilli()
		duration := endTime.Sub(startTime).Milliseconds()

		// Get status code
		statusCode := recorder.Status()

		// Get response body
		var responseString string
		contentEncoding := c.Writer.Header().Get("Content-Encoding")
		if strings.Contains(contentEncoding, "gzip") {
			decompressed, err := decompressGzip(recorder.body.Bytes())
			if err == nil {
				responseString = decompressed
			} else {
				responseString = recorder.body.String()
			}
		} else {
			responseString = recorder.body.String()
		}

		// Truncate bodies if needed
		requestBodyStr := string(requestBody)
		if config.MaxBodySize > 0 {
			if len(requestBodyStr) > config.MaxBodySize {
				requestBodyStr = requestBodyStr[:config.MaxBodySize] + "... [truncated]"
			}
			if len(responseString) > config.MaxBodySize {
				responseString = responseString[:config.MaxBodySize] + "... [truncated]"
			}
		}

		// Get request ID from header or context
		requestID := c.GetHeader(security.HeaderRequestID)
		if requestID == "" {
			requestID = security.GetRequestID(c)
		}

		// Extract response code from response body
		respCode := extractResponseCode(responseString)

		// Build operation from method and path
		operation := c.Request.URL.Path

		// Build data field with dynamic content
		data := make(map[string]interface{})
		data["method"] = c.Request.Method
		data["path"] = c.Request.URL.Path
		data["clientIp"] = c.ClientIP()

		// Add query params if present
		if c.Request.URL.RawQuery != "" {
			data["query"] = c.Request.URL.RawQuery
		}

		// Add user agent
		if userAgent := c.Request.UserAgent(); userAgent != "" {
			data["userAgent"] = userAgent
		}

		// Add caller if present
		if caller := c.GetHeader("X-Caller"); caller != "" {
			data["caller"] = caller
		}

		// Add request body if configured
		if config.IncludeRequestBody && requestBodyStr != "" {
			// Try to parse as JSON for better formatting
			var reqJSON interface{}
			if err := json.Unmarshal([]byte(requestBodyStr), &reqJSON); err == nil {
				data["request"] = reqJSON
			} else {
				data["request"] = requestBodyStr
			}
		}

		// Add response body if configured
		if config.IncludeResponseBody && responseString != "" {
			// Try to parse as JSON for better formatting
			var respJSON interface{}
			if err := json.Unmarshal([]byte(responseString), &respJSON); err == nil {
				data["response"] = respJSON
			} else {
				data["response"] = responseString
			}
		}

		// Create structured log fields following the specified format
		logFields := logrus.Fields{
			"module":      config.Module,
			"operation":   operation,
			"opVer":       config.Version,
			"requestId":   requestID,
			"localIp":     localIP,
			"httpCode":    statusCode,
			"respCode":    respCode,
			"startTime":   startTimeMillis,
			"endTime":     endTimeMillis,
			"duration_ms": duration,
			"data":        data,
		}

		// Log at appropriate level based on status code
		logEntry := logger.WithFields(logFields)
		if statusCode >= 500 {
			logEntry.Error("HTTP Request")
		} else if statusCode >= 400 {
			logEntry.Warn("HTTP Request")
		} else {
			logEntry.Info("HTTP Request")
		}
	}
}

// MinimalStructuredLoggerMiddleware creates a lightweight structured logging middleware
// with less verbose output for high-throughput scenarios
func MinimalStructuredLoggerMiddleware(logger *logrus.Logger, config *Config) gin.HandlerFunc {
	if config == nil {
		config = DefaultConfig("ds.HubMicroService")
	}

	localIP := getLocalIP()

	return func(c *gin.Context) {
		startTime := time.Now()
		startTimeMillis := startTime.UnixMilli()

		// Process request without capturing bodies
		c.Next()

		// Calculate end time and duration
		endTime := time.Now()
		endTimeMillis := endTime.UnixMilli()
		duration := endTime.Sub(startTime).Milliseconds()

		// Get status code
		statusCode := c.Writer.Status()

		// Get request ID
		requestID := c.GetHeader(security.HeaderRequestID)
		if requestID == "" {
			requestID = security.GetRequestID(c)
		}

		// Build operation
		operation := c.Request.URL.Path

		// Minimal data field
		data := map[string]interface{}{
			"method":   c.Request.Method,
			"clientIp": c.ClientIP(),
		}

		if c.Request.URL.RawQuery != "" {
			data["query"] = c.Request.URL.RawQuery
		}

		// Create structured log fields
		logFields := logrus.Fields{
			"module":      config.Module,
			"operation":   operation,
			"opVer":       config.Version,
			"requestId":   requestID,
			"localIp":     localIP,
			"httpCode":    statusCode,
			"respCode":    "OK",
			"startTime":   startTimeMillis,
			"endTime":     endTimeMillis,
			"duration_ms": duration,
			"data":        data,
		}

		// Log based on status
		logEntry := logger.WithFields(logFields)
		if statusCode >= 500 {
			logEntry.Error("HTTP Request")
		} else if statusCode >= 400 {
			logEntry.Warn("HTTP Request")
		} else {
			logEntry.Info("HTTP Request")
		}
	}
}

// NewConfig creates a new configuration for the middleware
func NewConfig(module string) *Config {
	return &Config{
		Module:              module,
		Version:             GetVersionFromEnv(),
		IncludeRequestBody:  true,
		IncludeResponseBody: true,
		MaxBodySize:         GetMaxBodySizeFromEnv(),
	}
}

// WithModule sets the module name
func (c *Config) WithModule(module string) *Config {
	c.Module = module
	return c
}

// WithVersion sets the version
func (c *Config) WithVersion(version string) *Config {
	c.Version = version
	return c
}

// WithBodyLogging configures body logging
func (c *Config) WithBodyLogging(includeRequest, includeResponse bool) *Config {
	c.IncludeRequestBody = includeRequest
	c.IncludeResponseBody = includeResponse
	return c
}

// WithMaxBodySize sets the maximum body size to log
func (c *Config) WithMaxBodySize(size int) *Config {
	c.MaxBodySize = size
	return c
}

// GetVersionFromEnv gets the version from environment variable
// Returns the value from SERVICE_VERSION env var, or "v1" as default
func GetVersionFromEnv() string {
	if version := os.Getenv("SERVICE_VERSION"); version != "" {
		return version
	}
	return "v1"
}

// GetMaxBodySizeFromEnv gets the maximum body size from environment variable
// Returns the value from MAX_BODY_SIZE env var, or 10000 (10KB) as default
func GetMaxBodySizeFromEnv() int {
	const defaultSize = 10000 // 10KB default

	if sizeStr := os.Getenv("MAX_BODY_SIZE"); sizeStr != "" {
		var size int
		if _, err := fmt.Sscanf(sizeStr, "%d", &size); err == nil && size >= 0 {
			return size
		}
	}

	return defaultSize
}

// GetModuleFromEnv gets the module name from environment variable
func GetModuleFromEnv() string {
	module := os.Getenv("SERVICE_MODULE")
	if module == "" {
		// Try to infer from service name
		serviceName := os.Getenv("SERVICE_NAME")
		switch serviceName {
		default:
			return "game-platform"
		}
	}
	return module
}
