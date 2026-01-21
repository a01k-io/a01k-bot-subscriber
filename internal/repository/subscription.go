package repository

import (
	"context"
	"fmt"

	"github.com/a01k-io/a01k-bot-subscriber/internal/models"
	"gorm.io/gorm"
)

type SubscriptionRepository struct {
	db *gorm.DB
}

func NewSubscriptionRepository(db *gorm.DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

// FindSubscription находит существующую подписку
func (r *SubscriptionRepository) FindSubscription(ctx context.Context, chatID string, subscriberID, targetID int) (*models.Subscription, error) {
	var subscription models.Subscription
	err := r.db.WithContext(ctx).
		Where("chat_id = ? AND subscriber_id = ? AND target_id = ?", chatID, subscriberID, targetID).
		First(&subscription).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find subscription: %w", err)
	}

	return &subscription, nil
}

// CreateSubscription создает новую подписку
func (r *SubscriptionRepository) CreateSubscription(ctx context.Context, subscriberID, targetID int, chatID string) (*models.Subscription, error) {
	subscription := models.Subscription{
		SubscriberID: subscriberID,
		TargetID:     targetID,
		ChatID:       chatID,
	}

	if err := r.db.WithContext(ctx).Create(&subscription).Error; err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	return &subscription, nil
}

// DeleteSubscription удаляет подписку
func (r *SubscriptionRepository) DeleteSubscription(ctx context.Context, targetID, subscriberID int, chatID string) error {
	result := r.db.WithContext(ctx).
		Where("target_id = ? AND subscriber_id = ? AND chat_id = ?", targetID, subscriberID, chatID).
		Delete(&models.Subscription{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete subscription: %w", result.Error)
	}

	return nil
}

// FindSubscriptionsByTarget находит всех подписчиков на пользователя в чате
func (r *SubscriptionRepository) FindSubscriptionsByTarget(ctx context.Context, targetID int, chatID string) ([]models.Subscription, error) {
	var subscriptions []models.Subscription
	err := r.db.WithContext(ctx).
		Preload("Subscriber").
		Where("target_id = ? AND chat_id = ?", targetID, chatID).
		Find(&subscriptions).Error

	if err != nil {
		return nil, fmt.Errorf("failed to find subscriptions: %w", err)
	}

	return subscriptions, nil
}
