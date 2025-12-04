# SysMonitorBot

is a Discord Bot written in Go and Copilot.
CPU、Memoryの使用率を表示します。

![](./about.png)

## Usage

```bash
$ touch .env
$ echo "DISCORD_TOKEN=YourDiscordTokenHere" >> .env

$ go get
$ go run . -interval=5
```
intervalはStatusの更新間隔です。デフォルトは1800秒です。

## Libraries

- [DiscordGo](https://github.com/bwmarrin/discordgo)
- [gopsutil](https://github.com/shirou/gopsutil)
- [dotenv](https://github.com/bkeepers/dotenv)
