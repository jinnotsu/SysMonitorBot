package utils

import (
	"context"
	"log"
	"os"

	"SysMonitorBot/internal/storage"
)

var configStore storage.ConfigStore

// InitializeConfigStore はグローバルなコンフィグストアを初期化します
func InitializeConfigStore(store storage.ConfigStore) {
	configStore = store
	log.Println("Config store initialized")
}

// SetEnv は指定された環境変数をストレージに書き込みます
func SetEnv(key, value string) error {
	if configStore == nil {
		return ErrStoreNotInitialized
	}

	ctx := context.Background()
	return configStore.Set(ctx, key, value)
}

// GetEnv はストレージから環境変数を取得します
// 優先度：環境変数 > ストレージ（Redis/ファイル）> デフォルト値
func GetEnv(key string, defaultValue string) string {
	// 優先度1：環境変数から取得
	if envValue := os.Getenv(key); envValue != "" {
		return envValue
	}

	// 優先度2：ストレージから取得
	if configStore == nil {
		log.Printf("Warning: Config store not initialized, using default value for %s", key)
		return defaultValue
	}

	ctx := context.Background()
	value, err := configStore.Get(ctx, key)
	if err != nil {
		log.Printf("Warning: Failed to get %s from config store: %v", key, err)
		return defaultValue
	}

	if value == "" {
		return defaultValue
	}
	return value
}

// ListConfigVariables はストレージからすべての設定変数を取得します
// 優先度：環境変数 > ストレージ（Redis/ファイル）
func ListConfigVariables() map[string]string {
	if configStore == nil {
		log.Println("Warning: Config store not initialized")
		return make(map[string]string)
	}

	ctx := context.Background()
	configs, err := configStore.List(ctx)
	if err != nil {
		log.Printf("Warning: Failed to list config variables: %v", err)
		return make(map[string]string)
	}

	// 環境変数で上書き（優先度）
	varNames := []string{
		"SYSTEM_MONITOR_ENABLED",
		"HEALTH_CHECK_ENABLED",
		"PORT",
		"ANONYMOUS_BUTTON_CHANNEL_ID",
		"ANONYMOUS_POST_CHANNEL_ID",
		"ANONYMOUS_MESSAGE_DELETE_SECONDS",
	}

	for _, varName := range varNames {
		if envValue := os.Getenv(varName); envValue != "" {
			configs[varName] = envValue
		}
	}

	return configs
}

// ErrStoreNotInitialized はストアが初期化されていない場合のエラーです
var ErrStoreNotInitialized = &storeError{"config store not initialized"}

type storeError struct {
	message string
}

func (e *storeError) Error() string {
	return e.message
}
