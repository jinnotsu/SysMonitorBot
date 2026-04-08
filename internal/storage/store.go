package storage

import "context"

// ConfigStore はボット設定の永続化層を抽象化するインターフェースです
type ConfigStore interface {
	// Set は設定変数を保存します
	Set(ctx context.Context, key, value string) error

	// Get は設定変数を取得します。キーが存在しない場合は空文字列を返します
	Get(ctx context.Context, key string) (string, error)

	// List はすべての設定変数をマップで取得します
	List(ctx context.Context) (map[string]string, error)

	// Delete は設定変数を削除します
	Delete(ctx context.Context, key string) error

	// Close はストア接続をクローズします
	Close() error

	// IsEmpty はストアが空（データがない）かどうかを判定します
	IsEmpty(ctx context.Context) (bool, error)

	// InitializeFromFile は .env ファイルからストアを初期化します
	InitializeFromFile(ctx context.Context) error

	// InitializeFromEnv は環境変数からストアを初期化します
	InitializeFromEnv(ctx context.Context) error
}
