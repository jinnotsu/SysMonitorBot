package storage

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
)

const envFile = ".env"

// FileStore は .env ファイルをバックエンドとした ConfigStore の実装です
type FileStore struct{}

// NewFileStore は FileStore を作成します
func NewFileStore() *FileStore {
	return &FileStore{}
}

// Set は設定変数を .env ファイルに保存します
func (f *FileStore) Set(ctx context.Context, key, value string) error {
	// ファイルの内容を読み込む
	content, err := os.ReadFile(envFile)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf(".env file read error: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	found := false
	var newLines []string

	// 既存の行を置き換えまたは保持
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// コメント行や空行、対象の変数以外はそのまま保持
		if strings.HasPrefix(trimmed, key+"=") {
			newLines = append(newLines, fmt.Sprintf("%s=%s", key, value))
			found = true
		} else {
			newLines = append(newLines, line)
		}
	}

	// 変数が存在しない場合は新規追加
	if !found {
		newLines = append(newLines, fmt.Sprintf("%s=%s", key, value))
	}

	// .env ファイルに書き込む
	newContent := strings.Join(newLines, "\n")
	if err := os.WriteFile(envFile, []byte(newContent), 0644); err != nil {
		return fmt.Errorf(".env file write error: %w", err)
	}

	// メモリ上の環境変数も更新
	os.Setenv(key, value)

	log.Printf("Updated %s in .env file", key)
	return nil
}

// Get は設定変数を .env ファイルから取得します
func (f *FileStore) Get(ctx context.Context, key string) (string, error) {
	value := os.Getenv(key)
	return value, nil
}

// List はすべての設定変数を .env ファイルから取得します
func (f *FileStore) List(ctx context.Context) (map[string]string, error) {
	content, err := os.ReadFile(envFile)
	if err != nil {
		log.Printf("Warning: Cannot read .env file: %v", err)
		return make(map[string]string), nil
	}

	result := make(map[string]string)
	lines := strings.Split(string(content), "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// コメント行、空行、DISCORD_TOKENをスキップ
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "DISCORD_TOKEN=") {
			continue
		}

		// key=value の形式で分割
		parts := strings.SplitN(trimmed, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// 既知の設定変数のみに限定
		knownVars := []string{
			"SYSTEM_MONITOR_ENABLED",
			"HEALTH_CHECK_ENABLED",
			"PORT",
			"ANONYMOUS_BUTTON_CHANNEL_ID",
			"ANONYMOUS_POST_CHANNEL_ID",
			"ANONYMOUS_MESSAGE_DELETE_SECONDS",
			"ADMIN_USER_IDS",
			"USE_REDIS",
			"REDIS_URL",
		}

		for _, configVar := range knownVars {
			if key == configVar {
				result[key] = value
				break
			}
		}
	}

	return result, nil
}

// Delete は設定変数を .env ファイルから削除します（この実装では未使用）
func (f *FileStore) Delete(ctx context.Context, key string) error {
	return fmt.Errorf("delete operation not supported for file store")
}

// Close はファイルストアをクローズします（何もしない）
func (f *FileStore) Close() error {
	return nil
}
