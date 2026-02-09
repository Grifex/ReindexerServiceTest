package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"reindexer-service/internal/dto"

	"github.com/redis/go-redis/v9"
)

type DocCache interface {
	Get(ctx context.Context, id int64) (dto.Document, bool, error)
	Set(ctx context.Context, id int64, doc dto.Document) error
	Delete(ctx context.Context, id int64) error
}

type RedisDocCache struct {
	rdb    *redis.Client
	ttl    time.Duration
	prefix string
}

func NewRedisDocCache(rdb *redis.Client, ttl time.Duration, prefix string) *RedisDocCache {
	if prefix == "" {
		prefix = "doc:"
	}
	return &RedisDocCache{rdb: rdb, ttl: ttl, prefix: prefix}
}

func (c *RedisDocCache) key(id int64) string {
	return fmt.Sprintf("%s%d", c.prefix, id)
}

func (c *RedisDocCache) Get(ctx context.Context, id int64) (dto.Document, bool, error) {
	b, err := c.rdb.Get(ctx, c.key(id)).Bytes()
	if err == redis.Nil { // ключа нет → MISS
		return dto.Document{}, false, nil
	}
	if err != nil {
		return dto.Document{}, false, err
	}

	var out dto.Document
	if err := json.Unmarshal(b, &out); err != nil {
		// битый кеш — удаляем ключ и считаем MISS
		_ = c.rdb.Del(ctx, c.key(id)).Err()
		return dto.Document{}, false, nil
	}

	return out, true, nil
}

func (c *RedisDocCache) Set(ctx context.Context, id int64, doc dto.Document) error {
	b, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	// TTL задаётся прямо в Set (последний аргумент).
	return c.rdb.Set(ctx, c.key(id), b, c.ttl).Err()
}

func (c *RedisDocCache) Delete(ctx context.Context, id int64) error {
	return c.rdb.Del(ctx, c.key(id)).Err()
}
