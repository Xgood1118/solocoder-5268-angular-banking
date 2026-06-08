package redis

import (
	"context"
	"net/url"

	"github.com/redis/go-redis/v9"
)

func NewClient(redisURL string) (*redis.Client, error) {
	u, err := url.Parse(redisURL)
	if err != nil {
		return nil, err
	}

	password, _ := u.User.Password()
	db := 0
	if u.Path != "" && len(u.Path) > 1 {
		// parse db number from path
	}

	client := redis.NewClient(&redis.Options{
		Addr:     u.Host,
		Password: password,
		DB:       db,
	})

	_, err = client.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}

	return client, nil
}
