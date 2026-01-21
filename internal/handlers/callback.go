package handlers

import (
	"context"
	"log"

	"github.com/a01k-io/a01k-bot-subscriber/internal/service"
	"github.com/a01k-io/a01k-bot-subscriber/internal/utils"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type CallbackHandlers struct {
	subscriptionService *service.SubscriptionService
	accessService       *service.AccessService
}

func NewCallbackHandlers(
	subscriptionService *service.SubscriptionService,
	accessService *service.AccessService,
) *CallbackHandlers {
	return &CallbackHandlers{
		subscriptionService: subscriptionService,
		accessService:       accessService,
	}
}

// HandleCallbackQuery обработка callback_query
func (h *CallbackHandlers) HandleCallbackQuery(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.CallbackQuery == nil {
		return
	}

	query := update.CallbackQuery

	// Проверка доступа
	hasAccess, err := h.accessService.CheckAccess(ctx, query.From.ID)
	if err != nil || !hasAccess {
		return
	}

	// Парсинг callback_data
	targetID, subscriberID, chatID, err := utils.ParseCallbackData(query.Data)
	if err != nil {
		log.Printf("Failed to parse callback data: %v", err)
		return
	}

	// Удаление подписки
	err = h.subscriptionService.UnsubscribeUser(ctx, targetID, subscriberID, chatID)
	if err != nil {
		log.Printf("Failed to unsubscribe: %v", err)
		_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
			Text:            "Произошла ошибка при отписке",
			ShowAlert:       true,
		})
		return
	}

	// Отправка подтверждения
	_, _ = b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: query.ID,
		Text:            "Вы успешно отписались от пользователя!",
	})
}
