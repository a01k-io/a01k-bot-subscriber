package handlers

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/a01k-io/a01k-bot-subscriber/internal/service"
	"github.com/a01k-io/a01k-bot-subscriber/internal/utils"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const AdminID = 435300492

type CommandHandlers struct {
	subscriptionService *service.SubscriptionService
	accessService       *service.AccessService
}

func NewCommandHandlers(
	subscriptionService *service.SubscriptionService,
	accessService *service.AccessService,
) *CommandHandlers {
	return &CommandHandlers{
		subscriptionService: subscriptionService,
		accessService:       accessService,
	}
}

// HandleStartCommand обработка команды /start
func (h *CommandHandlers) HandleStartCommand(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	// Проверка доступа
	hasAccess, err := h.accessService.CheckAccess(ctx, update.Message.From.ID)
	if err != nil || !hasAccess {
		return
	}

	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Привет! Я бот, который поможет тебе подписаться на сообщения пользователей. Для того, чтобы подписаться на сообщения пользователя, напиши /sub в нужном чате.",
	})
}

// HandleSubCommand обработка команды /sub
func (h *CommandHandlers) HandleSubCommand(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	msg := update.Message
	chatID := msg.Chat.ID

	// Проверка доступа
	hasAccess, err := h.accessService.CheckAccess(ctx, msg.From.ID)
	if err != nil || !hasAccess {
		errorMsg, _ := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "У вас нет доступа, попробуйте немного позже",
		})
		if errorMsg != nil {
			utils.DeleteMessagesAfterDelay(b, chatID, []int{msg.ID, errorMsg.ID})
		}
		return
	}

	// Проверка наличия reply_to_message
	if msg.ReplyToMessage == nil || msg.ReplyToMessage.From == nil {
		errorMsg, _ := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "Пожалуйста, ответьте на сообщение пользователя, на которого хотите подписаться и укажите /sub",
		})
		if errorMsg != nil {
			utils.DeleteMessagesAfterDelay(b, chatID, []int{msg.ID, errorMsg.ID})
		}
		return
	}

	// Получение/создание пользователей
	targetTelegramID := strconv.FormatInt(msg.ReplyToMessage.From.ID, 10)
	var targetUsername *string
	if msg.ReplyToMessage.From.Username != "" {
		targetUsername = &msg.ReplyToMessage.From.Username
	}

	fromTelegramID := strconv.FormatInt(msg.From.ID, 10)
	var fromUsername *string
	if msg.From.Username != "" {
		fromUsername = &msg.From.Username
	}

	targetUser, err := h.subscriptionService.GetOrCreateUser(ctx, targetTelegramID, targetUsername)
	if err != nil {
		log.Printf("Failed to get target user: %v", err)
		errorMsg, _ := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "Произошла ошибка при обработке запроса.",
		})
		if errorMsg != nil {
			utils.DeleteMessagesAfterDelay(b, chatID, []int{msg.ID, errorMsg.ID})
		}
		// Отправляем ошибку администратору
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: AdminID,
			Text:   fmt.Sprintf("Error in HandleSubCommand: %v", err),
		})
		return
	}

	fromUser, err := h.subscriptionService.GetOrCreateUser(ctx, fromTelegramID, fromUsername)
	if err != nil {
		log.Printf("Failed to get from user: %v", err)
		errorMsg, _ := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "Произошла ошибка при обработке запроса.",
		})
		if errorMsg != nil {
			utils.DeleteMessagesAfterDelay(b, chatID, []int{msg.ID, errorMsg.ID})
		}
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: AdminID,
			Text:   fmt.Sprintf("Error in HandleSubCommand: %v", err),
		})
		return
	}

	chatIDStr := strconv.FormatInt(chatID, 10)

	// Проверка существующей подписки
	existingSubscription, err := h.subscriptionService.CheckExistingSubscription(ctx, chatIDStr, fromUser.ID, targetUser.ID)
	if err != nil {
		log.Printf("Failed to check existing subscription: %v", err)
	}

	if existingSubscription != nil {
		username := "пользователя"
		if targetUser.Username != nil {
			username = "@" + *targetUser.Username
		}
		errorMsg, _ := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   fmt.Sprintf("Вы уже подписаны на %s", username),
		})
		if errorMsg != nil {
			utils.DeleteMessagesAfterDelay(b, chatID, []int{msg.ID, errorMsg.ID})
		}
		return
	}

	// Проверка самоподписки
	if err := h.subscriptionService.ValidateSubscription(fromUser.ID, targetUser.ID); err != nil {
		errorMsg, _ := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "Вы не можете подписаться сами на себя",
		})
		if errorMsg != nil {
			utils.DeleteMessagesAfterDelay(b, chatID, []int{msg.ID, errorMsg.ID})
		}
		return
	}

	// Отправка уведомления в личные сообщения
	username := "пользователя"
	if targetUser.Username != nil {
		username = "@" + *targetUser.Username
	}

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: msg.From.ID,
		Text:   fmt.Sprintf("Вы подписались на %s", username),
	})
	if err != nil {
		errorMsg, _ := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "Чтобы подписаться, вам нужно перейти в личные сообщения с ботом @a01k_sub_bot и запустить его!",
		})
		if errorMsg != nil {
			utils.DeleteMessagesAfterDelay(b, chatID, []int{errorMsg.ID})
		}
		return
	}

	// Создание подписки
	if err := h.subscriptionService.CreateSubscription(ctx, fromUser.ID, targetUser.ID, chatIDStr); err != nil {
		log.Printf("Failed to create subscription: %v", err)
		errorMsg, _ := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "Произошла ошибка при обработке запроса.",
		})
		if errorMsg != nil {
			utils.DeleteMessagesAfterDelay(b, chatID, []int{msg.ID, errorMsg.ID})
		}
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: AdminID,
			Text:   fmt.Sprintf("Error in HandleSubCommand: %v", err),
		})
		return
	}

	// Отправка подтверждения
	checkMsg, _ := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   "✅",
		ReplyParameters: &models.ReplyParameters{
			MessageID: msg.ID,
		},
	})

	if checkMsg != nil {
		utils.DeleteMessagesAfterDelay(b, chatID, []int{msg.ID, checkMsg.ID})
	}
}
