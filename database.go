package main

import (
	"database/sql"
	"fmt"
)

func initDB() {
	db.Exec(`CREATE TABLE IF NOT EXISTS queue (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		link TEXT
	)`)
	db.Exec(`CREATE TABLE IF NOT EXISTS likes (
		user_id INTEGER,
		queue_id INTEGER
	)`)
	db.Exec(`CREATE TABLE IF NOT EXISTS stats (
		user_id INTEGER PRIMARY KEY,
		likes_given INTEGER DEFAULT 0
	)`)
}

func getQueue() string {
	rows, _ := db.Query("SELECT id, link FROM queue ORDER BY id ASC")
	defer rows.Close()
	result := "Очередь видео:\n"
	count := 0
	for rows.Next() {
		var id int
		var link string
		rows.Scan(&id, &link)
		result += fmt.Sprintf("%d. %s\n", id, link)
		count++
	}
	if count == 0 {
		return "Очередь пуста."
	}
	return result
}

func addToQueue(userID int, link string) {
	db.Exec("INSERT INTO queue (user_id, link) VALUES (?, ?)", userID, link)
}

func confirmLike(userID, queueID int) string {
	var targetUser int
	row := db.QueryRow("SELECT user_id FROM queue WHERE id = ?", queueID)
	err := row.Scan(&targetUser)
	if err != nil {
		return "❌ Видео не найдено."
	}
	if targetUser == userID {
		return "❌ Нельзя лайкать своё видео."
	}
	var exists int
	db.QueryRow("SELECT 1 FROM likes WHERE user_id = ? AND queue_id = ?", userID, queueID).Scan(&exists)
	if exists == 1 {
		return "⚠️ Уже лайкнуто."
	}
	db.Exec("INSERT INTO likes (user_id, queue_id) VALUES (?, ?)", userID, queueID)
	db.Exec("INSERT INTO stats (user_id, likes_given) VALUES (?, 1) ON CONFLICT(user_id) DO UPDATE SET likes_given = likes_given + 1")
	return "✅ Лайк подтверждён."
}

func getUserLikes(userID int) int {
	var count int
	db.QueryRow("SELECT likes_given FROM stats WHERE user_id = ?", userID).Scan(&count)
	return count
}

func canSubmit(userID int) bool {
	return getUserLikes(userID) >= 3
}

func resetUserLikes(userID int) {
	db.Exec("INSERT OR REPLACE INTO stats (user_id, likes_given) VALUES (?, 0)", userID)
}
