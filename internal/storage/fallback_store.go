package storage

import (
	"context"
	"fmt"
	"log"
)

// FallbackStore は Redis への接続に失敗した場合、.env ファイルにフォールバックします
type FallbackStore struct {
	primary     ConfigStore
	fallback    ConfigStore
	useFallback bool
}

// NewFallbackStore は FallbackStore を作成します
func NewFallbackStore(redisURL string) *FallbackStore {
	fs := &FallbackStore{
		fallback:    NewFileStore(),
		useFallback: false,
	}

	// Redis に接続を試みる
	redis, err := NewRedisStore(redisURL)
	if err != nil {
		log.Printf("Warning: Failed to connect to Redis, falling back to file store: %v", err)
		fs.useFallback = true
		fs.primary = fs.fallback
		return fs
	}

	fs.primary = redis
	return fs
}

// Set は設定変数を保存します（プライマリ、またはフォールバック）
func (f *FallbackStore) Set(ctx context.Context, key, value string) error {
	// プライマリに書き込む
	if err := f.primary.Set(ctx, key, value); err != nil {
		log.Printf("Error: Primary store write failed: %v", err)

		// フォールバック店に書き込む
		if f.fallback != f.primary {
			log.Println("Attempting to write to fallback store...")
			if fbErr := f.fallback.Set(ctx, key, value); fbErr != nil {
				return fmt.Errorf("both primary and fallback stores failed: primary=%v, fallback=%v", err, fbErr)
			}
			log.Println("Successfully written to fallback store")
		}
		return err
	}

	return nil
}

// Get は設定変数を取得します（プライマリ、またはフォールバック）
func (f *FallbackStore) Get(ctx context.Context, key string) (string, error) {
	val, err := f.primary.Get(ctx, key)
	if err != nil {
		log.Printf("Error: Primary store read failed: %v", err)

		// フォールバック店から読み込む
		if f.fallback != f.primary {
			log.Println("Attempting to read from fallback store...")
			return f.fallback.Get(ctx, key)
		}
		return "", err
	}

	return val, nil
}

// List はすべての設定変数を取得します（プライマリ、またはフォールバック）
func (f *FallbackStore) List(ctx context.Context) (map[string]string, error) {
	configs, err := f.primary.List(ctx)
	if err != nil {
		log.Printf("Error: Primary store list failed: %v", err)

		// フォールバック店から読み込む
		if f.fallback != f.primary {
			log.Println("Attempting to list from fallback store...")
			return f.fallback.List(ctx)
		}
		return nil, err
	}

	return configs, nil
}

// Delete は設定変数を削除します（プライマリ、またはフォールバック）
func (f *FallbackStore) Delete(ctx context.Context, key string) error {
	if err := f.primary.Delete(ctx, key); err != nil {
		log.Printf("Error: Primary store delete failed: %v", err)

		// フォールバック店から削除する
		if f.fallback != f.primary {
			log.Println("Attempting to delete from fallback store...")
			return f.fallback.Delete(ctx, key)
		}
		return err
	}

	return nil
}

// Close はストア接続をクローズします
func (f *FallbackStore) Close() error {
	// プライマリをクローズ
	if err := f.primary.Close(); err != nil {
		log.Printf("Warning: Failed to close primary store: %v", err)
	}

	// フォールバックがプライマリと異なる場合もクローズ
	if f.fallback != f.primary {
		if err := f.fallback.Close(); err != nil {
			log.Printf("Warning: Failed to close fallback store: %v", err)
		}
	}

	return nil
}

// IsUsingFallback はフォールバックを使用しているかを返します
func (f *FallbackStore) IsUsingFallback() bool {
	return f.useFallback
}

// IsEmpty はストアが空かどうかを判定します（プライマリ、またはフォールバック）
func (f *FallbackStore) IsEmpty(ctx context.Context) (bool, error) {
	isEmpty, err := f.primary.IsEmpty(ctx)
	if err != nil {
		log.Printf("Error: Primary store IsEmpty failed: %v", err)

		// フォールバック店から確認する
		if f.fallback != f.primary {
			log.Println("Attempting to check IsEmpty from fallback store...")
			return f.fallback.IsEmpty(ctx)
		}
		return false, err
	}

	return isEmpty, nil
}

// InitializeFromFile は .env ファイルからストアを初期化します（プライマリ、またはフォールバック）
func (f *FallbackStore) InitializeFromFile(ctx context.Context) error {
	if err := f.primary.InitializeFromFile(ctx); err != nil {
		log.Printf("Error: Primary store InitializeFromFile failed: %v", err)

		// フォールバック店から初期化する
		if f.fallback != f.primary {
			log.Println("Attempting to InitializeFromFile from fallback store...")
			return f.fallback.InitializeFromFile(ctx)
		}
		return err
	}

	return nil
}
