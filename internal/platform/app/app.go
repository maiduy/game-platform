package app

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	dynamicconfig "game-platform/internal/platform/config"
	pool "game-platform/internal/platform/pools"
	util "game-platform/internal/platform/utils"
)

// App holds all shared application resources
type App struct {
	DB            *gorm.DB
	Redis         pool.RedisInterface
	MongoConn     *pool.MongoDBConnection
	DynamicConfig *dynamicconfig.DynamicConfig
	Logger        *logrus.Logger
}

// Initialize sets up all shared application resources
func Initialize() (*App, error) {
	// Load MACHINE_NODE from .env and convert to int64
	nodeIDStr := util.GodotEnv("MACHINE_NODE")
	nodeID, err := strconv.ParseInt(nodeIDStr, 10, 64)
	if err != nil {
		log.Fatalf("Invalid MACHINE_NODE value: %v", err)
	}

	// Initialize Snowflake node
	if err := util.InitSnowflakeNode(nodeID); err != nil {
		log.Fatalf("Failed to initialize Snowflake node: %v", err)
	}

	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	// Setup Database Connection (PostgreSQL)
	dbConn, err := pool.NewPostgresConnection()
	if err != nil {
		log.Printf("Failed to initialize PostgreSQL connection: %v", err)
	} else {
		log.Println("PostgreSQL connection established successfully")
	}

	var db *gorm.DB
	if dbConn != nil {
		db = dbConn.DB
	}

	// Setup Redis Connection based on REDIS_MODE environment variable
	var redisClient pool.RedisInterface
	redisMode := os.Getenv("REDIS_MODE")

	if redisMode == "cluster" {
		redisCluster, err := pool.NewRedisClusterConnection()
		if err != nil {
			log.Printf("Failed to connect to Redis Cluster: %v", err)
		} else {
			log.Println("Redis Cluster connection established successfully")
			redisClient = redisCluster
		}
	} else {
		// Default to Sentinel mode
		redisSentinel, err := pool.NewRedisConnection()
		if err != nil {
			log.Printf("Failed to connect to Redis Sentinel: %v", err)
		} else {
			log.Println("Redis Sentinel connection established successfully")
			redisClient = redisSentinel
		}
	}

	// Setup GDO MongoDB Connection
	gdoMongoConn, err := pool.NewGDOMongoDBConnection()
	if err != nil {
		log.Printf("Failed to initialize GDO MongoDB connection: %v", err)
	}

	// Load dynamic configuration
	var dynamicConfig *dynamicconfig.DynamicConfig
	if os.Getenv("ENABLE_DYNAMIC_CONFIG") == "true" {
		configFilePath := os.Getenv("CONFIG_FILE_PATH")
		if configFilePath == "" {
			configFilePath = "configs/app_config.json"
		}

		configSource := dynamicconfig.NewFileConfigSource(configFilePath, 30*time.Second)
		dynamicConfig, err = dynamicconfig.NewDynamicConfig(configSource)
		if err != nil {
			log.Printf("Failed to initialize dynamic configuration: %v", err)
		} else {
			log.Printf("Dynamic configuration initialized from %s", configFilePath)

			// Monitor configuration changes
			dynamicConfig.OnChange(func() {
				log.Println("Configuration changed, reloading...")
			})
		}
	}

	return &App{
		DB:            db,
		Redis:         redisClient,
		MongoConn:     gdoMongoConn,
		DynamicConfig: dynamicConfig,
		Logger:        logger,
	}, nil
}

// Close cleans up all application resources
func (a *App) Close() {
	if a.DB != nil {
		sqlDB, err := a.DB.DB()
		if err == nil {
			sqlDB.Close()
		}
	}
	if a.Redis != nil {
		a.Redis.Close()
	}
	if a.MongoConn != nil {
		a.MongoConn.Close()
	}
}
