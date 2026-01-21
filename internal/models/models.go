package models

import (
	"time"
)

// User представляет пользователя Telegram
type User struct {
	ID         int       `gorm:"primaryKey;autoIncrement;column:id"`
	TelegramID string    `gorm:"uniqueIndex;not null;column:telegram_id"`
	Username   *string   `gorm:"type:varchar(255);column:username"`
	CreatedAt  time.Time `gorm:"autoCreateTime;column:created_at"`

	// Связи
	SubscriptionsAsSubscriber []Subscription `gorm:"foreignKey:SubscriberID"`
	SubscriptionsAsTarget     []Subscription `gorm:"foreignKey:TargetID"`
}

// TableName определяет имя таблицы для модели User
func (User) TableName() string {
	return "users"
}

// Subscription представляет подписку одного пользователя на другого
type Subscription struct {
	ID           int       `gorm:"primaryKey;autoIncrement;column:id"`
	SubscriberID int       `gorm:"not null;column:subscriber_id;index:idx_subscriptions_lookup"`
	TargetID     int       `gorm:"not null;column:target_id;index:idx_subscriptions_lookup"`
	ChatID       string    `gorm:"not null;column:chat_id;index:idx_subscriptions_lookup"`
	CreatedAt    time.Time `gorm:"autoCreateTime;column:created_at"`

	// Связи
	Subscriber User `gorm:"foreignKey:SubscriberID"`
	Target     User `gorm:"foreignKey:TargetID"`
}

// TableName определяет имя таблицы для модели Subscription
func (Subscription) TableName() string {
	return "subscriptions"
}
