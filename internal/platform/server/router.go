package server

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	dynamicconfig "game-platform/internal/platform/config"
	pool "game-platform/internal/platform/pools"
	"game-platform/internal/platform/server/routes"
)

// RouterConfig holds dependencies needed for route initialization
type RouterConfig struct {
	Logger        *logrus.Logger
	MongoConn     *pool.MongoDBConnection
	DynamicConfig *dynamicconfig.DynamicConfig
}

// InitializeRoutes is the centralized entry point for all platform routes
// Add new feature routes here to keep routing modular and maintainable
func InitializeRoutes(router *gin.Engine, config RouterConfig) {
	// Register MasterConfig routes
	routes.RegisterMasterConfigRoutes(router, config.Logger, config.MongoConn, config.DynamicConfig)

	// Register Ability routes
	routes.RegisterAbilityRoutes(router, config.Logger, config.MongoConn, config.DynamicConfig)

	// Future routes can be added here:
	// routes.RegisterLeaderboardRoutes(router, config.Logger, config.MongoConn, config.DynamicConfig)
}
