package datastore

import (
	log "github.com/Sirupsen/logrus"
	"gopkg.in/redis.v3"
)

var (
	Conn *redis.Client
	Log  log.Logger
)

func init() {
	Conn = redis.NewClient(&redis.Options{
		Network: "tcp",
		Addr:    "127.0.0.1:6379",
		DB:      0,
	})
}
