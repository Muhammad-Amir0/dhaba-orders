package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisQueue struct {
	client *redis.Client
}

func NewRedisQueue(client *redis.Client) *RedisQueue {
	return &RedisQueue{client: client}
}

func (queue *RedisQueue) EnqueueOrder(ctx context.Context, order Order, now time.Time) error {
	pipe := queue.client.Pipeline()
	pipe.LPush(ctx, "queue:orders", strconv.FormatInt(order.ID, 10))
	pipe.ZIncrBy(ctx, "stats:orders:minute", 1, minuteBucket(now))
	for _, item := range order.Items {
		pipe.ZIncrBy(ctx, "stats:dishes", float64(item.Qty), item.Name)
	}
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("update order queue and stats: %w", err)
	}
	return nil
}

func (queue *RedisQueue) Ping(ctx context.Context) error {
	return queue.client.Ping(ctx).Err()
}
