package db

import "github.com/redis/go-redis/v9"

type DB struct {
	GetRedisClient func() (*redis.Client, error)
}
