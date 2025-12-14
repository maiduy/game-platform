package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"
)

// ConfigSource defines the interface for configuration sources
type ConfigSource interface {
	// Load loads the configuration from the source
	Load() (map[string]interface{}, error)

	// Watch watches for changes in the configuration source
	Watch(notifyChan chan<- struct{})
}

// FileConfigSource implements ConfigSource for file-based configuration
type FileConfigSource struct {
	filePath      string
	lastModTime   time.Time
	checkInterval time.Duration
}

// NewFileConfigSource creates a new file-based configuration source
func NewFileConfigSource(filePath string, checkInterval time.Duration) *FileConfigSource {
	return &FileConfigSource{
		filePath:      filePath,
		checkInterval: checkInterval,
	}
}

// Load loads the configuration from the file
func (f *FileConfigSource) Load() (map[string]interface{}, error) {
	// Check if file exists
	info, err := os.Stat(f.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat config file: %w", err)
	}

	// Update last modified time
	f.lastModTime = info.ModTime()

	// Read file
	data, err := ioutil.ReadFile(f.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse JSON
	var config map[string]interface{}
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}

// Watch watches for changes in the configuration file
func (f *FileConfigSource) Watch(notifyChan chan<- struct{}) {
	ticker := time.NewTicker(f.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Check if file has been modified
			info, err := os.Stat(f.filePath)
			if err != nil {
				continue
			}

			// If file has been modified, notify
			if info.ModTime().After(f.lastModTime) {
				f.lastModTime = info.ModTime()
				notifyChan <- struct{}{}
			}
		}
	}
}

// DynamicConfig provides dynamic configuration with runtime refresh support
type DynamicConfig struct {
	source   ConfigSource
	config   map[string]interface{}
	mutex    sync.RWMutex
	onChange []func()
}

// NewDynamicConfig creates a new dynamic configuration
func NewDynamicConfig(source ConfigSource) (*DynamicConfig, error) {
	config, err := source.Load()
	if err != nil {
		return nil, err
	}

	dc := &DynamicConfig{
		source: source,
		config: config,
	}

	// Start watching for changes
	go dc.watchForChanges()

	return dc, nil
}

// watchForChanges watches for changes in the configuration source
func (dc *DynamicConfig) watchForChanges() {
	notifyChan := make(chan struct{})
	go dc.source.Watch(notifyChan)

	for range notifyChan {
		// Reload configuration
		config, err := dc.source.Load()
		if err != nil {
			continue
		}

		// Update configuration
		dc.mutex.Lock()
		dc.config = config
		dc.mutex.Unlock()

		// Notify listeners
		for _, fn := range dc.onChange {
			go fn()
		}
	}
}

// Get returns the value for the given key, supporting nested keys with dot notation
// For example: Get("security.routes.sme.auth")
func (dc *DynamicConfig) Get(key string) (interface{}, bool) {
	dc.mutex.RLock()
	defer dc.mutex.RUnlock()

	// Handle simple case (no dots)
	if !strings.Contains(key, ".") {
		value, ok := dc.config[key]
		return value, ok
	}

	// Handle nested keys
	keys := strings.Split(key, ".")
	var current interface{} = dc.config

	for i, k := range keys {
		// For the first key, check in the top-level config
		if i == 0 {
			value, ok := dc.config[k]
			if !ok {
				return nil, false
			}
			current = value
			continue
		}

		// For nested keys, check if the current value is a map
		currentMap, ok := current.(map[string]interface{})
		if !ok {
			return nil, false
		}

		// Get the value for the current key
		value, ok := currentMap[k]
		if !ok {
			return nil, false
		}
		current = value
	}

	return current, true
}

// GetString returns the string value for the given key
func (dc *DynamicConfig) GetString(key string) (string, bool) {
	value, ok := dc.Get(key)
	if !ok {
		return "", false
	}

	str, ok := value.(string)
	return str, ok
}

// GetInt returns the int value for the given key
func (dc *DynamicConfig) GetInt(key string) (int, bool) {
	value, ok := dc.Get(key)
	if !ok {
		return 0, false
	}

	// Handle different number types
	switch v := value.(type) {
	case int:
		return v, true
	case float64:
		return int(v), true
	default:
		return 0, false
	}
}

// GetBool returns the bool value for the given key
func (dc *DynamicConfig) GetBool(key string) (bool, bool) {
	value, ok := dc.Get(key)
	if !ok {
		return false, false
	}

	b, ok := value.(bool)
	return b, ok
}

// OnChange registers a callback function to be called when the configuration changes
func (dc *DynamicConfig) OnChange(fn func()) {
	dc.mutex.Lock()
	defer dc.mutex.Unlock()

	dc.onChange = append(dc.onChange, fn)
}
