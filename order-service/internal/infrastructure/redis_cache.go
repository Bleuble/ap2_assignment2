package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"order-service/internal/domain"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisOrderCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisOrderCache(url string, ttl time.Duration) (*RedisOrderCache, error) {
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

	return &RedisOrderCache{
		client: client,
		ttl:    ttl,
	}, nil
}

func (c *RedisOrderCache) GetClient() *redis.Client {
	return c.client
}

func (c *RedisOrderCache) Set(order *domain.Order) error {
	ctx := context.Background()

	data, err := json.Marshal(order)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("order:%s", order.ID)
	return c.client.Set(ctx, key, data, c.ttl).Err()
}

func (c *RedisOrderCache) Get(id string) (*domain.Order, error) {
	ctx := context.Background()
	key := fmt.Sprintf("order:%s", id)

	data, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	var order domain.Order
	if err := json.Unmarshal(data, &order); err != nil {
		return nil, err
	}

	return &order, nil
}

func (c *RedisOrderCache) Invalidate(id string) error {
	ctx := context.Background()
	key := fmt.Sprintf("order:%s", id)
	return c.client.Del(ctx, key).Err()
}
