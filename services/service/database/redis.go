package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

func KeyMessageAckIndex(account string) string {
	return fmt.Sprintf("chat:ack:%s", account)
}

func InitRedis(addr, pass string) (*redis.Client, error) {
	redisdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     pass,
		DialTimeout:  time.Second * 5,
		ReadTimeout:  time.Second * 5,
		WriteTimeout: time.Second * 5,
	})

	_, err := redisdb.Ping(context.Background()).Result()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return redisdb, nil
}
