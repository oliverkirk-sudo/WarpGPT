package db

import (
	"WarpGPT/pkg/env"
	"WarpGPT/pkg/logger"
	"context"
	"github.com/redis/go-redis/v9"
)

type RedisDB struct {
}

func (r RedisDB) GetRedisClient() (*redis.Client, error) {

	redisClient := redis.NewClient(&redis.Options{
		Addr:           env.Env.RedisAddress,
		Password:       env.Env.RedisPasswd,
		DB:             env.Env.RedisDB,
		MaxRetries:     3,
		MaxActiveConns: 20,
	})

	_, err := redisClient.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}
	logger.Log.Info("成功连接到Redis")

	return redisClient, nil
}
