package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/a01k-io/a01k-bot-subscriber/config"
	"github.com/a01k-io/a01k-bot-subscriber/internal/handlers"
	"github.com/a01k-io/a01k-bot-subscriber/internal/repository"
	"github.com/a01k-io/a01k-bot-subscriber/internal/service"
	"github.com/a01k-io/a01k-bot-subscriber/pkg/database"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func main() {
	// Загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Подключение к MySQL
	db, err := database.NewMySQLConnection(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to MySQL: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}
	defer sqlDB.Close()

	// Подключение к Redis
	redisClient, err := database.NewRedisClient(cfg.RedisURI)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	// Инициализация репозиториев
	userRepo := repository.NewUserRepository(db)
	subscriptionRepo := repository.NewSubscriptionRepository(db)

	// Инициализация сервисов
	accessService := service.NewAccessService(redisClient, cfg)
	subscriptionService := service.NewSubscriptionService(userRepo, subscriptionRepo)

	// Инициализация обработчиков
	commandHandlers := handlers.NewCommandHandlers(subscriptionService, accessService)
	callbackHandlers := handlers.NewCallbackHandlers(subscriptionService, accessService)
	messageHandlers := handlers.NewMessageHandlers(subscriptionService, accessService)

	// Создание бота
	opts := []bot.Option{
		bot.WithDefaultHandler(messageHandlers.HandleMessage),
	}

	b, err := bot.New(cfg.BotToken, opts...)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	// Установка команд бота
	ctx := context.Background()
	_, err = b.SetMyCommands(ctx, &bot.SetMyCommandsParams{
		Commands: []models.BotCommand{
			{
				Command:     "sub",
				Description: "Подписаться на сообщение пользователя",
			},
		},
	})
	if err != nil {
		log.Printf("Failed to set bot commands: %v", err)
	}

	// Регистрация обработчиков команд
	b.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact, commandHandlers.HandleStartCommand)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/sub", bot.MatchTypePrefix, commandHandlers.HandleSubCommand)

	// Регистрация обработчика callback_query
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "", bot.MatchTypePrefix, callbackHandlers.HandleCallbackQuery)

	log.Println("Bot started successfully")

	// Graceful shutdown
	ctxWithCancel, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Запуск бота
	go b.Start(ctxWithCancel)

	// Ожидание сигнала завершения
	<-ctxWithCancel.Done()
	log.Println("Shutting down bot...")
}
