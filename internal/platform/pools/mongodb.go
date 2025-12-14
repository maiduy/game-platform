package pool

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"game-platform/internal/platform/config"
	util "game-platform/internal/platform/utils"
	"game-platform/pkg/resilience"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

// MongoDBConfig holds the configuration for MongoDB connection
type MongoDBConfig struct {
	// Connection pool settings
	MaxPoolSize     uint64
	MinPoolSize     uint64
	MaxConnIdleTime time.Duration

	// Connection timeout settings
	ConnectTimeout         time.Duration
	SocketTimeout          time.Duration
	ServerSelectionTimeout time.Duration

	// Write concern settings
	WriteConcern string

	// Read preference settings
	ReadPreference string

	// Retry settings
	MaxRetries    int
	RetryInterval time.Duration
}

// MongoDBConnectionParams holds the parameters for a MongoDB connection
type MongoDBConnectionParams struct {
	// Connection URI
	URI string

	// Database name
	DatabaseName string

	// Circuit breaker name
	CircuitBreakerName string

	// Connection config
	Config MongoDBConfig
}

// MongoDBConnection represents a connection to MongoDB with resilience features
type MongoDBConnection struct {
	Client   *mongo.Client
	cb       *resilience.CircuitBreaker
	Database string // Store the database name

	// Connection parameters
	params MongoDBConnectionParams
}

// Singleton instances for different MongoDB connections
var (
	defaultConnection *MongoDBConnection
	gdoConnection     *MongoDBConnection

	defaultOnce sync.Once
	gdoOnce     sync.Once
)

// GetMongoDBConfig returns the MongoDB configuration from environment or dynamic config
func GetMongoDBConfig() MongoDBConfig {
	// Default values
	config := MongoDBConfig{
		MaxPoolSize:            100,
		MinPoolSize:            10,
		MaxConnIdleTime:        30 * time.Minute,
		ConnectTimeout:         10 * time.Second,
		SocketTimeout:          30 * time.Second,
		ServerSelectionTimeout: 30 * time.Second,
		WriteConcern:           "majority",
		ReadPreference:         "primaryPreferred",
		MaxRetries:             3,
		RetryInterval:          500 * time.Millisecond,
	}

	// Try to get values from dynamic config if available
	dynamicConfig, err := getMongoDBConfigFromDynamic()
	if err == nil {
		if maxPoolSize, ok := dynamicConfig.GetInt("mongodb.max_pool_size"); ok {
			config.MaxPoolSize = uint64(maxPoolSize)
		}

		if minPoolSize, ok := dynamicConfig.GetInt("mongodb.min_pool_size"); ok {
			config.MinPoolSize = uint64(minPoolSize)
		}

		if maxConnIdleTime, ok := dynamicConfig.GetInt("mongodb.max_conn_idle_time"); ok {
			config.MaxConnIdleTime = time.Duration(maxConnIdleTime) * time.Second
		}

		if connectTimeout, ok := dynamicConfig.GetInt("mongodb.connect_timeout"); ok {
			config.ConnectTimeout = time.Duration(connectTimeout) * time.Second
		}

		if socketTimeout, ok := dynamicConfig.GetInt("mongodb.socket_timeout"); ok {
			config.SocketTimeout = time.Duration(socketTimeout) * time.Second
		}

		if serverSelectionTimeout, ok := dynamicConfig.GetInt("mongodb.server_selection_timeout"); ok {
			config.ServerSelectionTimeout = time.Duration(serverSelectionTimeout) * time.Second
		}

		if writeConcern, ok := dynamicConfig.GetString("mongodb.write_concern"); ok {
			config.WriteConcern = writeConcern
		}

		if readPreference, ok := dynamicConfig.GetString("mongodb.read_preference"); ok {
			config.ReadPreference = readPreference
		}

		if maxRetries, ok := dynamicConfig.GetInt("mongodb.max_retries"); ok {
			config.MaxRetries = maxRetries
		}

		if retryInterval, ok := dynamicConfig.GetInt("mongodb.retry_interval"); ok {
			config.RetryInterval = time.Duration(retryInterval) * time.Millisecond
		}
	} else {
		// Fallback to environment variables
		if maxPoolSize, err := strconv.ParseUint(os.Getenv("MONGODB_MAX_POOL_SIZE"), 10, 64); err == nil {
			config.MaxPoolSize = maxPoolSize
		}

		if minPoolSize, err := strconv.ParseUint(os.Getenv("MONGODB_MIN_POOL_SIZE"), 10, 64); err == nil {
			config.MinPoolSize = minPoolSize
		}

		if maxConnIdleTime, err := strconv.Atoi(os.Getenv("MONGODB_MAX_CONN_IDLE_TIME")); err == nil {
			config.MaxConnIdleTime = time.Duration(maxConnIdleTime) * time.Second
		}

		if connectTimeout, err := strconv.Atoi(os.Getenv("MONGODB_CONNECT_TIMEOUT")); err == nil {
			config.ConnectTimeout = time.Duration(connectTimeout) * time.Second
		}

		if socketTimeout, err := strconv.Atoi(os.Getenv("MONGODB_SOCKET_TIMEOUT")); err == nil {
			config.SocketTimeout = time.Duration(socketTimeout) * time.Second
		}

		if serverSelectionTimeout, err := strconv.Atoi(os.Getenv("MONGODB_SERVER_SELECTION_TIMEOUT")); err == nil {
			config.ServerSelectionTimeout = time.Duration(serverSelectionTimeout) * time.Second
		}

		if writeConcern := os.Getenv("MONGODB_WRITE_CONCERN"); writeConcern != "" {
			config.WriteConcern = writeConcern
		}

		if readPreference := os.Getenv("MONGODB_READ_PREFERENCE"); readPreference != "" {
			config.ReadPreference = readPreference
		}

		if maxRetries, err := strconv.Atoi(os.Getenv("MONGODB_MAX_RETRIES")); err == nil {
			config.MaxRetries = maxRetries
		}

		if retryInterval, err := strconv.Atoi(os.Getenv("MONGODB_RETRY_INTERVAL")); err == nil {
			config.RetryInterval = time.Duration(retryInterval) * time.Millisecond
		}
	}

	return config
}

// getMongoDBConfigFromDynamic attempts to get the dynamic configuration for MongoDB
func getMongoDBConfigFromDynamic() (*config.DynamicConfig, error) {
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

// NewMongoDBConnection creates and initializes a new default MongoDB connection
func NewMongoDBConnection() (*MongoDBConnection, error) {
	defaultOnce.Do(func() {
		// Get MongoDB URI from environment
		var mongoURI string
		if os.Getenv("GO_ENV") != "production" {
			mongoURI = util.GodotEnv("MONGODB_URI_DEV")
		} else {
			mongoURI = os.Getenv("MONGODB_URI_PROD")
		}

		// Get MongoDB database name
		dbName := os.Getenv("MONGODB_DATABASE")
		if dbName == "" {
			dbName = "vng" // Default database name
		}

		// Get MongoDB configuration
		mongoConfig := GetMongoDBConfig()

		// Create connection parameters
		params := MongoDBConnectionParams{
			URI:                mongoURI,
			DatabaseName:       dbName,
			CircuitBreakerName: "mongodb",
			Config:             mongoConfig,
		}

		// Initialize connection
		conn := &MongoDBConnection{}
		err := conn.Initialize(params)
		if err != nil {
			logrus.Errorf("Failed to initialize default MongoDB connection: %v", err)
			return
		}

		defaultConnection = conn
	})

	if defaultConnection == nil {
		return nil, fmt.Errorf("failed to initialize default MongoDB connection")
	}

	return defaultConnection, nil
}

// NewGDOMongoDBConnection creates and initializes a new MongoDB connection for GDO
func NewGDOMongoDBConnection() (*MongoDBConnection, error) {
	gdoOnce.Do(func() {
		// Get MongoDB URI from environment
		var mongoURI string
		if os.Getenv("GO_ENV") != "production" {
			mongoURI = util.GodotEnv("MONGODB_URI_GDO_DEV")
		} else {
			mongoURI = os.Getenv("MONGODB_URI_GDO_PROD")
		}

		// Get MongoDB database name
		dbName := os.Getenv("MONGODB_DATABASE_GDO")
		if dbName == "" {
			dbName = "gdo" // Default database name
		}

		// Get MongoDB configuration
		mongoConfig := GetMongoDBConfig()

		// Create connection parameters
		params := MongoDBConnectionParams{
			URI:                mongoURI,
			DatabaseName:       dbName,
			CircuitBreakerName: "mongodb-gdo",
			Config:             mongoConfig,
		}

		// Initialize connection
		conn := &MongoDBConnection{}
		err := conn.Initialize(params)
		if err != nil {
			logrus.Errorf("Failed to initialize GDO MongoDB connection: %v", err)
			return
		}

		gdoConnection = conn
	})

	if gdoConnection == nil {
		return nil, fmt.Errorf("failed to initialize GDO MongoDB connection")
	}

	return gdoConnection, nil
}

// NewCustomMongoDBConnection creates a new MongoDB connection with custom parameters
func NewCustomMongoDBConnection(params MongoDBConnectionParams) (*MongoDBConnection, error) {
	conn := &MongoDBConnection{}
	if err := conn.Initialize(params); err != nil {
		return nil, err
	}
	return conn, nil
}

// Initialize sets up the MongoDB connection with the provided parameters
func (m *MongoDBConnection) Initialize(params MongoDBConnectionParams) error {
	// Store the parameters
	m.params = params

	// Store the database name
	m.Database = params.DatabaseName

	if params.URI == "" {
		logrus.Warn("MongoDB URI is empty, skipping connection initialization")
		return nil
	}

	// Set up client options
	clientOptions := options.Client().
		ApplyURI(params.URI).
		SetMaxPoolSize(params.Config.MaxPoolSize).
		SetMinPoolSize(params.Config.MinPoolSize).
		SetMaxConnIdleTime(params.Config.MaxConnIdleTime).
		SetConnectTimeout(params.Config.ConnectTimeout).
		SetSocketTimeout(params.Config.SocketTimeout).
		SetServerSelectionTimeout(params.Config.ServerSelectionTimeout)

	// Set read preference
	readPref, err := getReadPreference(params.Config.ReadPreference)
	if err == nil && readPref != nil {
		clientOptions.SetReadPreference(readPref)
	}

	// Set write concern
	writeConcern, err := getWriteConcern(params.Config.WriteConcern)
	if err == nil && writeConcern != nil {
		clientOptions.SetWriteConcern(writeConcern)
	}

	// Set retry writes for replica set operations
	clientOptions.SetRetryWrites(true)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), params.Config.ConnectTimeout)
	defer cancel()

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		logrus.Errorf("Failed to connect to MongoDB: %v", err)
		return err
	}

	// Ping the MongoDB server to verify connection
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		logrus.Errorf("Failed to ping MongoDB: %v", err)
		return err
	}

	m.Client = client

	// Initialize circuit breaker
	cbConfig := resilience.DefaultCircuitBreakerConfig(params.CircuitBreakerName)
	m.cb = resilience.NewCircuitBreaker(cbConfig)

	logrus.Infof("Connected to MongoDB successfully: (database: %s)", params.DatabaseName)
	return nil
}

// Close closes the MongoDB connection
func (m *MongoDBConnection) Close() error {
	if m.Client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return m.Client.Disconnect(ctx)
	}
	return nil
}

// ExecuteWithRetry executes a MongoDB operation with retry logic and circuit breaker
func (m *MongoDBConnection) ExecuteWithRetry(operation func() (interface{}, error)) (interface{}, error) {
	return m.cb.Execute(context.Background(), operation)
}

// GetDatabase returns a database with the given name
func (m *MongoDBConnection) GetDatabase(name string) *mongo.Database {
	return m.Client.Database(name)
}

// GetDefaultDatabase returns the default database for this connection
func (m *MongoDBConnection) GetDefaultDatabase() *mongo.Database {
	return m.Client.Database(m.Database)
}

// GetCollection returns a collection from the specified database
func (m *MongoDBConnection) GetCollection(dbName, collName string) *mongo.Collection {
	return m.Client.Database(dbName).Collection(collName)
}

// GetDefaultCollection returns a collection from the default database
func (m *MongoDBConnection) GetDefaultCollection(collName string) *mongo.Collection {
	return m.Client.Database(m.Database).Collection(collName)
}

// Helper functions

// getReadPreference returns a readpref.ReadPref based on the string value
func getReadPreference(pref string) (*readpref.ReadPref, error) {
	switch pref {
	case "primary":
		return readpref.Primary(), nil
	case "primaryPreferred":
		return readpref.PrimaryPreferred(), nil
	case "secondary":
		return readpref.Secondary(), nil
	case "secondaryPreferred":
		return readpref.SecondaryPreferred(), nil
	case "nearest":
		return readpref.Nearest(), nil
	default:
		return readpref.PrimaryPreferred(), nil
	}
}

// getWriteConcern returns a write concern based on the string value
func getWriteConcern(concern string) (*writeconcern.WriteConcern, error) {
	switch concern {
	case "majority":
		return writeconcern.New(writeconcern.WMajority()), nil
	case "1":
		return writeconcern.New(writeconcern.W(1)), nil
	case "0":
		return writeconcern.New(writeconcern.W(0)), nil
	default:
		return writeconcern.New(writeconcern.WMajority()), nil
	}
}
