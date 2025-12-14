package security

import (
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// WhitelistConfig holds configuration for IP and domain whitelisting
type WhitelistConfig struct {
	// List of allowed IP addresses or CIDR ranges
	AllowedIPs []string

	// List of allowed domain names
	AllowedDomains []string

	// Whether to allow localhost connections
	AllowLocalhost bool

	// Whether to check Origin header for domains
	CheckOrigin bool

	// Whether to check Referer header for domains
	CheckReferer bool
}

// DefaultWhitelistConfig returns a default whitelist configuration
func DefaultWhitelistConfig() WhitelistConfig {
	return WhitelistConfig{
		AllowedIPs:     []string{},
		AllowedDomains: []string{},
		AllowLocalhost: true,
		CheckOrigin:    true,
		CheckReferer:   true,
	}
}

// WhitelistMiddleware returns a middleware that restricts access based on IP and domain
func WhitelistMiddleware(config WhitelistConfig) gin.HandlerFunc {
	// Preprocess IP whitelist for faster lookup
	ipNets := make([]*net.IPNet, 0, len(config.AllowedIPs))
	singleIPs := make(map[string]bool)

	for _, ipStr := range config.AllowedIPs {
		// Check if it's a CIDR notation
		if strings.Contains(ipStr, "/") {
			_, ipNet, err := net.ParseCIDR(ipStr)
			if err == nil {
				ipNets = append(ipNets, ipNet)
			}
		} else {
			// It's a single IP
			singleIPs[ipStr] = true
		}
	}

	// Preprocess domain whitelist for faster lookup
	domainMap := make(map[string]bool)
	for _, domain := range config.AllowedDomains {
		domainMap[strings.ToLower(domain)] = true
	}

	return func(c *gin.Context) {
		logger := logrus.WithContext(c)

		// Check IP whitelist
		clientIP := c.ClientIP()

		// Allow localhost if configured
		if config.AllowLocalhost && isLocalhost(clientIP) {
			c.Next()
			return
		}

		// Check if IP is explicitly whitelisted
		if singleIPs[clientIP] {
			c.Next()
			return
		}

		// Check if IP is in any of the allowed CIDR ranges
		ip := net.ParseIP(clientIP)
		if ip != nil {
			for _, ipNet := range ipNets {
				if ipNet.Contains(ip) {
					c.Next()
					return
				}
			}
		}

		// Check domain whitelist if enabled
		if len(config.AllowedDomains) > 0 {
			// Check Origin header
			if config.CheckOrigin {
				origin := c.GetHeader("Origin")
				if origin != "" {
					domain := extractDomain(origin)
					if domainMap[strings.ToLower(domain)] {
						c.Next()
						return
					}
				}
			}

			// Check Referer header
			if config.CheckReferer {
				referer := c.GetHeader("Referer")
				if referer != "" {
					domain := extractDomain(referer)
					if domainMap[strings.ToLower(domain)] {
						c.Next()
						return
					}
				}
			}
		}

		// If we get here, the request is not allowed
		logger.Warnf("Access denied for IP: %s", clientIP)
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Access denied",
		})
		c.Abort()
	}
}

// DomainWhitelistMiddleware returns a middleware that restricts access based on domain
func DomainWhitelistMiddleware(allowedDomains []string) gin.HandlerFunc {
	config := DefaultWhitelistConfig()
	config.AllowedDomains = allowedDomains
	config.AllowedIPs = []string{}
	return WhitelistMiddleware(config)
}

// EnhancedIPWhitelistMiddleware returns a middleware that restricts access based on IP
// This is an enhanced version of IPWhitelistMiddleware that supports CIDR notation
func EnhancedIPWhitelistMiddleware(whitelist []string) gin.HandlerFunc {
	config := DefaultWhitelistConfig()
	config.AllowedIPs = whitelist
	config.AllowedDomains = []string{}
	return WhitelistMiddleware(config)
}

// Helper functions

// isLocalhost checks if an IP address is a localhost address
func isLocalhost(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	// Check if it's a loopback address
	if ip.IsLoopback() {
		return true
	}

	// Check common localhost addresses
	if ipStr == "127.0.0.1" || ipStr == "::1" {
		return true
	}

	return false
}

// extractDomain extracts the domain from a URL
func extractDomain(urlStr string) string {
	// Remove protocol
	domain := urlStr
	if idx := strings.Index(domain, "://"); idx != -1 {
		domain = domain[idx+3:]
	}

	// Remove path and query
	if idx := strings.Index(domain, "/"); idx != -1 {
		domain = domain[:idx]
	}

	// Remove port
	if idx := strings.Index(domain, ":"); idx != -1 {
		domain = domain[:idx]
	}

	return domain
}
