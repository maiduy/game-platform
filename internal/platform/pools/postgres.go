package pool

import (
	"os"
	"strconv"
	"time"

	"game-platform/internal/platform/config"
	util "game-platform/internal/platform/utils"
	"game-platform/pkg/resilience"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// PostgresConfig holds the configuration for the PostgreSQL connection
type PostgresConfig struct {
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
}

// GetPostgresConfig returns the PostgreSQL configuration
func GetPostgresConfig() PostgresConfig {
	// Default values
	config := PostgresConfig{
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: time.Hour,
	}

	// Try to get values from dynamic config if available
	dynamicConfig, err := getPostgresConfigFromDynamic()
	if err == nil {
		if maxIdleConns, ok := dynamicConfig.GetInt("database.max_idle_conns"); ok {
			config.MaxIdleConns = maxIdleConns
		}

		if maxOpenConns, ok := dynamicConfig.GetInt("database.max_open_conns"); ok {
			config.MaxOpenConns = maxOpenConns
		}

		if connMaxLifetime, ok := dynamicConfig.GetInt("database.conn_max_lifetime"); ok {
			config.ConnMaxLifetime = time.Duration(connMaxLifetime) * time.Second
		}
	} else {
		// Fallback to environment variables
		if maxIdleConns, err := strconv.Atoi(os.Getenv("DB_MAX_IDLE_CONNS")); err == nil {
			config.MaxIdleConns = maxIdleConns
		}

		if maxOpenConns, err := strconv.Atoi(os.Getenv("DB_MAX_OPEN_CONNS")); err == nil {
			config.MaxOpenConns = maxOpenConns
		}

		if connMaxLifetime, err := strconv.Atoi(os.Getenv("DB_CONN_MAX_LIFETIME")); err == nil {
			config.ConnMaxLifetime = time.Duration(connMaxLifetime) * time.Second
		}
	}

	return config
}

// getPostgresConfigFromDynamic attempts to get the dynamic configuration for PostgreSQL
func getPostgresConfigFromDynamic() (*config.DynamicConfig, error) {
	// Check if dynamic config is enabled
	if os.Getenv("ENABLE_DYNAMIC_CONFIG") != "true" {
		return nil, nil
	}

	// Get config file path
	configFilePath := os.Getenv("CONFIG_FILE_PATH")
	if configFilePath == "" {
		configFilePath = "configs/app_config.json"
	}

	// Create config source
	configSource := config.NewFileConfigSource(configFilePath, 30*time.Second)

	// Create dynamic config
	return config.NewDynamicConfig(configSource)
}

// PostgresConnection represents a connection to PostgreSQL with resilience features
type PostgresConnection struct {
	DB *gorm.DB
	cb *resilience.CircuitBreaker
}

// NewPostgresConnection creates and initializes a new PostgreSQL connection
func NewPostgresConnection() (*PostgresConnection, error) {
	conn := &PostgresConnection{}
	if err := conn.Initialize(); err != nil {
		return nil, err
	}
	return conn, nil
}

// Initialize sets up the PostgreSQL connection with optimized settings
func (p *PostgresConnection) Initialize() error {
	var databaseURI string

	if os.Getenv("GO_ENV") != "production" {
		databaseURI = util.GodotEnv("DATABASE_URI_DEV")
	} else {
		databaseURI = os.Getenv("DATABASE_URI_PROD")
	}

	if databaseURI == "" {
		return nil
	}

	// Connect to PostgreSQL
	db, err := gorm.Open(postgres.Open(databaseURI), &gorm.Config{})
	if err != nil {
		logrus.Errorf("Connection to PostgreSQL failed: %v", err)
		return err
	}

	// Get the underlying SQL DB
	sqlDB, err := db.DB()
	if err != nil {
		logrus.Errorf("Failed to get database connection: %v", err)
		return err
	}

	// Get PostgreSQL configuration
	dbConfig := GetPostgresConfig()

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(dbConfig.MaxIdleConns)
	sqlDB.SetMaxOpenConns(dbConfig.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(dbConfig.ConnMaxLifetime)

	// Ping the database to verify connection
	if err := sqlDB.Ping(); err != nil {
		logrus.Errorf("Failed to ping PostgreSQL: %v", err)
		return err
	}

	p.DB = db

	// Initialize circuit breaker
	cbConfig := resilience.DefaultCircuitBreakerConfig("postgres")
	p.cb = resilience.NewCircuitBreaker(cbConfig)

	if os.Getenv("GO_ENV") != "production" {
		logrus.Info("Connection to PostgreSQL successful")
	}

	return nil
}

// Close closes the PostgreSQL connection
func (p *PostgresConnection) Close() error {
	if p.DB != nil {
		sqlDB, err := p.DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// ExecuteWithRetry executes a PostgreSQL operation with retry logic and circuit breaker
func (p *PostgresConnection) ExecuteWithRetry(operation func() (interface{}, error)) (interface{}, error) {
	return p.cb.Execute(nil, operation)
}

// GetDB returns the GORM DB instance
func (p *PostgresConnection) GetDB() *gorm.DB {
	return p.DB
}
