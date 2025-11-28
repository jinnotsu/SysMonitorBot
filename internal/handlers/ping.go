package handlers

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

// RegisterSlashCommands はスラッシュコマンドを登録します
func RegisterSlashCommands(s *discordgo.Session) {
	s.AddHandler(HandlePing)

	cmd := &discordgo.ApplicationCommand{
		Name:        "ping",
		Description: "Reply with Pong!",
	}
	// 登録
	_, err := s.ApplicationCommandCreate(s.State.User.ID, "", cmd)
	if err != nil {
		log.Fatalf("Error: Failed to create slash command: %v", err)
	}
}

// HandlePing はPingスラッシュコマンドの処理を行います
func HandlePing(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// ApplicationCommandタイプのインタラクションのみ処理
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}
	if i.ApplicationCommandData().Name != "ping" {
		return
	}
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Pong! (%dms)", s.HeartbeatLatency().Milliseconds()),
		},
	})
	if err != nil {
		log.Printf("Error: Failed to respond to ping command: %v", err)
	}
}
