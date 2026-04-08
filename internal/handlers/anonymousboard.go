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
	// メンション入力のカスタムID
	AnonymousMentionInputID = "anonymous_mention_input"
	// Embed色入力のカスタムID
	AnonymousColorInputID = "anonymous_color_input"
	// デフォルトの削除時間（秒）
	DefaultDeleteSeconds = 86400 // 24時間
	// 最大削除時間（秒）
	MaxDeleteSeconds = 604800 // 7日
	// デフォルトの埋め込み色
	DefaultEmbedColor = 0x5865F2 // Discord Blurple
)

// getDeleteDuration は環境変数から削除までの時間を取得します
// 0を返す場合は自動削除が無効であることを示します
func getDeleteDuration() time.Duration {
	secondsStr := os.Getenv("ANONYMOUS_MESSAGE_DELETE_SECONDS")
	if secondsStr == "" {
		return time.Duration(DefaultDeleteSeconds) * time.Second
	}

	seconds, err := strconv.Atoi(secondsStr)
	if err != nil || seconds < 0 {
		log.Printf("Warning: Invalid ANONYMOUS_MESSAGE_DELETE_SECONDS value '%s', using default %d seconds", secondsStr, DefaultDeleteSeconds)
		return time.Duration(DefaultDeleteSeconds) * time.Second
	}

	// 0の場合は自動削除を無効化（0を返す）
	if seconds == 0 {
		return 0
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

// parseColorCode は16進数カラーコード（#RRGGBB または RRGGBB）をint型に変換します
// 入力が有効でない場合は-1を返します
func parseColorCode(colorStr string) int {
	if colorStr == "" {
		return DefaultEmbedColor
	}

	// ハッシュを削除
	if len(colorStr) > 0 && colorStr[0] == '#' {
		colorStr = colorStr[1:]
	}

	// 6文字の16進数コードであることを確認
	if len(colorStr) != 6 {
		return -1
	}

	// 16進数形式のバリデーション
	color, err := strconv.ParseInt(colorStr, 16, 32)
	if err != nil {
		return -1
	}

	return int(color)
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
	var descriptionText string
	if deleteDuration == 0 {
		descriptionText = "投稿されたメッセージは自動削除されません。"
	} else {
		deleteTimeStr := formatDuration(deleteDuration)
		descriptionText = fmt.Sprintf("投稿されたメッセージは%s後に自動削除されます。", deleteTimeStr)
	}

	// ボタン付きメッセージを送信
	_, err := s.ChannelMessageSendComplex(buttonChannelID, &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{
			{
				Title:       "<a:noted:1446011172754161788> 匿名メッセージボード",
				Description: descriptionText,
				Color:       DefaultEmbedColor,
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

	// プレースホルダーとデフォルト値を設定
	var placeholder string
	var defaultValue string
	if defaultSeconds == 0 {
		placeholder = "0～604800秒（0で自動削除無効）"
		defaultValue = "0"
	} else {
		placeholder = fmt.Sprintf("0～604800秒（デフォルト: %d秒、0で無効）", defaultSeconds)
		defaultValue = strconv.Itoa(defaultSeconds)
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: AnonymousPostModalID,
			Title:    "メッセージを投稿",
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
							CustomID:    AnonymousMentionInputID,
							Label:       "メンション",
							Style:       discordgo.TextInputShort,
							Placeholder: "@everyone, @here, @ロール名 など",
							Required:    false,
							MinLength:   0,
							MaxLength:   100,
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    AnonymousDeleteTimeInputID,
							Label:       "削除までの時間（秒）",
							Style:       discordgo.TextInputShort,
							Placeholder: placeholder,
							Required:    false,
							MinLength:   0,
							MaxLength:   7,
							Value:       defaultValue,
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    AnonymousColorInputID,
							Label:       "Embed の色",
							Style:       discordgo.TextInputShort,
							Placeholder: "#5865F2",
							Required:    false,
							MinLength:   0,
							MaxLength:   7,
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
	var mentionContent string
	var colorStr string
	for _, comp := range data.Components {
		if row, ok := comp.(*discordgo.ActionsRow); ok {
			for _, rowComp := range row.Components {
				if textInput, ok := rowComp.(*discordgo.TextInput); ok {
					switch textInput.CustomID {
					case AnonymousPostInputID:
						messageContent = textInput.Value
					case AnonymousDeleteTimeInputID:
						deleteTimeStr = textInput.Value
					case AnonymousMentionInputID:
						mentionContent = textInput.Value
					case AnonymousColorInputID:
						colorStr = textInput.Value
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
		if seconds < 0 {
			respondWithError(s, i, "削除時間は0秒以上で指定してください。")
			return
		}
		if seconds > MaxDeleteSeconds {
			respondWithError(s, i, fmt.Sprintf("削除時間は%d秒以下で指定してください。", MaxDeleteSeconds))
			return
		}
		deleteDuration = time.Duration(seconds) * time.Second
	}

	// 色を解析
	embedColor := parseColorCode(colorStr)
	if embedColor == -1 {
		respondWithError(s, i, "Embed の色は16進数カラーコード（#RRGGBB または RRGGBB）で指定してください。")
		return
	}

	// メッセージを埋め込み形式で投稿
	// メンションはEmbed外のContentに追加（Embed内のメンションは通知されないため）
	messageSend := &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{
			{
				Description: messageContent,
				Color:       embedColor,
			},
		},
	}

	// メンションが指定されている場合、Contentに追加しAllowedMentionsを設定
	if mentionContent != "" {
		messageSend.Content = mentionContent
		messageSend.AllowedMentions = &discordgo.MessageAllowedMentions{
			Parse: []discordgo.AllowedMentionType{
				discordgo.AllowedMentionTypeEveryone,
				discordgo.AllowedMentionTypeRoles,
				discordgo.AllowedMentionTypeUsers,
			},
		}
	}

	msg, err := s.ChannelMessageSendComplex(postChannelID, messageSend)
	if err != nil {
		log.Printf("Error: Failed to send anonymous message: %v", err)
		respondWithError(s, i, "メッセージの投稿に失敗しました。")
		return
	}

	// 削除時間が0より大きい場合のみ、自動削除をスケジュール
	var successDescription string
	if deleteDuration > 0 {
		scheduleMessageDeletion(s, postChannelID, msg.ID, deleteDuration)
		deleteTimeDisplayStr := formatDuration(deleteDuration)
		successDescription = fmt.Sprintf("メッセージが投稿されました！\n%s後に自動削除されます。", deleteTimeDisplayStr)
	} else {
		successDescription = "メッセージが投稿されました！\n自動削除は無効です。"
	}

	// 成功レスポンス（Embed）
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "✅ 投稿成功",
					Description: successDescription,
					Color:       0x00FF00, // 緑
				},
			},
			Flags: discordgo.MessageFlagsEphemeral,
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

// respondWithError はエラーメッセージをエフェメラル埋め込みメッセージとして返します
func respondWithError(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "❌ エラー",
					Description: message,
					Color:       0xFF0000, // 赤
				},
			},
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		log.Printf("Error: Failed to respond with error message: %v", err)
	}
}
