package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"

	"github.com/vladislavprovich/TG-bot/internal/handler"
	"github.com/vladislavprovich/TG-bot/internal/keyboard"
	"github.com/vladislavprovich/TG-bot/internal/repository"
	"github.com/vladislavprovich/TG-bot/internal/repository/postgres"
	"github.com/vladislavprovich/TG-bot/internal/service"
	"github.com/vladislavprovich/TG-bot/pkg/logger"
	"github.com/vladislavprovich/TG-bot/pkg/shortener"
)

// Seconds are indicated. Example: "timeOut = 10" --> time to wait for a request is 10 seconds.
const timeOut = 10

func main() {
	ctx := context.Background()
	cfg, err := LoadConfig(ctx)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	bot := initBot(cfg.Bot)
	logger := initLogger(cfg.Logger)
	db, err := initDataBase(ctx, cfg.Database, logger)
	if err != nil {
		logger.Fatalf("Data base error: %v", err)
	}
	defer func() {
		if err = db.Close(); err != nil {
			logger.Errorf("Data base error closing: %v", err)
		}
	}()

	httpClient := &http.Client{
		Timeout: timeOut * time.Second,
	}

	client := initClient(cfg.Client, httpClient, logger)
	urlRepo := initURLRepo(db, logger)
	userRepo := initUserRepo(db, logger)

	params := service.Params{
		urlRepo,
		userRepo,
		logger,
		&client,
	}
	urlService := initService(params)

	buttonHandler := initButtonHandler(urlService, logger)
	messageHandler := initMessageHandler(urlService, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go handler.ProcessUpdates(ctx, bot, buttonHandler, messageHandler, logger)

	ctx, cancel = context.WithCancel(context.Background())

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	<-stop
	logger.Info("Service graceful shutdown...")
}

func initBot(cfg handler.BotConfig) *tgbotapi.BotAPI {
	return handler.BotInit(cfg)
}

func initLogger(cfg logger.ConfigLogger) *logrus.Logger {
	return logger.NewLogger(cfg)
}

func initDataBase(ctx context.Context, config postgres.Config, logger *logrus.Logger) (*sql.DB, error) {
	return postgres.PrepareConnection(ctx, config, logger)
}

func initClient(config *shortener.Config, httpClient *http.Client, logger *logrus.Logger) shortener.BasicClient {
	return shortener.NewBasicClient(config, httpClient, logger)
}

func initURLRepo(db *sql.DB, logger *logrus.Logger) repository.URLRepository {
	return repository.NewBotRepository(db, logger)
}

func initUserRepo(db *sql.DB, logger *logrus.Logger) repository.UserRepository {
	return repository.NewUserRepository(db, logger)
}

func initService(params service.Params) service.URLService {
	return service.NewService(params)
}

func initButtonHandler(service service.URLService, logger *logrus.Logger) *keyboard.HandleButtons {
	return keyboard.NewHandleButtons(service, logger)
}

func initMessageHandler(service service.URLService, logger *logrus.Logger) *keyboard.HandleButtons {
	return keyboard.NewMessageHandler(service, logger)
}
