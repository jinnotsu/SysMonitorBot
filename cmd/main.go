package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"

	"SysMonitorBot/internal/handlers"
	"SysMonitorBot/internal/server"
	"SysMonitorBot/internal/services"
	"SysMonitorBot/internal/storage"
	"SysMonitorBot/internal/utils"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

func main() {
	interval := flag.Int("interval", 1800, "Interval to update system status in seconds")
	flag.Parse()

	// .envファイルから環境変数を読み込む（ファイルが存在しない場合は無視）
	// Docker環境では環境変数が直接渡されるため、エラーは致命的ではない
	if err := godotenv.Load(); err != nil {
		log.Println("Info: .env file not found, using environment variables")
	}

	// コンフィグストアを初期化
	redisURL := os.Getenv("REDIS_URL")
	useRedis := strings.ToLower(os.Getenv("USE_REDIS")) == "true"

	var store storage.ConfigStore
	if useRedis && redisURL != "" {
		log.Println("Initializing Redis config store...")
		store = storage.NewFallbackStore(redisURL)
	} else {
		log.Println("Initializing file-based config store...")
		store = storage.NewFileStore()
	}
	utils.InitializeConfigStore(store)
	defer store.Close()

	// ストアが空の場合、.env から初期化
	ctx, cancel := context.WithTimeout(context.Background(), 10*1000000000) // 10秒
	defer cancel()

	isEmpty, err := store.IsEmpty(ctx)
	if err != nil {
		log.Printf("Warning: Failed to check if store is empty: %v", err)
	} else if isEmpty {
		log.Println("Store is empty. Initializing from .env file...")
		if err := store.InitializeFromFile(ctx); err != nil {
			log.Printf("Warning: Failed to initialize store from .env file: %v", err)
		}
	} else {
		log.Println("Store already contains data. Skipping initialization.")
	}

	// 管理者ユーザーIDを読み込む
	adminUserIDs := os.Getenv("ADMIN_USER_IDS")
	utils.LoadAdminUsers(adminUserIDs)

	// 環境変数からDISCORD_TOKENを取得
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("Error: DISCORD_TOKEN is not set")
	}

	// Discord Bot セッションの作成
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("Error: Failed to create session")
	}

	// スラッシュコマンドの登録
	dg.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		handlers.RegisterSlashCommands(s)
		handlers.RegisterConfigCommand(s)
		log.Println("Registered slash commands")

		// 匿名ボードの設置
		handlers.SetupAnonymousBoard(s)
	})

	// 匿名ボードのインタラクションハンドラーを追加
	dg.AddHandler(handlers.HandleAnonymousBoardInteraction)

	// /config コマンドハンドラーを追加
	dg.AddHandler(handlers.HandleConfigCommand)

	// セッションの開始
	err = dg.Open()
	if err != nil {
		log.Fatalf("Error: Failed to open session: %v", err)
	}
	defer dg.Close()
	log.Println("Its running!")

	// ヘルスチェックサーバーを起動（環境変数で制御）
	healthCheckEnabled := os.Getenv("HEALTH_CHECK_ENABLED")
	if healthCheckEnabled == "" || strings.ToLower(healthCheckEnabled) == "true" {
		go server.StartHealthCheckServer()
	} else {
		log.Println("Health check server is disabled")
	}

	// システムモニタリングの開始（環境変数で制御）
	systemMonitorEnabled := os.Getenv("SYSTEM_MONITOR_ENABLED")
	if systemMonitorEnabled == "" || strings.ToLower(systemMonitorEnabled) == "true" {
		go services.UpdateSystemStatus(dg, *interval)
	} else {
		log.Println("System monitor is disabled")
	}

	// Ctrl+Cで終了
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to stop")
	<-stop
}
