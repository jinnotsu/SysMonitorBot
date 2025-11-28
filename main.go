package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

// ヘルスチェック用
func startHealthCheckServer() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("SysMonitorBot is running"))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	log.Printf("Starting health check server on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Printf("Health check server error: %v", err)
	}
}

func main() {
	interval := flag.Int("interval", 1800, "Interval to update system status in seconds")
	flag.Parse()

	// .envファイルから環境変数を読み込む（ファイルが存在しない場合は無視）
	// Docker環境では環境変数が直接渡されるため、エラーは致命的ではない
	if err := godotenv.Load(); err != nil {
		log.Println("Info: .env file not found, using environment variables")
	}

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

	// スラッシュコマンドの呼び出し
	dg.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		SlashCommand(s)
		log.Println("Registered slash commands")
	})

	// セッションの開始
	err = dg.Open()
	if err != nil {
		log.Fatalf("Error: Failed to open session: %v", err)
	}
	defer dg.Close()
	log.Println("Its running!")

	// ヘルスチェックサーバーを起動
	go startHealthCheckServer()

	// status.goの呼び出し
	go UpdateSystemStatus(dg, *interval)

	// Ctrl+Cで終了
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to stop")
	<-stop
}
