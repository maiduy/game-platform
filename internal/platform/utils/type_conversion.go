package util

// GetIntValue safely extracts int value from interface{}
func GetIntValue(value interface{}) int {
	switch v := value.(type) {
	case int:
		return v
	case float64:
		return int(v)
	case int64:
		return int(v)
	default:
		return 0
	}
}

// GetFloat64Value safely extracts float64 value from interface{}
func GetFloat64Value(value interface{}) float64 {
	switch v := value.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case int64:
		return float64(v)
	default:
		return 0.0
	}
}
