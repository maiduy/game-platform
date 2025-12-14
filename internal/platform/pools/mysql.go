package pool

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"game-platform/internal/platform/config"
	util "game-platform/internal/platform/utils"
	"game-platform/pkg/resilience"

	_ "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
)

// MySQLConfig holds the configuration for MySQL connection
type MySQLConfig struct {
	// Connection pool settings
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration

	// Connection timeout settings
	Timeout      time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration

	// Query settings
	InterpolateParams bool
	MultiStatements   bool
	ParseTime         bool

	// TLS settings
	TLSConfig string
}

// GetMySQLConfig returns the MySQL configuration
func GetMySQLConfig() MySQLConfig {
	// Default values
	config := MySQLConfig{
		MaxIdleConns:      10,
		MaxOpenConns:      100,
		ConnMaxLifetime:   time.Hour,
		ConnMaxIdleTime:   time.Minute * 30,
		Timeout:           time.Second * 10,
		ReadTimeout:       time.Second * 30,
		WriteTimeout:      time.Second * 30,
		InterpolateParams: true,
		MultiStatements:   false,
		ParseTime:         true,
		TLSConfig:         "",
	}

	// Try to get values from dynamic config if available
	dynamicConfig, err := getMySQLConfigFromDynamic()
	if err == nil {
		if maxIdleConns, ok := dynamicConfig.GetInt("mysql.max_idle_conns"); ok {
			config.MaxIdleConns = maxIdleConns
		}

		if maxOpenConns, ok := dynamicConfig.GetInt("mysql.max_open_conns"); ok {
			config.MaxOpenConns = maxOpenConns
		}

		if connMaxLifetime, ok := dynamicConfig.GetInt("mysql.conn_max_lifetime"); ok {
			config.ConnMaxLifetime = time.Duration(connMaxLifetime) * time.Second
		}

		if connMaxIdleTime, ok := dynamicConfig.GetInt("mysql.conn_max_idle_time"); ok {
			config.ConnMaxIdleTime = time.Duration(connMaxIdleTime) * time.Second
		}

		if timeout, ok := dynamicConfig.GetInt("mysql.timeout"); ok {
			config.Timeout = time.Duration(timeout) * time.Second
		}

		if readTimeout, ok := dynamicConfig.GetInt("mysql.read_timeout"); ok {
			config.ReadTimeout = time.Duration(readTimeout) * time.Second
		}

		if writeTimeout, ok := dynamicConfig.GetInt("mysql.write_timeout"); ok {
			config.WriteTimeout = time.Duration(writeTimeout) * time.Second
		}

		if interpolateParams, ok := dynamicConfig.GetBool("mysql.interpolate_params"); ok {
			config.InterpolateParams = interpolateParams
		}

		if multiStatements, ok := dynamicConfig.GetBool("mysql.multi_statements"); ok {
			config.MultiStatements = multiStatements
		}

		if parseTime, ok := dynamicConfig.GetBool("mysql.parse_time"); ok {
			config.ParseTime = parseTime
		}

		if tlsConfig, ok := dynamicConfig.GetString("mysql.tls_config"); ok {
			config.TLSConfig = tlsConfig
		}
	} else {
		// Fallback to environment variables
		if maxIdleConns, err := strconv.Atoi(os.Getenv("MYSQL_MAX_IDLE_CONNS")); err == nil {
			config.MaxIdleConns = maxIdleConns
		}

		if maxOpenConns, err := strconv.Atoi(os.Getenv("MYSQL_MAX_OPEN_CONNS")); err == nil {
			config.MaxOpenConns = maxOpenConns
		}

		if connMaxLifetime, err := strconv.Atoi(os.Getenv("MYSQL_CONN_MAX_LIFETIME")); err == nil {
			config.ConnMaxLifetime = time.Duration(connMaxLifetime) * time.Second
		}

		if connMaxIdleTime, err := strconv.Atoi(os.Getenv("MYSQL_CONN_MAX_IDLE_TIME")); err == nil {
			config.ConnMaxIdleTime = time.Duration(connMaxIdleTime) * time.Second
		}

		if timeout, err := strconv.Atoi(os.Getenv("MYSQL_TIMEOUT")); err == nil {
			config.Timeout = time.Duration(timeout) * time.Second
		}

		if readTimeout, err := strconv.Atoi(os.Getenv("MYSQL_READ_TIMEOUT")); err == nil {
			config.ReadTimeout = time.Duration(readTimeout) * time.Second
		}

		if writeTimeout, err := strconv.Atoi(os.Getenv("MYSQL_WRITE_TIMEOUT")); err == nil {
			config.WriteTimeout = time.Duration(writeTimeout) * time.Second
		}

		if interpolateParams := os.Getenv("MYSQL_INTERPOLATE_PARAMS"); interpolateParams != "" {
			config.InterpolateParams = interpolateParams == "true"
		}

		if multiStatements := os.Getenv("MYSQL_MULTI_STATEMENTS"); multiStatements != "" {
			config.MultiStatements = multiStatements == "true"
		}

		if parseTime := os.Getenv("MYSQL_PARSE_TIME"); parseTime != "" {
			config.ParseTime = parseTime == "true"
		}

		if tlsConfig := os.Getenv("MYSQL_TLS_CONFIG"); tlsConfig != "" {
			config.TLSConfig = tlsConfig
		}
	}

	return config
}

// getMySQLConfigFromDynamic attempts to get the dynamic configuration for MySQL
func getMySQLConfigFromDynamic() (*config.DynamicConfig, error) {
	// Check if dynamic config is enabled
	if os.Getenv("ENABLE_DYNAMIC_CONFIG") != "true" {
		return nil, fmt.Errorf("dynamic config not enabled")
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

// MySQLConnection represents a connection to MySQL with resilience features
type MySQLConnection struct {
	DB *sql.DB
	cb *resilience.CircuitBreaker
}

// NewMySQLConnection creates and initializes a new MySQL connection
func NewMySQLConnection() (*MySQLConnection, error) {
	conn := &MySQLConnection{}
	if err := conn.Initialize(); err != nil {
		return nil, err
	}
	return conn, nil
}

// Initialize sets up the MySQL connection with optimized settings
func (m *MySQLConnection) Initialize() error {
	var dsn string

	// Get DSN from environment variables
	if os.Getenv("GO_ENV") != "production" {
		dsn = util.GodotEnv("MYSQL_URI_DEV")
	} else {
		dsn = os.Getenv("MYSQL_URI_PROD")
	}

	if dsn == "" {
		return fmt.Errorf("MySQL DSN not configured")
	}

	// Get MySQL configuration
	mysqlConfig := GetMySQLConfig()

	// Add configuration parameters to DSN if not already present
	if !strings.Contains(dsn, "timeout=") {
		dsn = fmt.Sprintf("%s&timeout=%ds", dsn, int(mysqlConfig.Timeout.Seconds()))
	}
	if !strings.Contains(dsn, "readTimeout=") {
		dsn = fmt.Sprintf("%s&readTimeout=%ds", dsn, int(mysqlConfig.ReadTimeout.Seconds()))
	}
	if !strings.Contains(dsn, "writeTimeout=") {
		dsn = fmt.Sprintf("%s&writeTimeout=%ds", dsn, int(mysqlConfig.WriteTimeout.Seconds()))
	}
	if !strings.Contains(dsn, "interpolateParams=") {
		dsn = fmt.Sprintf("%s&interpolateParams=%t", dsn, mysqlConfig.InterpolateParams)
	}
	if !strings.Contains(dsn, "multiStatements=") {
		dsn = fmt.Sprintf("%s&multiStatements=%t", dsn, mysqlConfig.MultiStatements)
	}
	if !strings.Contains(dsn, "parseTime=") {
		dsn = fmt.Sprintf("%s&parseTime=%t", dsn, mysqlConfig.ParseTime)
	}
	if mysqlConfig.TLSConfig != "" && !strings.Contains(dsn, "tls=") {
		dsn = fmt.Sprintf("%s&tls=%s", dsn, mysqlConfig.TLSConfig)
	}

	// Open connection to MySQL
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		logrus.Errorf("Failed to connect to MySQL: %v", err)
		return err
	}

	// Configure connection pooling
	db.SetMaxIdleConns(mysqlConfig.MaxIdleConns)
	db.SetMaxOpenConns(mysqlConfig.MaxOpenConns)
	db.SetConnMaxLifetime(mysqlConfig.ConnMaxLifetime)
	db.SetConnMaxIdleTime(mysqlConfig.ConnMaxIdleTime)

	// Verify connection
	if err := db.Ping(); err != nil {
		logrus.Errorf("Failed to ping MySQL: %v", err)
		return err
	}

	m.DB = db

	// Initialize circuit breaker
	cbConfig := resilience.DefaultCircuitBreakerConfig("mysql")
	m.cb = resilience.NewCircuitBreaker(cbConfig)

	if os.Getenv("GO_ENV") != "production" {
		logrus.Info("Connected to MySQL successfully")
	}

	return nil
}

// Close closes the MySQL connection
func (m *MySQLConnection) Close() error {
	if m.DB != nil {
		return m.DB.Close()
	}
	return nil
}

// ExecuteWithRetry executes a MySQL operation with retry logic and circuit breaker
func (m *MySQLConnection) ExecuteWithRetry(operation func() (interface{}, error)) (interface{}, error) {
	return m.cb.Execute(nil, operation)
}

// Transaction starts a new transaction with the given options
func (m *MySQLConnection) Transaction(options *sql.TxOptions) (*sql.Tx, error) {
	result, err := m.ExecuteWithRetry(func() (interface{}, error) {
		return m.DB.BeginTx(context.Background(), options)
	})

	if err != nil {
		return nil, err
	}

	return result.(*sql.Tx), nil
}

// Query executes a query that returns rows with resilience patterns
func (m *MySQLConnection) Query(query string, args ...interface{}) (*sql.Rows, error) {
	result, err := m.ExecuteWithRetry(func() (interface{}, error) {
		return m.DB.Query(query, args...)
	})

	if err != nil {
		return nil, err
	}

	return result.(*sql.Rows), nil
}

// QueryRow executes a query that is expected to return at most one row
func (m *MySQLConnection) QueryRow(query string, args ...interface{}) *sql.Row {
	return m.DB.QueryRow(query, args...)
}

// Exec executes a query without returning any rows with resilience patterns
func (m *MySQLConnection) Exec(query string, args ...interface{}) (sql.Result, error) {
	result, err := m.ExecuteWithRetry(func() (interface{}, error) {
		return m.DB.Exec(query, args...)
	})

	if err != nil {
		return nil, err
	}

	return result.(sql.Result), nil
}
