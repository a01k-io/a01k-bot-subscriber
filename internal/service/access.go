package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/a01k-io/a01k-bot-subscriber/config"
	"github.com/redis/go-redis/v9"
)

type Subscription struct {
	Logo                  string    `json:"logo"`
	Product               string    `json:"product"`
	DateStartSubscription time.Time `json:"date_start_subscription"`
	DateEndSubscription   time.Time `json:"date_end_subscription"`
}

type GlobalUserResponseAPI struct {
	ID            int64                    `json:"id"`
	AdminLevel    uint8                    `json:"admin_level,omitempty"`
	UserID        int64                    `json:"user_id"`
	BotID         int64                    `json:"bot_id"`
	Username      string                   `json:"user_name"`
	FirstName     string                   `json:"first_name"`
	LastName      string                   `json:"last_name"`
	Language      string                   `json:"language"`
	Subscriptions map[string]*Subscription `json:"subscriptions"`
	Integrations  map[string]string        `json:"integrations"`
}

type AccessService struct {
	redis      *redis.Client
	httpClient *http.Client
	cfg        *config.Config
}

func NewAccessService(redisClient *redis.Client, cfg *config.Config) *AccessService {
	return &AccessService{
		redis: redisClient,
		cfg:   cfg,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// CheckAccess проверяет доступ пользователя с кешированием
func (s *AccessService) CheckAccess(ctx context.Context, userID int64) (bool, error) {
	cacheKey := fmt.Sprintf("subscription:%d", userID)

	// Проверка в кеше
	cachedData, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		// Данные найдены в кеше
		var userData GlobalUserResponseAPI
		if err := json.Unmarshal([]byte(cachedData), &userData); err == nil {
			return s.checkAccessDates(&userData), nil
		}
	}

	// Данные не найдены в кеше или произошла ошибка, делаем запрос к API
	userData, err := s.fetchUserData(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("failed to fetch user data: %w", err)
	}

	// Сохраняем в кеш
	jsonData, err := json.Marshal(userData)
	if err == nil {
		_ = s.redis.Set(ctx, cacheKey, jsonData, 5*time.Minute).Err()
	}

	return s.checkAccessDates(userData), nil
}

func (s *AccessService) fetchUserData(ctx context.Context, userID int64) (*GlobalUserResponseAPI, error) {
	url := fmt.Sprintf(s.cfg.API+"/api/v2/internal/user/%d?bot_id=5383263408", userID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Sub-Key", s.cfg.XSubKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch user data: %s", string(body))
	}

	var apiResponse GlobalUserResponseAPI
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, err
	}

	return &apiResponse, nil
}

func (s *AccessService) checkAccessDates(userData *GlobalUserResponseAPI) bool {
	if userData.Subscriptions == nil || len(userData.Subscriptions) == 0 {
		return false
	}

	alphaData, ok := userData.Subscriptions["alpha"]
	if !ok {
		return false
	}

	now := time.Now()
	return now.After(alphaData.DateStartSubscription) && now.Before(alphaData.DateEndSubscription)
}
