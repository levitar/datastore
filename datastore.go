package datastore

import (
	"strings"
)

// Joins multiple strings separated by /
// to form a single string to be used as
// a key on the Redis.
func joinKey(s []string) string {
	return strings.Join(s, "/")
}
