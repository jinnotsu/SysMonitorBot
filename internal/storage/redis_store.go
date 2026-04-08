package storage

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

const (
	// キープレフィックス：Redis内で設定変数を識別
	configKeyPrefix = "bot_config:"
)

// RedisStore は Redis をバックエンドとした ConfigStore の実装です
type RedisStore struct {
	client *redis.Client
}

// NewRedisStore は Redis 接続を初期化し、RedisStore を返します
func NewRedisStore(redisURL string) (*RedisStore, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse REDIS_URL: %w", err)
	}

	client := redis.NewClient(opt)

	// 接続確認
	ctx, cancel := context.WithTimeout(context.Background(), 5*1000000000) // 5秒
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Println("Successfully connected to Redis")

	return &RedisStore{client: client}, nil
}

// Set は設定変数を Redis に保存します
func (r *RedisStore) Set(ctx context.Context, key, value string) error {
	redisKey := configKeyPrefix + key
	return r.client.Set(ctx, redisKey, value, 0).Err()
}

// Get は設定変数を Redis から取得します
func (r *RedisStore) Get(ctx context.Context, key string) (string, error) {
	redisKey := configKeyPrefix + key
	val, err := r.client.Get(ctx, redisKey).Result()
	if err == redis.Nil {
		// キーが存在しない場合は空文字列を返す
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return val, nil
}

// List はすべての設定変数を Redis から取得します
func (r *RedisStore) List(ctx context.Context) (map[string]string, error) {
	result := make(map[string]string)

	// bot_config:* キーをスキャン
	iter := r.client.Scan(ctx, 0, configKeyPrefix+"*", 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		// プレフィックスを除去して変数名を取得
		varName := key[len(configKeyPrefix):]

		val, err := r.client.Get(ctx, key).Result()
		if err != nil {
			log.Printf("Warning: failed to get %s: %v", key, err)
			continue
		}
		result[varName] = val
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan Redis keys: %w", err)
	}

	return result, nil
}

// Delete は設定変数を Redis から削除します
func (r *RedisStore) Delete(ctx context.Context, key string) error {
	redisKey := configKeyPrefix + key
	return r.client.Del(ctx, redisKey).Err()
}

// Close は Redis 接続をクローズします
func (r *RedisStore) Close() error {
	return r.client.Close()
}

// IsEmpty は Redis にデータが格納されているかどうかを判定します
func (r *RedisStore) IsEmpty(ctx context.Context) (bool, error) {
	// bot_config:* キーが存在するか確認
	keys, err := r.client.Keys(ctx, configKeyPrefix+"*").Result()
	if err != nil {
		return false, fmt.Errorf("failed to check Redis keys: %w", err)
	}

	return len(keys) == 0, nil
}

// InitializeFromFile は .env ファイルから Redis を初期化します
func (r *RedisStore) InitializeFromFile(ctx context.Context) error {
	log.Println("Initializing Redis from .env file...")

	fileStore := NewFileStore()
	configs, err := fileStore.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to read .env file: %w", err)
	}

	if len(configs) == 0 {
		log.Println("No configuration variables found in .env file")
		return nil
	}

	// .env から読み込んだ全設定変数を Redis に書き込む
	for key, value := range configs {
		if err := r.Set(ctx, key, value); err != nil {
			return fmt.Errorf("failed to set %s in Redis: %w", key, err)
		}
		log.Printf("Initialized Redis: %s=%s", key, value)
	}

	log.Printf("Successfully initialized Redis with %d configuration variables", len(configs))
	return nil
}

// InitializeFromEnv は環境変数から Redis を初期化します
func (r *RedisStore) InitializeFromEnv(ctx context.Context) error {
	log.Println("Initializing Redis from environment variables...")

	// 環境変数から初期化する設定変数一覧
	varNames := []string{
		"SYSTEM_MONITOR_ENABLED",
		"HEALTH_CHECK_ENABLED",
		"PORT",
		"ANONYMOUS_BUTTON_CHANNEL_ID",
		"ANONYMOUS_POST_CHANNEL_ID",
		"ANONYMOUS_MESSAGE_DELETE_SECONDS",
	}

	count := 0
	for _, varName := range varNames {
		value := os.Getenv(varName)
		if value != "" {
			if err := r.Set(ctx, varName, value); err != nil {
				return fmt.Errorf("failed to set %s in Redis: %w", varName, err)
			}
			log.Printf("Initialized Redis: %s=%s", varName, value)
			count++
		}
	}

	log.Printf("Successfully initialized Redis with %d environment variables", count)
	return nil
}
