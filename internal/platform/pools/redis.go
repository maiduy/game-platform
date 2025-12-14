package pool

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"game-platform/internal/platform/config"
	"game-platform/pkg/resilience"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

var ctx = context.Background()

// RedisInterface defines the common interface for Redis connections
type RedisInterface interface {
	SetValue(key, field, value string)
	GetValue(key, field string) string
	DelValue(key, field string)
	Close()
}

// RedisConfig holds the configuration for Redis connections
type RedisConfig struct {
	MaxRetries      int
	MinRetryBackoff time.Duration
	MaxRetryBackoff time.Duration
	DialTimeout     time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	PoolSize        int
	MinIdleConns    int
	PoolTimeout     time.Duration
}

// GetRedisConfig returns the Redis configuration
func GetRedisConfig() RedisConfig {
	// Default values
	config := RedisConfig{
		MaxRetries:      3,
		MinRetryBackoff: 8 * time.Millisecond,
		MaxRetryBackoff: 512 * time.Millisecond,
		DialTimeout:     5 * time.Second,
		ReadTimeout:     3 * time.Second,
		WriteTimeout:    3 * time.Second,
		PoolSize:        10,
		MinIdleConns:    5,
		PoolTimeout:     4 * time.Second,
	}

	// Try to get values from dynamic config if available
	dynamicConfig, err := getRedisConfigFromDynamic()
	if err == nil {
		if maxRetries, ok := dynamicConfig.GetInt("redis.max_retries"); ok {
			config.MaxRetries = maxRetries
		}

		if minRetryBackoff, ok := dynamicConfig.GetInt("redis.min_retry_backoff"); ok {
			config.MinRetryBackoff = time.Duration(minRetryBackoff) * time.Millisecond
		}

		if maxRetryBackoff, ok := dynamicConfig.GetInt("redis.max_retry_backoff"); ok {
			config.MaxRetryBackoff = time.Duration(maxRetryBackoff) * time.Millisecond
		}

		if dialTimeout, ok := dynamicConfig.GetInt("redis.dial_timeout"); ok {
			config.DialTimeout = time.Duration(dialTimeout) * time.Second
		}

		if readTimeout, ok := dynamicConfig.GetInt("redis.read_timeout"); ok {
			config.ReadTimeout = time.Duration(readTimeout) * time.Second
		}

		if writeTimeout, ok := dynamicConfig.GetInt("redis.write_timeout"); ok {
			config.WriteTimeout = time.Duration(writeTimeout) * time.Second
		}

		if poolSize, ok := dynamicConfig.GetInt("redis.pool_size"); ok {
			config.PoolSize = poolSize
		}

		if minIdleConns, ok := dynamicConfig.GetInt("redis.min_idle_conns"); ok {
			config.MinIdleConns = minIdleConns
		}

		if poolTimeout, ok := dynamicConfig.GetInt("redis.pool_timeout"); ok {
			config.PoolTimeout = time.Duration(poolTimeout) * time.Second
		}
	} else {
		// Fallback to environment variables
		if maxRetries, err := strconv.Atoi(os.Getenv("REDIS_MAX_RETRIES")); err == nil {
			config.MaxRetries = maxRetries
		}

		if minRetryBackoff, err := strconv.Atoi(os.Getenv("REDIS_MIN_RETRY_BACKOFF")); err == nil {
			config.MinRetryBackoff = time.Duration(minRetryBackoff) * time.Millisecond
		}

		if maxRetryBackoff, err := strconv.Atoi(os.Getenv("REDIS_MAX_RETRY_BACKOFF")); err == nil {
			config.MaxRetryBackoff = time.Duration(maxRetryBackoff) * time.Millisecond
		}

		if dialTimeout, err := strconv.Atoi(os.Getenv("REDIS_DIAL_TIMEOUT")); err == nil {
			config.DialTimeout = time.Duration(dialTimeout) * time.Second
		}

		if readTimeout, err := strconv.Atoi(os.Getenv("REDIS_READ_TIMEOUT")); err == nil {
			config.ReadTimeout = time.Duration(readTimeout) * time.Second
		}

		if writeTimeout, err := strconv.Atoi(os.Getenv("REDIS_WRITE_TIMEOUT")); err == nil {
			config.WriteTimeout = time.Duration(writeTimeout) * time.Second
		}

		if poolSize, err := strconv.Atoi(os.Getenv("REDIS_POOL_SIZE")); err == nil {
			config.PoolSize = poolSize
		}

		if minIdleConns, err := strconv.Atoi(os.Getenv("REDIS_MIN_IDLE_CONNS")); err == nil {
			config.MinIdleConns = minIdleConns
		}

		if poolTimeout, err := strconv.Atoi(os.Getenv("REDIS_POOL_TIMEOUT")); err == nil {
			config.PoolTimeout = time.Duration(poolTimeout) * time.Second
		}
	}

	return config
}

// getRedisConfigFromDynamic attempts to get the dynamic configuration for Redis
func getRedisConfigFromDynamic() (*config.DynamicConfig, error) {
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

// -----------------------
// Redis Sentinel
// -----------------------
type RedisConnection struct {
	RDB *redis.Client
	cb  *resilience.CircuitBreaker
}

// NewRedisConnection creates and initializes a new Redis connection
func NewRedisConnection() (*RedisConnection, error) {
	conn := &RedisConnection{}
	if err := conn.Initialize(); err != nil {
		return nil, err
	}
	return conn, nil
}

func (r *RedisConnection) Initialize() error {
	hosts := os.Getenv("REDIS_HOST")
	port := os.Getenv("REDIS_PORT")
	password := os.Getenv("REDIS_PASSWORD")
	master := os.Getenv("REDIS_SERVICE_NAME")

	if hosts == "" || port == "" || master == "" {
		logrus.Error("Missing Redis Sentinel environment variables")
		return fmt.Errorf("missing Redis Sentinel environment variables")
	}

	sentinels := buildSentinelAddrs(hosts, port)

	// Get Redis configuration
	redisConfig := GetRedisConfig()

	client := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:      master,
		SentinelAddrs:   sentinels,
		Password:        password,
		MaxRetries:      redisConfig.MaxRetries,
		MinRetryBackoff: redisConfig.MinRetryBackoff,
		MaxRetryBackoff: redisConfig.MaxRetryBackoff,
		DialTimeout:     redisConfig.DialTimeout,
		ReadTimeout:     redisConfig.ReadTimeout,
		WriteTimeout:    redisConfig.WriteTimeout,
		PoolSize:        redisConfig.PoolSize,
		MinIdleConns:    redisConfig.MinIdleConns,
		PoolTimeout:     redisConfig.PoolTimeout,
	})

	if err := pingRedis(client); err != nil {
		logrus.WithError(err).Error("Redis Sentinel ping failed")
		return err
	}

	r.RDB = client

	// Initialize circuit breaker
	cbConfig := resilience.DefaultCircuitBreakerConfig("redis-sentinel")
	r.cb = resilience.NewCircuitBreaker(cbConfig)

	return nil
}

func (r *RedisConnection) SetValue(key, field, value string) {
	_, err := r.cb.Execute(ctx, func() (interface{}, error) {
		return r.RDB.HSet(ctx, key, field, value).Result()
	})

	if err != nil {
		handleRedisError("HSet", key, field, err)
	}
}

func (r *RedisConnection) GetValue(key, field string) string {
	result, err := r.cb.Execute(ctx, func() (interface{}, error) {
		return r.RDB.HGet(ctx, key, field).Result()
	})

	if err != nil {
		handleRedisError("HGet", key, field, err)
		return ""
	}

	if str, ok := result.(string); ok {
		return str
	}

	return ""
}

func (r *RedisConnection) DelValue(key, field string) {
	_, err := r.cb.Execute(ctx, func() (interface{}, error) {
		return r.RDB.HDel(ctx, key, field).Result()
	})

	if err != nil {
		handleRedisError("HDel", key, field, err)
	}
}

func (r *RedisConnection) Close() {
	if r.RDB != nil {
		_ = r.RDB.Close()
	}
}

// -----------------------
// Redis Cluster
// -----------------------
type RedisClusterConnection struct {
	RDB *redis.ClusterClient
	cb  *resilience.CircuitBreaker
}

// NewRedisClusterConnection creates and initializes a new Redis cluster connection
func NewRedisClusterConnection() (*RedisClusterConnection, error) {
	conn := &RedisClusterConnection{}
	if err := conn.Initialize(); err != nil {
		return nil, err
	}
	return conn, nil
}

func (r *RedisClusterConnection) Initialize() error {
	hosts := os.Getenv("REDIS_HOST")
	password := os.Getenv("REDIS_PASSWORD")

	if hosts == "" {
		logrus.Error("Missing Redis Cluster environment variables")
		return fmt.Errorf("missing Redis Cluster environment variables")
	}

	// Validate and clean addresses
	var validAddrs []string
	clusterAddrs := strings.Split(hosts, ",")
	for _, addr := range clusterAddrs {
		addr = strings.TrimSpace(addr)
		// Ensure address has the correct format (host:port)
		parts := strings.Split(addr, ":")
		if len(parts) == 2 {
			validAddrs = append(validAddrs, addr)
		} else {
			logrus.Warnf("Invalid Redis address format: %s", addr)
		}
	}

	if len(validAddrs) == 0 {
		logrus.Error("No valid Redis addresses found")
		return fmt.Errorf("no valid Redis addresses found")
	}

	logrus.Infof("Initializing Redis Cluster with hosts: %s", strings.Join(validAddrs, ","))

	// Get Redis configuration
	redisConfig := GetRedisConfig()

	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:           validAddrs,
		Password:        password,
		MaxRetries:      redisConfig.MaxRetries,
		MinRetryBackoff: redisConfig.MinRetryBackoff,
		MaxRetryBackoff: redisConfig.MaxRetryBackoff,
		DialTimeout:     redisConfig.DialTimeout,
		ReadTimeout:     redisConfig.ReadTimeout,
		WriteTimeout:    redisConfig.WriteTimeout,
		PoolSize:        redisConfig.PoolSize,
		MinIdleConns:    redisConfig.MinIdleConns,
		PoolTimeout:     redisConfig.PoolTimeout,
	})

	if err := pingRedis(client); err != nil {
		logrus.WithError(err).Error("Redis Cluster ping failed")
		return err
	}

	r.RDB = client

	// Initialize circuit breaker
	cbConfig := resilience.DefaultCircuitBreakerConfig("redis-cluster")
	r.cb = resilience.NewCircuitBreaker(cbConfig)

	return nil
}

func (r *RedisClusterConnection) SetValue(key, field, value string) {
	_, err := r.cb.Execute(ctx, func() (interface{}, error) {
		return r.RDB.HSet(ctx, key, field, value).Result()
	})

	if err != nil {
		handleRedisError("HSet", key, field, err)
	}
}

func (r *RedisClusterConnection) GetValue(key, field string) string {
	result, err := r.cb.Execute(ctx, func() (interface{}, error) {
		return r.RDB.HGet(ctx, key, field).Result()
	})

	if err != nil {
		handleRedisError("HGet", key, field, err)
		return ""
	}

	if str, ok := result.(string); ok {
		return str
	}

	return ""
}

func (r *RedisClusterConnection) DelValue(key, field string) {
	_, err := r.cb.Execute(ctx, func() (interface{}, error) {
		return r.RDB.HDel(ctx, key, field).Result()
	})

	if err != nil {
		handleRedisError("HDel", key, field, err)
	}
}

func (r *RedisClusterConnection) Close() {
	if r.RDB != nil {
		_ = r.RDB.Close()
	}
}

// -----------------------
// Shared Utility Functions
// -----------------------
func buildSentinelAddrs(hosts, port string) []string {
	var addrs []string
	for _, host := range strings.Split(hosts, ",") {
		addrs = append(addrs, fmt.Sprintf("%s:%s", host, port))
	}
	return addrs
}

func pingRedis(client redis.UniversalClient) error {
	pong, err := client.Ping(ctx).Result()
	logrus.Infof("Redis ping response: %s", pong)
	return err
}

func handleRedisError(op, key, field string, err error) {
	if err != nil {
		logrus.Errorf("%s key=%s field=%s error=%s", op, key, field, err)
	}
}
