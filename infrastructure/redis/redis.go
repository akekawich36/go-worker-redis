package redis

import "github.com/go-redis/redis/v8"

func SetupRedis() *redis.Client {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	return redisClient
}
