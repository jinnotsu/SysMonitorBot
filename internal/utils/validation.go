package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ConfigValidation は設定値の型と範囲チェックを行います
type ConfigValidation struct {
	ValidateFunc func(value string) error
	Description  string
}

var configValidations = map[string]ConfigValidation{
	"SYSTEM_MONITOR_ENABLED": {
		ValidateFunc: validateBool,
		Description:  "true または false を入力してください",
	},
	"HEALTH_CHECK_ENABLED": {
		ValidateFunc: validateBool,
		Description:  "true または false を入力してください",
	},
	"PORT": {
		ValidateFunc: validatePort,
		Description:  "1-65535 の整数を入力してください",
	},
	"ANONYMOUS_BUTTON_CHANNEL_ID": {
		ValidateFunc: validateChannelID,
		Description:  "Discord チャンネルID（数値）または空を入力してください",
	},
	"ANONYMOUS_POST_CHANNEL_ID": {
		ValidateFunc: validateChannelID,
		Description:  "Discord チャンネルID（数値）または空を入力してください",
	},
	"ANONYMOUS_MESSAGE_DELETE_SECONDS": {
		ValidateFunc: validateDeleteSeconds,
		Description:  "0-604800 の整数を入力してください（0は削除なし）",
	},
}

// ValidateConfig は指定された設定変数と値を検証します
func ValidateConfig(varName, value string) error {
	validation, exists := configValidations[varName]
	if !exists {
		return fmt.Errorf("不明な設定変数: %s", varName)
	}

	return validation.ValidateFunc(value)
}

// GetConfigDescription は設定変数の説明を返します
func GetConfigDescription(varName string) string {
	validation, exists := configValidations[varName]
	if !exists {
		return ""
	}
	return validation.Description
}

// GetAllConfigVariables は変更可能なすべての設定変数名を返します
func GetAllConfigVariables() []string {
	vars := make([]string, 0, len(configValidations))
	for v := range configValidations {
		vars = append(vars, v)
	}
	return vars
}

// validateBool は値が true または false（大文字小文字問わず）であることを確認します
func validateBool(value string) error {
	lower := strings.ToLower(value)
	if lower != "true" && lower != "false" {
		return fmt.Errorf("true または false を入力してください")
	}
	return nil
}

// validatePort は値が1-65535の整数であることを確認します
func validatePort(value string) error {
	port, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("整数を入力してください")
	}

	if port < 1 || port > 65535 {
		return fmt.Errorf("ポート番号は1-65535の範囲で指定してください")
	}

	return nil
}

// validateChannelID はDiscord チャンネルID形式（数値）またはそれが空であることを確認します
func validateChannelID(value string) error {
	if value == "" {
		return nil // 空は許可（チャンネルID未設定）
	}

	// 数値であることを確認
	if !regexp.MustCompile(`^\d+$`).MatchString(value) {
		return fmt.Errorf("チャンネルIDは数値である必要があります")
	}

	return nil
}

// validateDeleteSeconds は値が0-604800の整数であることを確認します
func validateDeleteSeconds(value string) error {
	seconds, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("整数を入力してください")
	}

	if seconds < 0 || seconds > 604800 {
		return fmt.Errorf("削除時間は0-604800秒の範囲で指定してください（0は削除なし、604800は7日）")
	}

	return nil
}
