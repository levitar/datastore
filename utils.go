package datastore

import (
	"strings"
	"encoding/json"
)

// Joins multiple strings separated by /
// to form a single string to be used as
// a key on the Redis.
func joinKey(s []string) string {
	return strings.Join(s, "/")
}

// Helper to convert Struct to Map
func FromStructToMap(stru interface{}) map[string]interface{} {
	var (
		json_string []byte
		err error
		ma map[string]interface{}
	)

	json_string, err = json.Marshal(stru)

	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(json_string, &ma)

	if err != nil {
		panic(err)
	}

	return ma
}
