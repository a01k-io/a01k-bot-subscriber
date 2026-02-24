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

type SubscriptionResponse struct {
	Id          int       `json:"id"`
	UserID      int       `json:"user_id"`
	Logo        string    `json:"logo"`
	ProductType string    `json:"product_type"`
	StartedAt   time.Time `json:"started_at"`
	ExpiresAt   time.Time `json:"expires_at"`
	Active      bool      `json:"active"`
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
	cacheKey := fmt.Sprintf("subscriptions:%d", userID)

	// Проверка в кеше
	cachedData, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var userData SubscriptionResponse
		if err := json.Unmarshal([]byte(cachedData), &userData); err == nil {
			return s.checkAccessDates(&userData), nil
		}
	}

	userData, err := s.fetchUserData(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("failed to fetch user data: %w", err)
	}

	subscriptionData, err := s.fetchSubscriptions(ctx, userData.ID)
	if err != nil {
		return false, fmt.Errorf("failed to fetch subscriptions: %w", err)
	}

	// Сохраняем в кеш
	jsonData, err := json.Marshal(subscriptionData)
	if err == nil {
		_ = s.redis.Set(ctx, cacheKey, jsonData, 5*time.Minute).Err()
	}

	return s.checkAccessDates(subscriptionData), nil
}

func (s *AccessService) fetchUserData(ctx context.Context, userID int64) (*GlobalUserResponseAPI, error) {
	url := fmt.Sprintf(s.cfg.API+"/api/v2/internal/user/%d?bot_id=5383263408", userID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
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

func (s *AccessService) fetchSubscriptions(ctx context.Context, userID int64) (*SubscriptionResponse, error) {
	url := fmt.Sprintf(s.cfg.API+"/api/v2/internal/user/subscriptions/%d", userID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
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

	var apiResponse []*SubscriptionResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, err
	}

	for _, a := range apiResponse {
		if a.ProductType == "alpha" {
			return a, nil
		}
	}

	return nil, fmt.Errorf("failed to fetch subscriptions: %s", string(body))
}

func (s *AccessService) checkAccessDates(alphaData *SubscriptionResponse) bool {
	now := time.Now()
	return now.After(alphaData.StartedAt) && now.Before(alphaData.ExpiresAt)
}
