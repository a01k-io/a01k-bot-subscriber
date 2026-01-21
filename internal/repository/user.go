package repository

import (
	"context"
	"fmt"

	"github.com/a01k-io/a01k-bot-subscriber/internal/models"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// GetOrCreateUser получает или создает пользователя (аналог getUser из JS)
func (r *UserRepository) GetOrCreateUser(ctx context.Context, telegramID string, username *string) (*models.User, error) {
	var user models.User

	// Пытаемся найти пользователя
	err := r.db.WithContext(ctx).Where("telegram_id = ?", telegramID).First(&user).Error
	if err == nil {
		// Пользователь найден
		return &user, nil
	}

	if err != gorm.ErrRecordNotFound {
		// Произошла ошибка базы данных
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// Пользователь не найден, создаем нового
	user = models.User{
		TelegramID: telegramID,
		Username:   username,
	}

	if err := r.db.WithContext(ctx).Create(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &user, nil
}

// GetByTelegramID получает пользователя по Telegram ID
func (r *UserRepository) GetByTelegramID(ctx context.Context, telegramID string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("telegram_id = ?", telegramID).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}
