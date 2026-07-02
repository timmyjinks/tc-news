package store

import "time"

type Data struct {
	Name string        `redis:"name"`
	TTL  time.Duration `redis:"ttl"`
}
