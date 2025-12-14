package api

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"game-platform/internal/platform/auth"
	dynamicconfig "game-platform/internal/platform/config"
	pool "game-platform/internal/platform/pools"
	"game-platform/internal/services/masterconfig/application"
	"game-platform/internal/services/masterconfig/repository"
	"game-platform/internal/services/masterconfig/resolver"
)

// InitMasterConfigRoutes initializes the routes for masterconfig service
func InitMasterConfigRoutes(
	logger *logrus.Logger,
	mongoConn *pool.MongoDBConnection,
	router *gin.Engine,
	dynamicConfig *dynamicconfig.DynamicConfig,
) {
	// Initialize characters repository, resolver, service, and handler
	charactersRepo := repository.NewRepository(mongoConn)
	charactersResolver := resolver.NewCharacterResolver(charactersRepo)
	charactersService := application.NewCharacterService(logger, charactersResolver)
	charactersHandler := NewHandler(logger, charactersService)

	// Define API group for characters routes
	charactersGroup := router.Group("/api/v1/masterconfig/characters")

	// Apply security configuration to characters routes
	// Check if masterconfig has its own auth configuration
	masterConfigAuthConfig, ok := dynamicConfig.Get("security.routes.masterconfig.auth")
	if ok {
		if config, ok := masterConfigAuthConfig.(map[string]interface{}); ok {
			if enabled, ok := config["enabled"].(bool); ok && enabled {
				authConfig := auth.AuthConfig{
					Strategy: auth.StrategyJWT,
				}

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

				charactersGroup.Use(auth.AuthMiddleware(authConfig))
				logger.Infof("Applied authentication to characters routes with strategy: %s", authConfig.Strategy)
			}
		}
	} else {
		// Fall back to global auth config
		if authEnabled, ok := dynamicConfig.GetBool("security.auth.enabled"); ok && authEnabled {
			authConfig := auth.AuthConfig{
				Strategy: auth.StrategyJWT,
			}

			if strategy, ok := dynamicConfig.GetString("security.auth.strategy"); ok {
				authConfig.Strategy = auth.AuthStrategy(strategy)
			}

			charactersGroup.Use(auth.AuthMiddleware(authConfig))
			logger.Info("Applied global authentication to characters routes")
		}
	}

	// Register characters routes
	{
		charactersGroup.POST("", charactersHandler.CreateCharacter)                           // Create a new character
		charactersGroup.GET("", charactersHandler.ListCharacters)                             // Retrieve all characters
		charactersGroup.GET("/:character_id", charactersHandler.GetCharacterByCharacterID)    // Retrieve a character by Character ID
		charactersGroup.PUT("/:character_id", charactersHandler.UpdateCharacterByCharacterID) // Update a character by Character ID
	}

	// TODO: Register game-specific character routes with entity normalization
	// Uncomment when X3 and 4F handlers are implemented
	/*
		{
			// X3 Game - Character routes with X3-specific field mappings
			handlerX3 := NewX3Handler(logger, charactersSvc)
			charactersGroup.GET("/X3", handlerX3.ListCharacters)                          // List characters with X3 normalization
			charactersGroup.GET("/X3/:character_id", handlerX3.GetCharacterByCharacterID) // Get character with X3 normalization

			// 4F Game - Character routes with 4F-specific field mappings
			handler4F := New4FHandler(logger, charactersSvc)
			charactersGroup.GET("/4F", handler4F.ListCharacters)                          // List characters with 4F normalization
			charactersGroup.GET("/4F/:character_id", handler4F.GetCharacterByCharacterID) // Get character with 4F normalization
		}
	*/

	logger.Info("Characters routes initialized successfully")
	// logger.Info("Game-specific character routes (X3, 4F) initialized successfully")
}
