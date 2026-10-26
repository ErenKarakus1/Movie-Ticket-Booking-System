package redis

import (
	"errors"

	"github.com/redis/go-redis/v9"
)

func NewClient(redisURL string) (*redis.Client, error) {
	parsedURL, err := redis.ParseURL(redisURL)
	if err != nil {
		return &redis.Client{}, errors.New("internal server error")
	}
	return redis.NewClient(parsedURL), nil
}
