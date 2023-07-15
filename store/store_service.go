package store

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

// StorageService是一个结构体，它将Redis客户端作为其成员。
type StorageService struct {
	redisClient *redis.Client
}

var (
	// storeService是一个全局变量，它将在整个应用程序中使用。
	storeService = &StorageService{}
	// ctx是一个上下文，它将在整个应用程序中使用。
	ctx = context.Background()
)

// CacheDuration表示缓存的持续时间。
const CacheDuration = 6 * time.Hour

func InitializeStore() *StorageService {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	// 通过调用Ping()方法来检查Redis是否已经启动。
	pong, err := redisClient.Ping(ctx).Result()
	if err != nil {
		panic(fmt.Sprintf("Error init Redis: %v", err))
	}

	fmt.Printf("\nRedis started successfully: pong message = {%s}", pong)
	storeService.redisClient = redisClient
	return storeService
}

// SaveUrlMapping方法将短网址和原始网址保存到Redis中。
func SaveUrlMapping(shortUrl string, originalUrl string, userId string) {
	err := storeService.redisClient.Set(ctx, shortUrl, originalUrl, CacheDuration).Err()
	if err != nil {
		panic(fmt.Sprintf("Failed saving key url | Error: %v - shortUrl: %s - originalUrl: %s\n", err, shortUrl, originalUrl))
	}
}

// RetrieveInitialUrl方法从Redis中检索原始网址。
func RetrieveInitialUrl(shortUrl string) string {
	result, err := storeService.redisClient.Get(ctx, shortUrl).Result()
	if err != nil {
		panic(fmt.Sprintf("Failed retrieving key url | Error: %v - shortUrl: %s\n", err, shortUrl))
	}
	return result
}
