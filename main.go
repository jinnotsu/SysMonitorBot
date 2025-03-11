package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

func main() {
    interval := flag.Int("interval", 10, "Interval to update system status in seconds")
	flag.Parse()

	// .envファイルから環境変数を読み込む
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error: Failed to load .env file: %v", err)
	}

	// .envファイルからDISCORD_TOKENを取得
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

	// status.goの呼び出し
	go UpdateSystemStatus(dg, *interval)
	
	// Ctrl+Cで終了
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to stop")
	<-stop
}
