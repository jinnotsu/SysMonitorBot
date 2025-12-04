# AnonymousBoard

Discord Bot with Golang。  
`Anonymous` + `Auto Delete`みたいなメッセージを投稿します。
## 機能

### 📝 匿名メッセージボード
- ボタンをクリックして匿名でメッセージを投稿
- 投稿されたメッセージは指定時間後に自動削除（0秒～7日で設定可能）
- **0秒を指定すると自動削除を無効化**
- デフォルトの削除時間は24時間
- 全てのメッセージは埋め込み（Embed）で表示

### 🖥️ システムモニタリング（option）
- CPU使用率とメモリ使用率をBotのステータスにリアルタイム表示
- 更新間隔はカスタマイズ可能（デフォルト: 1800秒）
- **環境変数 `SYSTEM_MONITOR_ENABLED` で有効/無効を切り替え可能**

### 🏥 HTTPサーバーでのヘルスチェック（option）
- `/health` エンドポイントでBotの稼働状態を確認可能
    - PaaSで運用する場合`/health`がないと自動で再起動するものがあるため

---

## 🚀 セットアップ

### 前提条件
- Go 1.24以上
- Discord Bot Token（[Discord Developer Portal](https://discord.com/developers/applications)で取得）

### ローカル環境での実行

```bash
# 環境変数ファイルを作成
cp .env.example .env
# .envファイルを編集してDISCORD_TOKENを設定

# 依存関係をインストール
go mod download

# 実行
go run ./cmd -interval=1800
```

### Docker環境での実行

```bash
# 環境変数ファイルを作成
cp .env.example .env
# .envファイルを編集してDISCORD_TOKENを設定

# Docker Composeで起動
docker compose up -d

# ログを確認
docker compose logs -f
```

---

## ⚙️ 環境変数

| 変数名 | 必須 | デフォルト | 説明 |
|--------|:----:|:----------:|------|
| `DISCORD_TOKEN` | ✅ | - | Discord BotのToken |
| `SYSTEM_MONITOR_ENABLED` | - | `true` | システムモニタリングの有効/無効 |
| `ANONYMOUS_BUTTON_CHANNEL_ID` | - | - | 匿名投稿ボタンを設置するチャンネルID |
| `ANONYMOUS_POST_CHANNEL_ID` | - | - | 匿名メッセージを投稿するチャンネルID |
| `ANONYMOUS_MESSAGE_DELETE_SECONDS` | - | `86400` | メッセージ自動削除までの秒数（0で自動削除無効） |
| `HEALTH_CHECK_ENABLED` | - | `true` | ヘルスチェックサーバーの有効/無効 |
| `PORT` | - | `8000` | ヘルスチェックサーバーのポート番号 |

### .env ファイルの例

```env
DISCORD_TOKEN=your_discord_bot_token_here

# システムモニタリング設定（オプション）
SYSTEM_MONITOR_ENABLED=true

# 匿名ボード設定（オプション）
ANONYMOUS_BUTTON_CHANNEL_ID=123456789012345678
ANONYMOUS_POST_CHANNEL_ID=123456789012345679
ANONYMOUS_MESSAGE_DELETE_SECONDS=86400  # 0にすると自動削除無効

# ヘルスチェック設定（オプション）
HEALTH_CHECK_ENABLED=true
PORT=8000
```

---

## 🛠️ コマンドラインオプション

| オプション | デフォルト | 説明 |
|------------|:----------:|------|
| `-interval` | `1800` | ステータス更新間隔（秒） |

```bash
# 例: 5分（300秒）ごとに更新
go run ./cmd -interval=300
```

---

## 📁 プロジェクト構成

```
SysMonitorBot/
├── cmd/
│   └── main.go              # エントリーポイント
├── internal/
│   ├── handlers/
│   │   ├── anonymousboard.go # 匿名ボード機能
│   │   └── ping.go           # Pingコマンド
│   ├── server/
│   │   └── health.go         # ヘルスチェックサーバー
│   └── services/
│       └── status.go         # ステータス更新サービス
├── docker-compose.yml
├── Dockerfile
├── go.mod
├── go.sum
└── README.md
```

---

## 📦 使用ライブラリ

| ライブラリ | 用途 |
|------------|------|
| [discordgo](https://github.com/bwmarrin/discordgo) | Discord API クライアント |
| [gopsutil](https://github.com/shirou/gopsutil) | システム情報取得 |
| [godotenv](https://github.com/joho/godotenv) | 環境変数の読み込み |

---
