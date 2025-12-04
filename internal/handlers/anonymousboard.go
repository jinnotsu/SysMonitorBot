package handlers

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
)

const (
	// ボタンのカスタムID
	AnonymousPostButtonID = "anonymous_post_button"
	// モーダルのカスタムID
	AnonymousPostModalID = "anonymous_post_modal"
	// TextInputのカスタムID
	AnonymousPostInputID = "anonymous_post_input"
	// 削除時間入力のカスタムID
	AnonymousDeleteTimeInputID = "anonymous_delete_time_input"
	// デフォルトの削除時間（秒）
	DefaultDeleteSeconds = 86400 // 24時間
	// 最大削除時間（秒）
	MaxDeleteSeconds = 604800 // 7日
)

// getDeleteDuration は環境変数から削除までの時間を取得します
func getDeleteDuration() time.Duration {
	secondsStr := os.Getenv("ANONYMOUS_MESSAGE_DELETE_SECONDS")
	if secondsStr == "" {
		return time.Duration(DefaultDeleteSeconds) * time.Second
	}

	seconds, err := strconv.Atoi(secondsStr)
	if err != nil || seconds <= 0 {
		log.Printf("Warning: Invalid ANONYMOUS_MESSAGE_DELETE_SECONDS value '%s', using default %d seconds", secondsStr, DefaultDeleteSeconds)
		return time.Duration(DefaultDeleteSeconds) * time.Second
	}

	return time.Duration(seconds) * time.Second
}

// formatDuration は時間を人間が読みやすい形式にフォーマットします
func formatDuration(d time.Duration) string {
	if d >= 24*time.Hour {
		days := int(d.Hours() / 24)
		if days == 1 {
			return "24時間"
		}
		return fmt.Sprintf("%d日", days)
	} else if d >= time.Hour {
		hours := int(d.Hours())
		return fmt.Sprintf("%d時間", hours)
	} else if d >= time.Minute {
		minutes := int(d.Minutes())
		return fmt.Sprintf("%d分", minutes)
	}
	return fmt.Sprintf("%d秒", int(d.Seconds()))
}

// SetupAnonymousBoard は指定されたチャンネルに匿名投稿ボタンを設置します
func SetupAnonymousBoard(s *discordgo.Session) {
	// 環境変数からボタンを設置するチャンネルIDを取得
	buttonChannelID := os.Getenv("ANONYMOUS_BUTTON_CHANNEL_ID")
	if buttonChannelID == "" {
		log.Println("Info: ANONYMOUS_BUTTON_CHANNEL_ID is not set, anonymous board feature disabled")
		return
	}

	// 投稿先チャンネルIDを確認
	postChannelID := os.Getenv("ANONYMOUS_POST_CHANNEL_ID")
	if postChannelID == "" {
		log.Println("Info: ANONYMOUS_POST_CHANNEL_ID is not set, anonymous board feature disabled")
		return
	}

	// 既存のボタンメッセージを削除
	deleteExistingButtonMessages(s, buttonChannelID)

	// 削除時間を取得
	deleteDuration := getDeleteDuration()
	deleteTimeStr := formatDuration(deleteDuration)

	// ボタン付きメッセージを送信
	_, err := s.ChannelMessageSendComplex(buttonChannelID, &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{
			{
				Title:       "<a:noted:1446011172754161788> 匿名メッセージボード",
				Description: fmt.Sprintf("投稿されたメッセージは%s後に自動削除されます。", deleteTimeStr),
				Color:       0x5865F2, // Discord Blurple
			},
		},
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    "メッセージを投稿",
						Style:    discordgo.PrimaryButton,
						CustomID: AnonymousPostButtonID,
						Emoji: &discordgo.ComponentEmoji{
							Name:     "right_arrow",
							ID:       "1446022236279541793",
							Animated: true,
						},
					},
				},
			},
		},
	})
	if err != nil {
		log.Printf("Error: Failed to send anonymous board button: %v", err)
		return
	}
	log.Println("Anonymous board button sent successfully")
}

// deleteExistingButtonMessages はチャンネル内の既存のボタンメッセージを削除します
func deleteExistingButtonMessages(s *discordgo.Session, channelID string) {
	// ボットのユーザーIDを取得
	botUserID := s.State.User.ID

	// チャンネル内の最近のメッセージを取得（最大100件）
	messages, err := s.ChannelMessages(channelID, 100, "", "", "")
	if err != nil {
		log.Printf("Warning: Failed to fetch channel messages: %v", err)
		return
	}

	// ボットが送信したボタン付きメッセージを探して削除
	for _, msg := range messages {
		// ボットが送信したメッセージかどうか確認
		if msg.Author.ID != botUserID {
			continue
		}

		// メッセージにコンポーネント（ボタン）が含まれているか確認
		if len(msg.Components) == 0 {
			continue
		}

		// AnonymousPostButtonIDを持つボタンが含まれているか確認
		if containsAnonymousPostButton(msg.Components) {
			err := s.ChannelMessageDelete(channelID, msg.ID)
			if err != nil {
				log.Printf("Warning: Failed to delete existing button message (ID=%s): %v", msg.ID, err)
			} else {
				log.Printf("Deleted existing anonymous board button message (ID=%s)", msg.ID)
			}
		}
	}
}

// containsAnonymousPostButton はコンポーネント内にAnonymousPostButtonIDを持つボタンが含まれているか確認します
func containsAnonymousPostButton(components []discordgo.MessageComponent) bool {
	for _, comp := range components {
		switch c := comp.(type) {
		case *discordgo.ActionsRow:
			for _, rowComp := range c.Components {
				if button, ok := rowComp.(*discordgo.Button); ok {
					if button.CustomID == AnonymousPostButtonID {
						return true
					}
				}
			}
		}
	}
	return false
}

// HandleAnonymousBoardInteraction は匿名ボード関連のインタラクションを処理します
func HandleAnonymousBoardInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionMessageComponent:
		// ボタンクリック時の処理
		if i.MessageComponentData().CustomID == AnonymousPostButtonID {
			handleButtonClick(s, i)
		}
	case discordgo.InteractionModalSubmit:
		// モーダル送信時の処理
		if i.ModalSubmitData().CustomID == AnonymousPostModalID {
			handleModalSubmit(s, i)
		}
	}
}

// handleButtonClick はボタンクリック時にモーダルを表示します
func handleButtonClick(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// デフォルトの削除時間を取得
	defaultDuration := getDeleteDuration()
	defaultSeconds := int(defaultDuration.Seconds())

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: AnonymousPostModalID,
			Title:    "匿名メッセージを投稿",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    AnonymousPostInputID,
							Label:       "メッセージ内容",
							Style:       discordgo.TextInputParagraph,
							Placeholder: "投稿したいメッセージを入力してください...",
							Required:    true,
							MinLength:   1,
							MaxLength:   2000,
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    AnonymousDeleteTimeInputID,
							Label:       "削除までの時間（秒）",
							Style:       discordgo.TextInputShort,
							Placeholder: fmt.Sprintf("1～604800秒（デフォルト: %d秒）", defaultSeconds),
							Required:    false,
							MinLength:   0,
							MaxLength:   7,
							Value:       strconv.Itoa(defaultSeconds),
						},
					},
				},
			},
		},
	})
	if err != nil {
		log.Printf("Error: Failed to respond with modal: %v", err)
	}
}

// handleModalSubmit はモーダル送信時にメッセージを投稿します
func handleModalSubmit(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// 投稿先チャンネルIDを取得
	postChannelID := os.Getenv("ANONYMOUS_POST_CHANNEL_ID")
	if postChannelID == "" {
		respondWithError(s, i, "投稿先チャンネルが設定されていません。")
		return
	}

	// モーダルからメッセージと削除時間を取得
	data := i.ModalSubmitData()
	var messageContent string
	var deleteTimeStr string
	for _, comp := range data.Components {
		if row, ok := comp.(*discordgo.ActionsRow); ok {
			for _, rowComp := range row.Components {
				if textInput, ok := rowComp.(*discordgo.TextInput); ok {
					switch textInput.CustomID {
					case AnonymousPostInputID:
						messageContent = textInput.Value
					case AnonymousDeleteTimeInputID:
						deleteTimeStr = textInput.Value
					}
				}
			}
		}
	}

	if messageContent == "" {
		respondWithError(s, i, "メッセージが空です。")
		return
	}

	// 削除時間を解析
	deleteDuration := getDeleteDuration() // デフォルト値
	if deleteTimeStr != "" {
		seconds, err := strconv.Atoi(deleteTimeStr)
		if err != nil {
			respondWithError(s, i, "削除時間は数字で入力してください。")
			return
		}
		if seconds <= 0 {
			respondWithError(s, i, "削除時間は1秒以上で指定してください。")
			return
		}
		if seconds > MaxDeleteSeconds {
			respondWithError(s, i, fmt.Sprintf("削除時間は%d秒以下で指定してください。", MaxDeleteSeconds))
			return
		}
		deleteDuration = time.Duration(seconds) * time.Second
	}

	deleteTimeDisplayStr := formatDuration(deleteDuration)

	// メッセージを投稿
	msg, err := s.ChannelMessageSend(postChannelID, messageContent)
	if err != nil {
		log.Printf("Error: Failed to send anonymous message: %v", err)
		respondWithError(s, i, "メッセージの投稿に失敗しました。")
		return
	}

	// 指定時間後にメッセージを削除するタイマーを設定
	scheduleMessageDeletion(s, postChannelID, msg.ID, deleteDuration)

	// 成功レスポンス
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("✅ メッセージが投稿されました！%s後に自動削除されます。", deleteTimeDisplayStr),
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		log.Printf("Error: Failed to respond to modal submit: %v", err)
	}
}

// scheduleMessageDeletion は指定時間後にメッセージを削除するスケジュールを設定します
func scheduleMessageDeletion(s *discordgo.Session, channelID, messageID string, duration time.Duration) {
	log.Printf("Scheduled message deletion: Channel=%s, Message=%s, Duration=%v", channelID, messageID, duration)

	time.AfterFunc(duration, func() {
		err := s.ChannelMessageDelete(channelID, messageID)
		if err != nil {
			log.Printf("Error: Failed to delete scheduled message (Channel=%s, Message=%s): %v", channelID, messageID, err)
		} else {
			log.Printf("Successfully deleted scheduled message: Channel=%s, Message=%s", channelID, messageID)
		}
	})
}

// respondWithError はエラーメッセージをエフェメラルメッセージとして返します
func respondWithError(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "❌ " + message,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		log.Printf("Error: Failed to respond with error message: %v", err)
	}
}
