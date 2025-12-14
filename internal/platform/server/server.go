package server

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"game-platform/internal/platform/accesslog"
	"game-platform/internal/platform/app"
	dynamicconfig "game-platform/internal/platform/config"
	"game-platform/internal/platform/security"
	"game-platform/pkg/discovery"
	"game-platform/pkg/metrics"
)

// Server wraps the HTTP server with middleware and lifecycle management
type Server struct {
	app    *app.App
	router *gin.Engine
	srv    *http.Server
}

// RouteInitializer is a function that registers routes for a specific service
type RouteInitializer func(*gin.Engine, *app.App)

// NewServer creates a new server with all middleware configured
func NewServer(app *app.App, port string, serviceName string, initRoutes RouteInitializer) *Server {
	// Set Gin to release mode to disable debug logs
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = io.Discard

	// Create router
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(accesslog.LoggerMiddleware(app.Logger))

	// Configure CORS
	corsConfig := getCorsConfig(app.DynamicConfig)
	router.Use(cors.New(corsConfig))

	// Add security headers
	router.Use(func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	})

	router.Use(gzip.Gzip(gzip.BestCompression))

	// Configure API versioning
	if app.DynamicConfig != nil {
		versionConfig := getVersioningConfig(app.DynamicConfig)
		publicAPI := router.Group("/api/v1")
		publicAPI.Use(security.VersioningMiddleware(versionConfig))

		// Apply request validation if enabled
		requestValidationConfig := getRequestValidationConfig(app.DynamicConfig)
		if requestValidationConfig.ValidateSignature || requestValidationConfig.RequireBase64Encoding || requestValidationConfig.ValidateKey {
			publicAPI.Use(security.RequestValidatorMiddleware(requestValidationConfig))
		}
	}

	// Initialize metrics
	metrics.Initialize()
	metrics.RegisterMetricsEndpoint(router)

	// Initialize routes for this service
	initRoutes(router, app)

	// Register service discovery if enabled
	if os.Getenv("ENABLE_SERVICE_DISCOVERY") == "true" {
		consulAddress := os.Getenv("CONSUL_ADDRESS")
		if consulAddress == "" {
			consulAddress = "localhost:8500"
		}

		serviceRegistry, err := discovery.NewConsulServiceRegistry(consulAddress)
		if err != nil {
			log.Printf("Failed to initialize service discovery: %v", err)
		} else {
			if serviceName == "" {
				serviceName = "micro-backend-service"
			}

			err = serviceRegistry.RegisterSelf(serviceName)
			if err != nil {
				log.Printf("Failed to register service: %v", err)
			} else {
				log.Printf("Service registered with Consul: %s", serviceName)
			}
		}
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	return &Server{
		app:    app,
		router: router,
		srv:    srv,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	log.Printf("Server started on port %s", s.srv.Addr)
	if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Shutting down gracefully...")
	return s.srv.Shutdown(ctx)
}

// getCorsConfig creates a CORS configuration from dynamic config
func getCorsConfig(dynamicConfig *dynamicconfig.DynamicConfig) cors.Config {
	corsConfig := cors.Config{
		AllowOrigins:     []string{"https://your-domain.com"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	if dynamicConfig == nil {
		return corsConfig
	}

	if corsInterface, ok := dynamicConfig.Get("security.cors"); ok {
		if corsMap, ok := corsInterface.(map[string]interface{}); ok {
			if originsInterface, ok := corsMap["allowed_origins"].([]interface{}); ok {
				origins := make([]string, 0, len(originsInterface))
				for _, origin := range originsInterface {
					if originStr, ok := origin.(string); ok {
						origins = append(origins, originStr)
					}
				}
				if os.Getenv("GO_ENV") == "production" && len(origins) == 1 && origins[0] == "*" {
					logrus.Warn("Using wildcard CORS origin in production is not recommended")
				}
				corsConfig.AllowOrigins = origins
			}

			if methodsInterface, ok := corsMap["allowed_methods"].([]interface{}); ok {
				methods := make([]string, 0, len(methodsInterface))
				for _, method := range methodsInterface {
					if methodStr, ok := method.(string); ok {
						methods = append(methods, methodStr)
					}
				}
				corsConfig.AllowMethods = methods
			}

			if headersInterface, ok := corsMap["allowed_headers"].([]interface{}); ok {
				headers := make([]string, 0, len(headersInterface))
				for _, header := range headersInterface {
					if headerStr, ok := header.(string); ok {
						headers = append(headers, headerStr)
					}
				}
				corsConfig.AllowHeaders = headers
			}

			if allowCredentials, ok := corsMap["allow_credentials"].(bool); ok {
				corsConfig.AllowCredentials = allowCredentials
			}

			if maxAge, ok := corsMap["max_age"].(float64); ok {
				corsConfig.MaxAge = time.Duration(maxAge) * time.Hour
			}
		}
	}

	return corsConfig
}

// getVersioningConfig creates a versioning configuration from dynamic config
func getVersioningConfig(dynamicConfig *dynamicconfig.DynamicConfig) security.VersioningConfig {
	versionConfig := security.DefaultVersioningConfig()

	if majorVersion, ok := dynamicConfig.GetInt("api.version.current.major"); ok {
		versionConfig.Current.Major = majorVersion
	}
	if minorVersion, ok := dynamicConfig.GetInt("api.version.current.minor"); ok {
		versionConfig.Current.Minor = minorVersion
	}
	if patchVersion, ok := dynamicConfig.GetInt("api.version.current.patch"); ok {
		versionConfig.Current.Patch = patchVersion
	}

	versionConfig.MaxSupported = versionConfig.Current

	if minMajor, ok := dynamicConfig.GetInt("api.version.min_supported.major"); ok {
		versionConfig.MinSupported.Major = minMajor
	}
	if minMinor, ok := dynamicConfig.GetInt("api.version.min_supported.minor"); ok {
		versionConfig.MinSupported.Minor = minMinor
	}
	if minPatch, ok := dynamicConfig.GetInt("api.version.min_supported.patch"); ok {
		versionConfig.MinSupported.Patch = minPatch
	}

	if headerName, ok := dynamicConfig.GetString("api.version.header_name"); ok {
		versionConfig.HeaderName = headerName
	}
	if checkURLPath, ok := dynamicConfig.GetBool("api.version.check_url_path"); ok {
		versionConfig.CheckURLPath = checkURLPath
	}
	if sendDeprecationWarnings, ok := dynamicConfig.GetBool("api.version.send_deprecation_warnings"); ok {
		versionConfig.SendDeprecationWarnings = sendDeprecationWarnings
	}

	return versionConfig
}

// getRequestValidationConfig creates a request validation configuration from dynamic config
func getRequestValidationConfig(dynamicConfig *dynamicconfig.DynamicConfig) security.RequestValidatorConfig {
	validatorConfig := security.DefaultRequestValidatorConfig()

	enabled, _ := dynamicConfig.GetBool("security.request_validation.enabled")
	if !enabled {
		validatorConfig.ValidateSignature = false
		validatorConfig.ValidateKey = false
		validatorConfig.RequireBase64Encoding = false
		return validatorConfig
	}

	if validateSignature, ok := dynamicConfig.GetBool("security.request_validation.validate_signature"); ok {
		validatorConfig.ValidateSignature = validateSignature
	}
	if validateKey, ok := dynamicConfig.GetBool("security.request_validation.validate_key"); ok {
		validatorConfig.ValidateKey = validateKey
	}
	if requireBase64Encoding, ok := dynamicConfig.GetBool("security.request_validation.require_base64_encoding"); ok {
		validatorConfig.RequireBase64Encoding = requireBase64Encoding
	}

	if skipPathsInterface, ok := dynamicConfig.Get("security.request_validation.skip_paths"); ok {
		if skipPaths, ok := skipPathsInterface.([]interface{}); ok {
			paths := make([]string, 0, len(skipPaths))
			for _, path := range skipPaths {
				if pathStr, ok := path.(string); ok {
					paths = append(paths, pathStr)
				}
			}
			validatorConfig.SkipPaths = paths
		}
	}

	return validatorConfig
}
