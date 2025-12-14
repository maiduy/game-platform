package pool

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"game-platform/internal/platform/config"
	"game-platform/pkg/resilience"

	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
)

// NATSConfig holds the configuration for NATS connections
type NATSConfig struct {
	MaxReconnects      int
	ReconnectWait      time.Duration
	ReconnectBufSize   int
	Timeout            time.Duration
	PingInterval       time.Duration
	MaxPingsOut        int
	DrainTimeout       time.Duration
	StreamMaxAge       time.Duration
	StreamMaxBytes     int64
	StreamReplicas     int
	ConsumerMaxDeliver int
	ConsumerAckWait    time.Duration
}

// GetNATSConfig returns the NATS configuration
func GetNATSConfig() NATSConfig {
	config := NATSConfig{
		MaxReconnects:      100,
		ReconnectWait:      1 * time.Second,
		ReconnectBufSize:   8 * 1024 * 1024, // 8MB
		Timeout:            10 * time.Second,
		PingInterval:       10 * time.Second,
		MaxPingsOut:        3,
		DrainTimeout:       30 * time.Second,
		StreamMaxAge:       168 * time.Hour,         // 7 days
		StreamMaxBytes:     10 * 1024 * 1024 * 1024, // 10GB
		StreamReplicas:     3,
		ConsumerMaxDeliver: 5,
		ConsumerAckWait:    30 * time.Second,
	}

	dynamicConfig, err := getNATSConfigFromDynamic()
	if err == nil {
		if maxReconnects, ok := dynamicConfig.GetInt("nats.max_reconnects"); ok {
			config.MaxReconnects = maxReconnects
		}
		if reconnectWait, ok := dynamicConfig.GetInt("nats.reconnect_wait"); ok {
			config.ReconnectWait = time.Duration(reconnectWait) * time.Second
		}
		if reconnectBufSize, ok := dynamicConfig.GetInt("nats.reconnect_buf_size"); ok {
			config.ReconnectBufSize = reconnectBufSize
		}
		if timeout, ok := dynamicConfig.GetInt("nats.timeout"); ok {
			config.Timeout = time.Duration(timeout) * time.Second
		}
		if pingInterval, ok := dynamicConfig.GetInt("nats.ping_interval"); ok {
			config.PingInterval = time.Duration(pingInterval) * time.Second
		}
		if maxPingsOut, ok := dynamicConfig.GetInt("nats.max_pings_out"); ok {
			config.MaxPingsOut = maxPingsOut
		}
		if drainTimeout, ok := dynamicConfig.GetInt("nats.drain_timeout"); ok {
			config.DrainTimeout = time.Duration(drainTimeout) * time.Second
		}
		if streamMaxAge, ok := dynamicConfig.GetInt("nats.stream_max_age"); ok {
			config.StreamMaxAge = time.Duration(streamMaxAge) * time.Hour
		}
		if streamMaxBytes, ok := dynamicConfig.GetInt("nats.stream_max_bytes"); ok {
			config.StreamMaxBytes = int64(streamMaxBytes)
		}
		if streamReplicas, ok := dynamicConfig.GetInt("nats.stream_replicas"); ok {
			config.StreamReplicas = streamReplicas
		}
		if consumerMaxDeliver, ok := dynamicConfig.GetInt("nats.consumer_max_deliver"); ok {
			config.ConsumerMaxDeliver = consumerMaxDeliver
		}
		if consumerAckWait, ok := dynamicConfig.GetInt("nats.consumer_ack_wait"); ok {
			config.ConsumerAckWait = time.Duration(consumerAckWait) * time.Second
		}
	} else {
		if maxReconnects, err := strconv.Atoi(os.Getenv("NATS_MAX_RECONNECTS")); err == nil {
			config.MaxReconnects = maxReconnects
		}
		if reconnectWait, err := strconv.Atoi(os.Getenv("NATS_RECONNECT_WAIT")); err == nil {
			config.ReconnectWait = time.Duration(reconnectWait) * time.Second
		}
		if reconnectBufSize, err := strconv.Atoi(os.Getenv("NATS_RECONNECT_BUF_SIZE")); err == nil {
			config.ReconnectBufSize = reconnectBufSize
		}
		if timeout, err := strconv.Atoi(os.Getenv("NATS_TIMEOUT")); err == nil {
			config.Timeout = time.Duration(timeout) * time.Second
		}
		if pingInterval, err := strconv.Atoi(os.Getenv("NATS_PING_INTERVAL")); err == nil {
			config.PingInterval = time.Duration(pingInterval) * time.Second
		}
		if maxPingsOut, err := strconv.Atoi(os.Getenv("NATS_MAX_PINGS_OUT")); err == nil {
			config.MaxPingsOut = maxPingsOut
		}
		if drainTimeout, err := strconv.Atoi(os.Getenv("NATS_DRAIN_TIMEOUT")); err == nil {
			config.DrainTimeout = time.Duration(drainTimeout) * time.Second
		}
	}

	return config
}

func getNATSConfigFromDynamic() (*config.DynamicConfig, error) {
	if os.Getenv("ENABLE_DYNAMIC_CONFIG") != "true" {
		return nil, fmt.Errorf("dynamic config not enabled")
	}

	configFilePath := os.Getenv("CONFIG_FILE_PATH")
	if configFilePath == "" {
		configFilePath = "configs/app_config.json"
	}

	configSource := config.NewFileConfigSource(configFilePath, 30*time.Second)
	return config.NewDynamicConfig(configSource)
}

// NATSConnection represents a NATS connection with JetStream support
type NATSConnection struct {
	Conn      *nats.Conn
	JetStream nats.JetStreamContext
	cb        *resilience.CircuitBreaker
	config    NATSConfig
}

var (
	natsInstance *NATSConnection
	natsOnce     sync.Once
)

// NewNATSConnection creates and initializes a new NATS connection
func NewNATSConnection() (*NATSConnection, error) {
	var initErr error
	natsOnce.Do(func() {
		conn := &NATSConnection{}
		if err := conn.Initialize(); err != nil {
			initErr = err
			return
		}
		natsInstance = conn
	})

	if initErr != nil {
		return nil, initErr
	}

	if natsInstance == nil {
		return nil, fmt.Errorf("failed to initialize NATS connection")
	}

	return natsInstance, nil
}

// Initialize sets up the NATS connection with JetStream
func (n *NATSConnection) Initialize() error {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = nats.DefaultURL
	}

	n.config = GetNATSConfig()

	opts := []nats.Option{
		nats.Name("micro-backend-service"),
		nats.MaxReconnects(n.config.MaxReconnects),
		nats.ReconnectWait(n.config.ReconnectWait),
		nats.ReconnectBufSize(n.config.ReconnectBufSize),
		nats.Timeout(n.config.Timeout),
		nats.PingInterval(n.config.PingInterval),
		nats.MaxPingsOutstanding(n.config.MaxPingsOut),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			logrus.Warnf("NATS disconnected: %v", err)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			logrus.Info("NATS reconnected")
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			logrus.Warn("NATS connection closed")
		}),
	}

	conn, err := nats.Connect(natsURL, opts...)
	if err != nil {
		logrus.Errorf("Failed to connect to NATS: %v", err)
		return err
	}

	js, err := conn.JetStream()
	if err != nil {
		logrus.Errorf("Failed to create JetStream context: %v", err)
		conn.Close()
		return err
	}

	n.Conn = conn
	n.JetStream = js

	cbConfig := resilience.DefaultCircuitBreakerConfig("nats-jetstream")
	n.cb = resilience.NewCircuitBreaker(cbConfig)

	logrus.Infof("Connected to NATS at %s with JetStream enabled", natsURL)
	return nil
}

// Close closes the NATS connection
func (n *NATSConnection) Close() error {
	if n.Conn != nil {
		ctx, cancel := context.WithTimeout(context.Background(), n.config.DrainTimeout)
		defer cancel()

		done := make(chan struct{})
		go func() {
			n.Conn.Drain()
			close(done)
		}()

		select {
		case <-done:
			logrus.Info("NATS connection drained successfully")
		case <-ctx.Done():
			logrus.Warn("NATS drain timeout, forcing close")
			n.Conn.Close()
		}
	}
	return nil
}

// ExecuteWithRetry executes a NATS operation with circuit breaker
func (n *NATSConnection) ExecuteWithRetry(operation func() (interface{}, error)) (interface{}, error) {
	return n.cb.Execute(context.Background(), operation)
}

// IsConnected checks if the NATS connection is active
func (n *NATSConnection) IsConnected() bool {
	return n.Conn != nil && n.Conn.IsConnected()
}
