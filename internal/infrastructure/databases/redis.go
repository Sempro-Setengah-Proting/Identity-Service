package databases

import (
	"context"
	"fmt"
	"time"

	redisClient "github.com/redis/go-redis/v9"
)

func ConnectRedis(address string, password string, database int) (*redisClient.Client, error) {
	client := redisClient.NewClient(&redisClient.Options{
		Addr:     address,
		Password: password,
		DB:       database,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("connect redis: %w", err)
	}

	return client, nil
}
