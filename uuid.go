package datastore

import (
	"fmt"
	"log"
	"os"
)

var random *os.File

func init() {
	f, err := os.Open("/dev/urandom")
	if err != nil {
		log.Fatal(err)
	}
	random = f
}

// GenerateID generate an ID to be used on the database
func GenerateID(size int) string {
	b := make([]byte, size)
	random.Read(b)
	return fmt.Sprintf("%x", b)
}
