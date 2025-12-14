package security

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// APIVersion represents a specific API version
type APIVersion struct {
	Major int
	Minor int
	Patch int
}

// VersioningConfig holds configuration for API versioning
type VersioningConfig struct {
	// Current API version
	Current APIVersion

	// Minimum supported API version
	MinSupported APIVersion

	// Maximum supported API version (usually same as Current)
	MaxSupported APIVersion

	// Header name for API version
	HeaderName string

	// Whether to check URL path for version
	CheckURLPath bool

	// Whether to send deprecation warnings for old versions
	SendDeprecationWarnings bool
}

// DefaultVersioningConfig returns a default versioning configuration
func DefaultVersioningConfig() VersioningConfig {
	return VersioningConfig{
		Current: APIVersion{
			Major: 1,
			Minor: 0,
			Patch: 0,
		},
		MinSupported: APIVersion{
			Major: 1,
			Minor: 0,
			Patch: 0,
		},
		MaxSupported: APIVersion{
			Major: 1,
			Minor: 0,
			Patch: 0,
		},
		HeaderName:              "X-API-Version",
		CheckURLPath:            true,
		SendDeprecationWarnings: true,
	}
}

// VersioningMiddleware returns a middleware that handles API versioning
func VersioningMiddleware(config VersioningConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := logrus.WithContext(c)

		// Try to get version from header first
		version, found := getVersionFromHeader(c, config.HeaderName)

		// If not found in header and URL path checking is enabled, try to get from URL
		if !found && config.CheckURLPath {
			version, found = getVersionFromURL(c)
		}

		// If version is still not found, use current version
		if !found {
			// Set current version in context
			c.Set("apiVersion", config.Current)
			c.Next()
			return
		}

		// Check if version is supported
		if !isVersionSupported(version, config.MinSupported, config.MaxSupported) {
			logger.Warnf("Unsupported API version requested: v%d.%d.%d",
				version.Major, version.Minor, version.Patch)

			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Unsupported API version",
				"supported": gin.H{
					"min": gin.H{
						"major": config.MinSupported.Major,
						"minor": config.MinSupported.Minor,
						"patch": config.MinSupported.Patch,
					},
					"max": gin.H{
						"major": config.MaxSupported.Major,
						"minor": config.MaxSupported.Minor,
						"patch": config.MaxSupported.Patch,
					},
				},
			})
			c.Abort()
			return
		}

		// Check if version is deprecated
		if config.SendDeprecationWarnings && isVersionDeprecated(version, config.Current) {
			c.Header("X-API-Deprecated", "true")
			c.Header("X-API-Recommended-Version",
				formatVersion(config.Current))
		}

		// Set version in context
		c.Set("apiVersion", version)
		c.Next()
	}
}

// getVersionFromHeader extracts API version from request header
func getVersionFromHeader(c *gin.Context, headerName string) (APIVersion, bool) {
	versionStr := c.GetHeader(headerName)
	if versionStr == "" {
		return APIVersion{}, false
	}

	return parseVersionString(versionStr)
}

// getVersionFromURL extracts API version from URL path
// Expects format like /v1/resource or /api/v2.1/resource
func getVersionFromURL(c *gin.Context) (APIVersion, bool) {
	path := c.Request.URL.Path
	segments := strings.Split(path, "/")

	for _, segment := range segments {
		if len(segment) > 0 && (segment[0] == 'v' || segment[0] == 'V') {
			versionStr := segment[1:]
			version, found := parseVersionString(versionStr)
			if found {
				return version, true
			}
		}
	}

	return APIVersion{}, false
}

// parseVersionString parses a version string like "1.0.0" or "2.1" into an APIVersion
func parseVersionString(versionStr string) (APIVersion, bool) {
	var major, minor, patch int
	var err error

	// Remove 'v' prefix if present
	if len(versionStr) > 0 && (versionStr[0] == 'v' || versionStr[0] == 'V') {
		versionStr = versionStr[1:]
	}

	parts := strings.Split(versionStr, ".")
	if len(parts) < 1 {
		return APIVersion{}, false
	}

	// Parse major version
	major, err = parseInt(parts[0])
	if err != nil {
		return APIVersion{}, false
	}

	// Parse minor version if present
	minor = 0
	if len(parts) > 1 {
		minor, err = parseInt(parts[1])
		if err != nil {
			return APIVersion{}, false
		}
	}

	// Parse patch version if present
	patch = 0
	if len(parts) > 2 {
		patch, err = parseInt(parts[2])
		if err != nil {
			return APIVersion{}, false
		}
	}

	return APIVersion{
		Major: major,
		Minor: minor,
		Patch: patch,
	}, true
}

// parseInt parses a string to an int, returning an error if parsing fails
func parseInt(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}

// isVersionSupported checks if a version is within the supported range
func isVersionSupported(version, min, max APIVersion) bool {
	// Compare major version first
	if version.Major < min.Major || version.Major > max.Major {
		return false
	}

	// If major version is at minimum, check minor version
	if version.Major == min.Major && version.Minor < min.Minor {
		return false
	}

	// If major version is at maximum, check minor version
	if version.Major == max.Major && version.Minor > max.Minor {
		return false
	}

	// If major and minor versions are at minimum, check patch version
	if version.Major == min.Major && version.Minor == min.Minor && version.Patch < min.Patch {
		return false
	}

	// If major and minor versions are at maximum, check patch version
	if version.Major == max.Major && version.Minor == max.Minor && version.Patch > max.Patch {
		return false
	}

	return true
}

// isVersionDeprecated checks if a version is deprecated compared to the current version
func isVersionDeprecated(version, current APIVersion) bool {
	// Only consider major and minor versions for deprecation
	return version.Major < current.Major ||
		(version.Major == current.Major && version.Minor < current.Minor)
}

// formatVersion formats an APIVersion as a string
func formatVersion(version APIVersion) string {
	return fmt.Sprintf("v%d.%d.%d", version.Major, version.Minor, version.Patch)
}
