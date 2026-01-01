package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/yaroslav/elias/internal/bot"
	"github.com/yaroslav/elias/internal/config"
	"github.com/yaroslav/elias/internal/handlers"
	"github.com/yaroslav/elias/internal/middleware"
	"github.com/yaroslav/elias/internal/services"
	"github.com/yaroslav/elias/internal/ws"
)

func main() {
	cfg := config.Load()

	// Database connection
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Unable to ping database: %v", err)
	}
	log.Println("Connected to PostgreSQL")

	// Redis connection
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
	})
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Unable to connect to Redis: %v", err)
	}
	log.Println("Connected to Redis")

	// Services
	roomService := services.NewRoomService(pool)
	gameService := services.NewGameService(pool, rdb)
	wordService := services.NewWordService(pool)

	// WebSocket hub
	hub := ws.NewHub(rdb, gameService, wordService, roomService)
	go hub.Run()

	// Start Telegram bot
	if cfg.TelegramBotToken != "" {
		telegramBot := bot.New(cfg.TelegramBotToken, cfg.AppURL)
		go telegramBot.Start()
	}

	// Fiber app
	app := fiber.New(fiber.Config{
		AppName: "Elias Local",
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-Telegram-Init-Data",
	}))

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// API routes
	api := app.Group("/api")

	// Auth middleware for API
	authMiddleware := middleware.NewTelegramAuth(cfg.TelegramBotToken)

	// Room routes
	roomHandler := handlers.NewRoomHandler(roomService, gameService, wordService, hub)
	rooms := api.Group("/rooms")
	rooms.Post("/", authMiddleware.Validate, roomHandler.CreateRoom)
	rooms.Get("/:id", authMiddleware.Validate, roomHandler.GetRoom)
	rooms.Post("/:id/join", authMiddleware.Validate, roomHandler.JoinRoom)
	rooms.Post("/:id/team", authMiddleware.Validate, roomHandler.ChangeTeam)
	rooms.Post("/:id/start", authMiddleware.Validate, roomHandler.StartGame)
	rooms.Get("/:id/stats", authMiddleware.Validate, roomHandler.GetStats)

	// WebSocket route
	wsHandler := handlers.NewWSHandler(hub, authMiddleware)
	app.Get("/ws/:room", wsHandler.HandleWebSocket)

	// Graceful shutdown
	go func() {
		if err := app.Listen(":" + cfg.ServerPort); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	log.Printf("Server started on port %s", cfg.ServerPort)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	if err := app.Shutdown(); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}
}
