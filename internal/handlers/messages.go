package handlers

import (
	"context"
	"fmt"
	"log"
	"strconv"

	internalModels "github.com/a01k-io/a01k-bot-subscriber/internal/models"
	"github.com/a01k-io/a01k-bot-subscriber/internal/service"
	"github.com/a01k-io/a01k-bot-subscriber/internal/utils"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type MessageHandlers struct {
	subscriptionService *service.SubscriptionService
	accessService       *service.AccessService
}

func NewMessageHandlers(
	subscriptionService *service.SubscriptionService,
	accessService *service.AccessService,
) *MessageHandlers {
	return &MessageHandlers{
		subscriptionService: subscriptionService,
		accessService:       accessService,
	}
}

// HandleMessage обработка всех сообщений
func (h *MessageHandlers) HandleMessage(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil || update.Message.From == nil {
		return
	}

	msg := update.Message
	chatID := msg.Chat.ID
	chatIDStr := strconv.FormatInt(chatID, 10)

	// Получение пользователя-отправителя
	telegramID := strconv.FormatInt(msg.From.ID, 10)
	var username *string
	if msg.From.Username != "" {
		username = &msg.From.Username
	}

	user, err := h.subscriptionService.GetOrCreateUser(ctx, telegramID, username)
	if err != nil {
		log.Printf("Ошибка обработки сообщений: %v", err)
		return
	}

	// Поиск подписчиков на этого пользователя в текущем чате
	subscriptions, err := h.subscriptionService.GetSubscribersByTarget(ctx, user.ID, chatIDStr)
	if err != nil {
		log.Printf("Ошибка получения подписчиков: %v", err)
		return
	}

	// Для каждого подписчика
	for _, subscription := range subscriptions {
		go h.sendMessageToSubscriber(ctx, b, msg, subscription, user)
	}
}

func (h *MessageHandlers) sendMessageToSubscriber(
	ctx context.Context,
	b *bot.Bot,
	msg *models.Message,
	subscription internalModels.Subscription,
	user *internalModels.User,
) {
	subscriberTelegramID, err := strconv.ParseInt(subscription.Subscriber.TelegramID, 10, 64)
	if err != nil {
		log.Printf("Ошибка парсинга telegram ID подписчика: %v", err)
		return
	}

	// Проверка доступа подписчика
	hasAccess, err := h.accessService.CheckAccess(ctx, subscriberTelegramID)
	if err != nil || !hasAccess {
		return
	}

	// Формирование inline-клавиатуры
	var messageThreadID *int
	if msg.MessageThreadID != 0 {
		messageThreadID = &msg.MessageThreadID
	}
	keyboard := utils.CreateInlineKeyboard(msg.Chat.ID, msg.ID, messageThreadID, int64(user.ID), int64(subscription.Subscriber.ID))

	// Формирование текста заголовка
	senderUsername := "пользователь"
	if msg.From.Username != "" {
		senderUsername = msg.From.Username
	}
	chatTitle := msg.Chat.Title
	if chatTitle == "" {
		chatTitle = "чат"
	}

	// Отправка сообщения подписчику
	if msg.Text != "" {
		// Текстовое сообщение
		messageText := fmt.Sprintf("%s (%s):\n%s", senderUsername, chatTitle, msg.Text)
		_, err = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      subscriberTelegramID,
			Text:        messageText,
			ReplyMarkup: keyboard,
		})
		if err != nil {
			log.Printf("Ошибка отправки сообщения чат: %s, отправитель %d, подписчик %d",
				subscription.ChatID, user.ID, subscription.Subscriber.ID)
		}
	} else if len(msg.Photo) > 0 {
		// Фото
		fileID := msg.Photo[len(msg.Photo)-1].FileID
		caption := fmt.Sprintf("%s (%s): ", senderUsername, chatTitle)
		if msg.Caption != "" {
			caption += msg.Caption
		} else {
			caption += "отправил фото"
		}

		_, err = b.SendPhoto(ctx, &bot.SendPhotoParams{
			ChatID:      subscriberTelegramID,
			Photo:       &models.InputFileString{Data: fileID},
			Caption:     caption,
			ReplyMarkup: keyboard,
		})
		if err != nil {
			log.Printf("Ошибка отправки фото чат: %s, отправитель %d, подписчик %d",
				subscription.ChatID, user.ID, subscription.Subscriber.ID)
		}
	}
}
