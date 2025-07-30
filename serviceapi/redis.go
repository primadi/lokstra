package serviceapi

import "github.com/redis/go-redis/v9"

type Redis interface {
	Client() *redis.Client
}
