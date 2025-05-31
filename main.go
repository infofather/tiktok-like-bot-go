package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	_ "modernc.org/sqlite"
)

var db *sql.DB

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Нет .env файла, продолжаем...")
	}

	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		log.Fatal("BOT_TOKEN не найден в окружении")
	}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	db, err = sql.Open("sqlite", "./queue.db")
	if err != nil {
		log.Fatal(err)
	}
	initDB()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}
		userID := update.Message.From.ID
		args := strings.Fields(update.Message.CommandArguments())

		switch update.Message.Command() {
		case "start":
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Привет! Отправь /queue чтобы посмотреть очередь. /submit [ссылка] — чтобы добавить видео.")
			bot.Send(msg)

		case "queue":
			text := getQueue()
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
			bot.Send(msg)

		case "submit":
			if len(args) == 0 || !strings.HasPrefix(args[0], "http") {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Используй формат: /submit [ссылка]"))
				continue
			}
			if !canSubmit(userID) {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Сначала лайкни минимум 3 чужих видео. /like [ID]"))
				continue
			}
			addToQueue(userID, args[0])
			resetUserLikes(userID)
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "✅ Добавлено в очередь."))

		case "like":
			if len(args) == 0 {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Укажи ID: /like 2"))
				continue
			}
			id, err := strconv.Atoi(args[0])
			if err != nil {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "ID должен быть числом"))
				continue
			}
			msg := confirmLike(userID, id)
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msg))

		case "status":
			count := getUserLikes(userID)
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("👍 Подтверждено лайков: %d/3", count)))
		}
	}
}
