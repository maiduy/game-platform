package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"game-platform/internal/platform/auth"
	dynamicconfig "game-platform/internal/platform/config"
	pool "game-platform/internal/platform/pools"
	api "game-platform/internal/services/ability/api/http"
	"game-platform/internal/services/ability/application"
	"game-platform/internal/services/ability/repository"
	"game-platform/internal/services/ability/resolver"
)

// RegisterAbilityRoutes registers all Ability service routes (effects, skills, etc.)
func RegisterAbilityRoutes(
	router *gin.Engine,
	logger *logrus.Logger,
	mongoConn *pool.MongoDBConnection,
	dynamicConfig *dynamicconfig.DynamicConfig,
) {
	// Initialize effect repository, resolver, service, and handler
	effectRepo := repository.NewRepository(mongoConn)
	effectResolver := resolver.NewEffectResolver(effectRepo)
	effectService := application.NewEffectService(logger, effectResolver)
	effectHandler := api.NewHandler(logger, effectService)

	// Define API group for effect routes
	effectsGroup := router.Group("/api/v1/ability/effects")

	// Apply authentication based on config
	applyAbilityAuth(effectsGroup, logger, dynamicConfig)

	// Register effect routes
	effectsGroup.POST("", effectHandler.CreateEffect)
	effectsGroup.GET("", effectHandler.ListEffects)
	effectsGroup.GET("/:id", effectHandler.GetEffectByID)
	effectsGroup.PUT("/:id", effectHandler.UpdateEffect)
	effectsGroup.DELETE("/:id", effectHandler.DeleteEffect)

	logger.Info("Ability routes initialized successfully")
}

// applyAbilityAuth applies authentication middleware to the ability route group
func applyAbilityAuth(group *gin.RouterGroup, logger *logrus.Logger, dynamicConfig *dynamicconfig.DynamicConfig) {
	// Check for service-specific auth config
	if abilityAuthConfig, ok := dynamicConfig.Get("security.routes.game.auth"); ok {
		if config, ok := abilityAuthConfig.(map[string]interface{}); ok {
			if enabled, ok := config["enabled"].(bool); ok && enabled {
				authConfig := buildAuthConfig(config)
				group.Use(auth.AuthMiddleware(authConfig))
				logger.Infof("Applied authentication to Ability routes with strategy: %s", authConfig.Strategy)
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
		logger.Info("Applied global authentication to Ability routes")
	}
}
