package utils

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// DeleteMessagesAfterDelay удаляет сообщения через 3 секунды
func DeleteMessagesAfterDelay(b *bot.Bot, chatID int64, messageIDs []int) {
	go func() {
		time.Sleep(3 * time.Second)
		ctx := context.Background()
		for _, messageID := range messageIDs {
			_, err := b.DeleteMessage(ctx, &bot.DeleteMessageParams{
				ChatID:    chatID,
				MessageID: messageID,
			})
			if err != nil {
				log.Printf("Не удалось удалить сообщения: chatId=%d, messageIds=%v, error=%v", chatID, messageIDs, err)
			}
		}
	}()
}

// CreateInlineKeyboard создает inline-клавиатуру с кнопками ссылки и отписки
func CreateInlineKeyboard(chatID int64, messageID int, messageThreadID *int, targetID, subscriberID int64) *models.InlineKeyboardMarkup {
	// Формируем URL сообщения
	messageURL := FormatMessageURL(chatID, messageID, messageThreadID)

	// Формируем callback_data для отписки
	callbackData := fmt.Sprintf("unsubscribe_%d_%d_%d", targetID, subscriberID, chatID)

	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{
					Text: "🔗",
					URL:  messageURL,
				},
				{
					Text:         "❌",
					CallbackData: callbackData,
				},
			},
		},
	}
}

// FormatMessageURL форматирует URL для ссылки на сообщение
func FormatMessageURL(chatID int64, messageID int, messageThreadID *int) string {
	// Преобразуем chatID в формат для ссылки
	// Если chatID отрицательный (группа/супергруппа), убираем -100 префикс
	chatIDStr := strconv.FormatInt(chatID, 10)
	if strings.HasPrefix(chatIDStr, "-100") {
		chatIDStr = strings.TrimPrefix(chatIDStr, "-100")
	}

	// Формируем базовый URL
	url := fmt.Sprintf("https://t.me/c/%s", chatIDStr)

	// Добавляем thread ID если есть
	if messageThreadID != nil && *messageThreadID != 0 {
		url = fmt.Sprintf("%s/%d", url, *messageThreadID)
	}

	// Добавляем message ID
	url = fmt.Sprintf("%s/%d", url, messageID)

	return url
}

// ParseCallbackData парсит callback_data для отписки
func ParseCallbackData(data string) (targetID, subscriberID int, chatID string, err error) {
	if !strings.HasPrefix(data, "unsubscribe_") {
		return 0, 0, "", fmt.Errorf("invalid callback data format")
	}

	parts := strings.Split(data, "_")
	if len(parts) != 4 {
		return 0, 0, "", fmt.Errorf("invalid callback data parts count")
	}

	targetID, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, "", fmt.Errorf("invalid target ID: %w", err)
	}

	subscriberID, err = strconv.Atoi(parts[2])
	if err != nil {
		return 0, 0, "", fmt.Errorf("invalid subscriber ID: %w", err)
	}

	chatID = parts[3]

	return targetID, subscriberID, chatID, nil
}
