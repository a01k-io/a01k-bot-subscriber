package service

import (
	"context"
	"fmt"

	"github.com/a01k-io/a01k-bot-subscriber/internal/models"
	"github.com/a01k-io/a01k-bot-subscriber/internal/repository"
)

type SubscriptionService struct {
	userRepo         *repository.UserRepository
	subscriptionRepo *repository.SubscriptionRepository
}

func NewSubscriptionService(
	userRepo *repository.UserRepository,
	subscriptionRepo *repository.SubscriptionRepository,
) *SubscriptionService {
	return &SubscriptionService{
		userRepo:         userRepo,
		subscriptionRepo: subscriptionRepo,
	}
}

// GetOrCreateUser получает или создает пользователя
func (s *SubscriptionService) GetOrCreateUser(ctx context.Context, telegramID string, username *string) (*models.User, error) {
	return s.userRepo.GetOrCreateUser(ctx, telegramID, username)
}

// CheckExistingSubscription проверяет существование подписки
func (s *SubscriptionService) CheckExistingSubscription(ctx context.Context, chatID string, subscriberID, targetID int) (*models.Subscription, error) {
	return s.subscriptionRepo.FindSubscription(ctx, chatID, subscriberID, targetID)
}

// CreateSubscription создает новую подписку
func (s *SubscriptionService) CreateSubscription(ctx context.Context, subscriberID, targetID int, chatID string) error {
	_, err := s.subscriptionRepo.CreateSubscription(ctx, subscriberID, targetID, chatID)
	return err
}

// UnsubscribeUser удаляет подписку
func (s *SubscriptionService) UnsubscribeUser(ctx context.Context, targetID, subscriberID int, chatID string) error {
	return s.subscriptionRepo.DeleteSubscription(ctx, targetID, subscriberID, chatID)
}

// GetSubscribersByTarget получает всех подписчиков на пользователя в чате
func (s *SubscriptionService) GetSubscribersByTarget(ctx context.Context, targetID int, chatID string) ([]models.Subscription, error) {
	return s.subscriptionRepo.FindSubscriptionsByTarget(ctx, targetID, chatID)
}

// ValidateSubscription проверяет возможность создания подписки
func (s *SubscriptionService) ValidateSubscription(subscriberID, targetID int) error {
	if subscriberID == targetID {
		return fmt.Errorf("cannot subscribe to yourself")
	}
	return nil
}
