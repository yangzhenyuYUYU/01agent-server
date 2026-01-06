package tools

import (
	"context"
	"fmt"
	"sync"
	"time"

	"01agent_server/internal/config"
	"01agent_server/internal/repository"

	"github.com/go-redis/redis/v8"
)

// Redis 通用的Redis工具类，支持多数据库
type Redis struct {
	clients map[int]*redis.Client
	mu      sync.RWMutex
}

var (
	redisInstance *Redis
	once          sync.Once
)

// GetRedisInstance 获取Redis单例实例
func GetRedisInstance() *Redis {
	once.Do(func() {
		redisInstance = &Redis{
			clients: make(map[int]*redis.Client),
		}
	})
	return redisInstance
}

// getClient 获取指定数据库的Redis客户端
// 如果不存在则创建新的客户端
func (r *Redis) getClient(db int) (*redis.Client, error) {
	r.mu.RLock()
	client, exists := r.clients[db]
	r.mu.RUnlock()

	if exists && client != nil {
		// 测试连接是否有效
		ctx := context.Background()
		if err := client.Ping(ctx).Err(); err == nil {
			return client, nil
		}
		// 连接无效，移除并重新创建
		r.mu.Lock()
		delete(r.clients, db)
		r.mu.Unlock()
	}

	// 创建新客户端
	cfg := config.AppConfig.Redis
	client = redis.NewClient(&redis.Options{
		Addr:       fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:   cfg.Password,
		DB:         db,
		PoolSize:   cfg.PoolSize,
		MaxRetries: cfg.MaxRetries,
	})

	// 测试连接
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis DB %d: %w", db, err)
	}

	// 保存客户端
	r.mu.Lock()
	r.clients[db] = client
	r.mu.Unlock()

	return client, nil
}

// Get 获取值
func (r *Redis) Get(key string, db int) (string, error) {
	client, err := r.getClient(db)
	if err != nil {
		return "", err
	}

	ctx := context.Background()
	val, err := client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil // key不存在，返回空字符串
	}
	return val, err
}

// Set 设置值
func (r *Redis) Set(key string, value string, expiration int, db int) error {
	client, err := r.getClient(db)
	if err != nil {
		return err
	}

	ctx := context.Background()
	if expiration > 0 {
		return client.Set(ctx, key, value, time.Duration(expiration)*time.Second).Err()
	}
	return client.Set(ctx, key, value, 0).Err()
}

// Delete 删除键
func (r *Redis) Delete(key string, db int) error {
	client, err := r.getClient(db)
	if err != nil {
		return err
	}

	ctx := context.Background()
	return client.Del(ctx, key).Err()
}

// DeleteKeys 批量删除键
func (r *Redis) DeleteKeys(keys []string, db int) error {
	if len(keys) == 0 {
		return nil
	}

	client, err := r.getClient(db)
	if err != nil {
		return err
	}

	ctx := context.Background()
	return client.Del(ctx, keys...).Err()
}

// Exists 检查键是否存在
func (r *Redis) Exists(key string, db int) (bool, error) {
	client, err := r.getClient(db)
	if err != nil {
		return false, err
	}

	ctx := context.Background()
	count, err := client.Exists(ctx, key).Result()
	return count > 0, err
}

// Keys 获取所有匹配模式的键
func (r *Redis) Keys(pattern string, db int) ([]string, error) {
	client, err := r.getClient(db)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	return client.Keys(ctx, pattern).Result()
}

// Scan 扫描键（推荐用于生产环境，避免阻塞）
func (r *Redis) Scan(cursor uint64, match string, count int64, db int) ([]string, uint64, error) {
	client, err := r.getClient(db)
	if err != nil {
		return nil, 0, err
	}

	ctx := context.Background()
	keys, nextCursor, err := client.Scan(ctx, cursor, match, count).Result()
	return keys, nextCursor, err
}

// GetAllKeys 获取所有键（使用Scan，避免阻塞）
func (r *Redis) GetAllKeys(pattern string, db int) ([]string, error) {
	var allKeys []string
	var cursor uint64 = 0
	const batchSize int64 = 100

	for {
		keys, nextCursor, err := r.Scan(cursor, pattern, batchSize, db)
		if err != nil {
			return nil, err
		}
		allKeys = append(allKeys, keys...)

		if nextCursor == 0 {
			break
		}
		cursor = nextCursor
	}

	return allKeys, nil
}

// GetKeyInfo 获取键的详细信息
type KeyInfo struct {
	Key        string `json:"key"`
	Type       string `json:"type"`
	TTL        int64  `json:"ttl"`        // 剩余过期时间（秒），-1表示永不过期，-2表示不存在
	Size       int64  `json:"size"`       // 值的字节大小
	Value      string `json:"value"`      // 值（仅字符串类型）
	ValueLen   int64  `json:"value_len"`  // 值的长度（列表、集合等）
}

// GetKeyInfo 获取键的详细信息
func (r *Redis) GetKeyInfo(key string, db int) (*KeyInfo, error) {
	client, err := r.getClient(db)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	info := &KeyInfo{Key: key}

	// 获取类型
	keyType, err := client.Type(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	info.Type = keyType

	// 如果键不存在
	if keyType == "none" {
		info.TTL = -2
		return info, nil
	}

	// 获取TTL
	ttl, err := client.TTL(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	info.TTL = int64(ttl.Seconds())

	// 根据类型获取值或大小
	switch keyType {
	case "string":
		val, err := client.Get(ctx, key).Result()
		if err != nil && err != redis.Nil {
			return nil, err
		}
		info.Value = val
		info.Size = int64(len(val))
	case "list":
		len, err := client.LLen(ctx, key).Result()
		if err != nil {
			return nil, err
		}
		info.ValueLen = len
	case "set":
		len, err := client.SCard(ctx, key).Result()
		if err != nil {
			return nil, err
		}
		info.ValueLen = len
	case "zset":
		len, err := client.ZCard(ctx, key).Result()
		if err != nil {
			return nil, err
		}
		info.ValueLen = len
	case "hash":
		len, err := client.HLen(ctx, key).Result()
		if err != nil {
			return nil, err
		}
		info.ValueLen = len
	}

	return info, nil
}

// ClearAll 清空指定数据库的所有数据（危险操作）
func (r *Redis) ClearAll(db int) error {
	client, err := r.getClient(db)
	if err != nil {
		return err
	}

	ctx := context.Background()
	return client.FlushDB(ctx).Err()
}

// DBSize 获取指定数据库的键数量
func (r *Redis) DBSize(db int) (int64, error) {
	client, err := r.getClient(db)
	if err != nil {
		return 0, err
	}

	ctx := context.Background()
	return client.DBSize(ctx).Result()
}

// ClearUserCache 清除用户相关的Redis缓存
// 清除 user:info:all:{user_id} 和 user:info:basic:{user_id} 两个key
// 使用 db3 作为用户缓存数据库
func ClearUserCache(userID string) error {
	if userID == "" {
		return fmt.Errorf("userID cannot be empty")
	}

	redisInstance := GetRedisInstance()
	keys := []string{
		fmt.Sprintf("user:info:all:%s", userID),
		fmt.Sprintf("user:info:basic:%s", userID),
	}

	// 使用 db3 作为用户缓存数据库
	err := redisInstance.DeleteKeys(keys, 3)
	if err != nil {
		repository.Warnf("Failed to clear user cache for user %s: %v", userID, err)
		return err
	}

	repository.Infof("Cleared user cache for user: %s (keys: %v)", userID, keys)
	return nil
}

// ClearUserCacheAsync 异步清除用户缓存（不阻塞主流程）
func ClearUserCacheAsync(userID string) {
	go func() {
		if err := ClearUserCache(userID); err != nil {
			repository.Warnf("Failed to clear user cache asynchronously: %v", err)
		}
	}()
}
