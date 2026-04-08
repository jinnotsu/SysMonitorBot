package handlers

import (
	"fmt"
	"log"
	"strings"

	"SysMonitorBot/internal/utils"
	"github.com/bwmarrin/discordgo"
)

// RegisterConfigCommand は /config スラッシュコマンドを登録します
func RegisterConfigCommand(s *discordgo.Session) {
	log.Println("DEBUG: Attempting to register /config command")

	// 既存のコマンドを削除してから登録（重複回避）
	commands, err := s.ApplicationCommands(s.State.User.ID, "")
	if err != nil {
		log.Printf("Warning: Failed to get existing commands: %v", err)
	} else {
		for _, cmd := range commands {
			if cmd.Name == "config" {
				log.Println("DEBUG: Removing existing /config command")
				s.ApplicationCommandDelete(s.State.User.ID, "", cmd.ID)
			}
		}
	}

	// 設定変数の選択肢を生成
	configVariableNames := []string{
		"SYSTEM_MONITOR_ENABLED",
		"HEALTH_CHECK_ENABLED",
		"PORT",
		"ANONYMOUS_BUTTON_CHANNEL_ID",
		"ANONYMOUS_POST_CHANNEL_ID",
		"ANONYMOUS_MESSAGE_DELETE_SECONDS",
	}

	var setChoices []*discordgo.ApplicationCommandOptionChoice
	var getChoices []*discordgo.ApplicationCommandOptionChoice

	for _, v := range configVariableNames {
		setChoices = append(setChoices, &discordgo.ApplicationCommandOptionChoice{
			Name:  v,
			Value: v,
		})
		getChoices = append(getChoices, &discordgo.ApplicationCommandOptionChoice{
			Name:  v,
			Value: v,
		})
	}

	cmd := &discordgo.ApplicationCommand{
		Name:        "config",
		Description: "Bot の設定を管理します（管理者のみ）",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "set",
				Description: "設定変数を変更します",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "variable",
						Description: "変数名",
						Required:    true,
						Choices:     setChoices,
					},
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "value",
						Description: "新しい値",
						Required:    true,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "get",
				Description: "設定変数の現在値を表示します",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "variable",
						Description: "変数名",
						Required:    true,
						Choices:     getChoices,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "list",
				Description: "すべての設定を表示します",
			},
		},
	}

	_, err = s.ApplicationCommandCreate(s.State.User.ID, "", cmd)
	if err != nil {
		log.Fatalf("Error: Failed to create /config command: %v", err)
	}
	log.Println("DEBUG: Successfully registered /config command")
}

// HandleConfigCommand は /config スラッシュコマンドを処理します
func HandleConfigCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Println("DEBUG: HandleConfigCommand called")

	// ApplicationCommandタイプのインタラクションのみ処理
	if i.Type != discordgo.InteractionApplicationCommand {
		log.Printf("DEBUG: Not ApplicationCommand type: %v", i.Type)
		return
	}

	commandName := i.ApplicationCommandData().Name
	log.Printf("DEBUG: Command name: %s", commandName)

	if commandName != "config" {
		return
	}

	log.Printf("DEBUG: Config command received from user: %s", i.Member.User.ID)

	// 権限チェック
	if !utils.IsAdmin(i.Member.User.ID) {
		log.Printf("DEBUG: User %s is not admin", i.Member.User.ID)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "権限がありません",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	options := i.ApplicationCommandData().Options
	if len(options) == 0 {
		log.Println("DEBUG: No options provided")
		return
	}

	subcommand := options[0].Name
	subcommandOptions := options[0].Options
	log.Printf("DEBUG: Subcommand: %s", subcommand)

	switch subcommand {
	case "set":
		handleConfigSet(s, i, subcommandOptions)
	case "get":
		handleConfigGet(s, i, subcommandOptions)
	case "list":
		handleConfigList(s, i)
	}
}

// handleConfigSet は "config set" サブコマンドを処理します
func handleConfigSet(s *discordgo.Session, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption) {
	var varName, value string

	for _, opt := range options {
		if opt.Name == "variable" {
			varName = opt.StringValue()
		} else if opt.Name == "value" {
			value = opt.StringValue()
		}
	}

	log.Printf("DEBUG: Set command - variable: %s, value: %s", varName, value)

	// バリデーション
	if err := utils.ValidateConfig(varName, value); err != nil {
		log.Printf("DEBUG: Validation error: %v", err)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("❌ エラー: %v", err),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	// ストレージに書き込む
	if err := utils.SetEnv(varName, value); err != nil {
		log.Printf("DEBUG: SetEnv error: %v", err)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("❌ エラー: %v", err),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	log.Printf("DEBUG: Successfully set %s to %s", varName, value)

	// 成功レスポンス
	successMsg := fmt.Sprintf("✅ %s を %s に変更しました\n⚠️ ボットを再起動してください", varName, value)
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: successMsg,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

// handleConfigGet は "config get" サブコマンドを処理します
func handleConfigGet(s *discordgo.Session, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption) {
	var varName string

	for _, opt := range options {
		if opt.Name == "variable" {
			varName = opt.StringValue()
		}
	}

	log.Printf("DEBUG: Get command - variable: %s", varName)

	currentValue := utils.GetEnv(varName, "(未設定)")
	if currentValue == "" {
		currentValue = "(空)"
	}

	msg := fmt.Sprintf("📋 %s の現在値:\n```\n%s\n```", varName, currentValue)
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: msg,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

// handleConfigList は "config list" サブコマンドを処理します
func handleConfigList(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Println("DEBUG: List command")

	configs := utils.ListConfigVariables()

	if len(configs) == 0 {
		log.Println("DEBUG: No configs found")
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "現在、設定されている変数がありません",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	// 設定情報を構築
	var configList strings.Builder
	configList.WriteString("📋 現在の設定一覧:\n```\n")

	for varName, value := range configs {
		if value == "" {
			value = "(空)"
		}
		configList.WriteString(fmt.Sprintf("%s=%s\n", varName, value))
	}

	configList.WriteString("```")

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: configList.String(),
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}
