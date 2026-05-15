package infrastructure

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisIdempotencyStore struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisIdempotencyStore(url string) (*RedisIdempotencyStore, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		opts = &redis.Options{
			Addr: url,
		}
	}
	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis connection failed: %v", err)
	}

	return &RedisIdempotencyStore{
		client: client,
		ttl:    24 * time.Hour,
	}, nil
}

func (s *RedisIdempotencyStore) IsProcessed(messageID string) (bool, error) {
	ctx := context.Background()
	key := fmt.Sprintf("processed_msg:%s", messageID)

	val, err := s.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return val == "1", nil
}

func (s *RedisIdempotencyStore) MarkProcessed(messageID string) error {
	ctx := context.Background()
	key := fmt.Sprintf("processed_msg:%s", messageID)
	return s.client.Set(ctx, key, "1", s.ttl).Err()
}
