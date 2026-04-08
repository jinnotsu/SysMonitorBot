package utils

import (
	"log"
	"strings"
)

var adminUserIDs map[string]bool

// LoadAdminUsers は環境変数からカンマ区切りの管理者ユーザーIDを読み込み、内部キャッシュに保存します
func LoadAdminUsers(adminIDsEnv string) {
	adminUserIDs = make(map[string]bool)

	if adminIDsEnv == "" {
		log.Println("Info: ADMIN_USER_IDS is not set, all /config commands will be rejected")
		return
	}

	ids := strings.Split(adminIDsEnv, ",")
	for _, id := range ids {
		trimmedID := strings.TrimSpace(id)
		if trimmedID != "" {
			adminUserIDs[trimmedID] = true
		}
	}

	log.Printf("Loaded %d admin user(s)", len(adminUserIDs))
}

// IsAdmin はユーザーIDが管理者かどうかを判定します
func IsAdmin(userID string) bool {
	return adminUserIDs[userID]
}

// GetAdminUsers は現在登録されている管理者ユーザーIDの一覧を返します（ログ用）
func GetAdminUsers() []string {
	var ids []string
	for id := range adminUserIDs {
		ids = append(ids, id)
	}
	return ids
}
