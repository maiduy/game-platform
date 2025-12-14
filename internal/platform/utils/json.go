package util

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"

	"github.com/sirupsen/logrus"
)

type Response struct {
	StatusCode int         `json:"statusCode"`
	Method     string      `json:"method"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data"`
}

func Strigify(payload interface{}) []byte {
	response, _ := json.Marshal(payload)
	return response
}

func Parse(payload []byte) Response {
	var jsonResponse Response
	err := json.Unmarshal(payload, &jsonResponse)

	if err != nil {
		logrus.Fatal(err.Error())
	}

	return jsonResponse
}

// ToJSON converts an interface to a JSON reader
func ToJSON(v interface{}) (io.Reader, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(data), nil
}

// FromJSON converts a JSON reader to an interface
func FromJSON(r io.Reader, v interface{}) error {
	return json.NewDecoder(r).Decode(v)
}

// PrettyJSON returns a pretty-printed JSON string
func PrettyJSON(v interface{}) (string, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// IsValidJSON checks if a string is valid JSON
func IsValidJSON(s string) bool {
	var js interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}

// CompactJSON removes whitespace from a JSON string
func CompactJSON(s string) (string, error) {
	var buf bytes.Buffer
	err := json.Compact(&buf, []byte(s))
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// MergeJSON merges two JSON objects
func MergeJSON(base, overlay string) (string, error) {
	var baseMap, overlayMap map[string]interface{}

	if err := json.Unmarshal([]byte(base), &baseMap); err != nil {
		return "", err
	}

	if err := json.Unmarshal([]byte(overlay), &overlayMap); err != nil {
		return "", err
	}

	// Merge overlay into base
	for k, v := range overlayMap {
		baseMap[k] = v
	}

	result, err := json.Marshal(baseMap)
	if err != nil {
		return "", err
	}

	return string(result), nil
}

// JSONToMap converts a JSON string to a map
func JSONToMap(s string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := json.Unmarshal([]byte(s), &result)
	return result, err
}

// MapToJSON converts a map to a JSON string
func MapToJSON(m map[string]interface{}) (string, error) {
	result, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(result), nil
}

// GetJSONValue gets a value from a JSON string by path
func GetJSONValue(jsonStr string, path string) (interface{}, error) {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil, err
	}

	keys := strings.Split(path, ".")
	current := data

	for i, key := range keys {
		if i == len(keys)-1 {
			return current[key], nil
		}

		if next, ok := current[key].(map[string]interface{}); ok {
			current = next
		} else {
			return nil, nil
		}
	}

	return nil, nil
}
