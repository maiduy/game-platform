package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"game-platform/internal/platform/auth"
	dynamicconfig "game-platform/internal/platform/config"
	pool "game-platform/internal/platform/pools"
	api "game-platform/internal/services/masterconfig/api/http"
	"game-platform/internal/services/masterconfig/application"
	"game-platform/internal/services/masterconfig/repository"
	"game-platform/internal/services/masterconfig/resolver"
)

// RegisterMasterConfigRoutes registers all MasterConfig service routes
func RegisterMasterConfigRoutes(
	router *gin.Engine,
	logger *logrus.Logger,
	mongoConn *pool.MongoDBConnection,
	dynamicConfig *dynamicconfig.DynamicConfig,
) {
	// Initialize characters repository, resolver, service, and handler
	charactersRepo := repository.NewRepository(mongoConn)
	charactersResolver := resolver.NewCharacterResolver(charactersRepo)
	charactersService := application.NewCharacterService(logger, charactersResolver)
	charactersHandler := api.NewHandler(logger, charactersService)

	// Define API group for characters routes
	charactersGroup := router.Group("/api/v1/masterconfig/characters")

	// Apply authentication based on config
	applyMasterConfigAuth(charactersGroup, logger, dynamicConfig)

	// Register characters routes
	charactersGroup.POST("", charactersHandler.CreateCharacter)
	charactersGroup.GET("", charactersHandler.ListCharacters)
	charactersGroup.GET("/:character_id", charactersHandler.GetCharacterByCharacterID)
	charactersGroup.PUT("/:character_id", charactersHandler.UpdateCharacterByCharacterID)

	logger.Info("MasterConfig routes initialized successfully")
}

// applyMasterConfigAuth applies authentication middleware to the route group
func applyMasterConfigAuth(group *gin.RouterGroup, logger *logrus.Logger, dynamicConfig *dynamicconfig.DynamicConfig) {
	// Check for service-specific auth config
	if masterConfigAuthConfig, ok := dynamicConfig.Get("security.routes.masterconfig.auth"); ok {
		if config, ok := masterConfigAuthConfig.(map[string]interface{}); ok {
			if enabled, ok := config["enabled"].(bool); ok && enabled {
				authConfig := buildAuthConfig(config)
				group.Use(auth.AuthMiddleware(authConfig))
				logger.Infof("Applied authentication to MasterConfig routes with strategy: %s", authConfig.Strategy)
				return
			}
		}
	}

	// Fall back to global auth config
	if authEnabled, ok := dynamicConfig.GetBool("security.auth.enabled"); ok && authEnabled {
		authConfig := auth.AuthConfig{Strategy: auth.StrategyJWT}
		if strategy, ok := dynamicConfig.GetString("security.auth.strategy"); ok {
			authConfig.Strategy = auth.AuthStrategy(strategy)
		}
		group.Use(auth.AuthMiddleware(authConfig))
		logger.Info("Applied global authentication to MasterConfig routes")
	}
}

// buildAuthConfig constructs an AuthConfig from a config map
func buildAuthConfig(config map[string]interface{}) auth.AuthConfig {
	authConfig := auth.AuthConfig{Strategy: auth.StrategyJWT}

	if strategyStr, ok := config["strategy"].(string); ok {
		authConfig.Strategy = auth.AuthStrategy(strategyStr)
	}

	if rolesInterface, ok := config["required_roles"].([]interface{}); ok {
		roles := make([]string, 0, len(rolesInterface))
		for _, role := range rolesInterface {
			if roleStr, ok := role.(string); ok {
				roles = append(roles, roleStr)
			}
		}
		authConfig.RequiredRoles = roles
	}

	return authConfig
}
