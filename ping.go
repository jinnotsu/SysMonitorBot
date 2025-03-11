package main

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

// Pingスラッシュコマンドを登録
func SlashCommand(s *discordgo.Session) {
	s.AddHandler(handlePing)

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

// Pingスラッシュコマンドの処理
func handlePing(s *discordgo.Session, i *discordgo.InteractionCreate) {
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
		log.Fatalf("Error: Failed to respond to ping command: %v", err)
	}
}
