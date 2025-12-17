package main

import (
	"log"
	"log/slog"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	fiberlogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/pricecompare/api/internal/compliance/robots"
	"github.com/pricecompare/api/internal/config"
	"github.com/pricecompare/api/internal/handlers"
	"github.com/pricecompare/api/internal/httpclient"
	"github.com/pricecompare/api/internal/jobs"
	"github.com/pricecompare/api/internal/providers"
	"github.com/pricecompare/api/internal/repository"
	"github.com/pricecompare/api/internal/shipping"
)

func main() {
	// Load .env file if exists
	_ = godotenv.Load()

	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer logger.Sync()

	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := repository.NewDB(cfg.DatabaseURL())
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Initialize Redis for asynq
	redisOpt := asynq.RedisClientOpt{
		Addr:     cfg.RedisAddr(),
		Password: cfg.RedisPassword,
	}
	asynqClient := asynq.NewClient(redisOpt)
	defer asynqClient.Close()

	asynqServer := asynq.NewServer(redisOpt, asynq.Config{
		Concurrency: 10,
	})

	// Initialize Redis client for httpclient (robots.txt cache)
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr(),
		Password: cfg.RedisPassword,
		DB:       0,
	})
	defer redisClient.Close()

	// Create slog logger for httpclient (structured logging)
	slogLogger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Initialize HTTP client with compliance features
	httpClientCfg := httpclient.LoadConfig()
	httpClient := httpclient.New(httpClientCfg, slogLogger, robots.NewRedisCache(redisClient))

	// Initialize repositories
	productRepo := repository.NewProductRepository(db)
	offerRepo := repository.NewOfferRepository(db)
	identifierRepo := repository.NewProductIdentifierRepository(db)
	sourceProductRepo := repository.NewSourceProductRepository(db)

	// Initialize providers
	providerManager := providers.NewManager()

	// Demo / PublicHTML providers are development-only. They can be enabled explicitly
	// via ENABLE_DEMO_PROVIDERS=true.
	if os.Getenv("ENABLE_DEMO_PROVIDERS") == "true" {
		providerManager.Register("demo", providers.NewDemoProvider())
		providerManager.Register("public_html", providers.NewPublicHTMLProvider(cfg.UserAgent))
	}

	// Live provider is the only provider intended for production use.
	providerManager.Register("live", providers.NewLiveProvider(httpClient))

	// Initialize shipping calculator
	shippingConfig := cfg.ShippingConfig()
	shippingCalc := shipping.NewCalculator(shipping.Config{
		Mode:       shippingConfig.Mode,
		FeePercent: shippingConfig.FeePercent,
		FXUSDJPY:   shippingConfig.FXUSDJPY,
	})

	// Initialize job processor
	jobProcessor := jobs.NewProcessor(productRepo, offerRepo, providerManager, shippingCalc, logger)
	mux := asynq.NewServeMux()
	mux.HandleFunc(jobs.TypeFetchPrices, jobProcessor.HandleFetchPrices)

	// Start job processor in background
	go func() {
		if err := asynqServer.Run(mux); err != nil {
			logger.Fatal("Failed to start job processor", zap.Error(err))
		}
	}()

	// Initialize handlers
	h := handlers.New(
		productRepo,
		offerRepo,
		identifierRepo,
		sourceProductRepo,
		providerManager,
		asynqClient,
		shippingCalc,
		logger,
	)

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: handlers.ErrorHandler,
	})

	// Middleware
	app.Use(recover.New())
	app.Use(fiberlogger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,OPTIONS",
		AllowHeaders: "Content-Type",
	}))

	// Routes
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"name":    "Price Compare API",
			"version": "1.0.0",
			"status":  "running",
			"endpoints": fiber.Map{
				"health":  "/health",
				"search":  "/api/search?query=<keyword>",
				"product": "/api/products/:id",
				"offers":  "/api/products/:id/offers",
				"admin":   "/api/admin/jobs/fetch_prices",
			},
		})
	})
	app.Get("/health", h.Health)

	api := app.Group("/api")
	{
		api.Get("/search", h.Search)
		api.Get("/products/:id", h.GetProduct)
		api.Get("/products/:id/offers", h.GetProductOffers)
		api.Get("/products/:id/compare", h.CompareProductOffers)
		api.Post("/resolve-url", h.ResolveURL)
		api.Post("/admin/jobs/fetch_prices", h.FetchPrices)
		api.Post("/image-search", h.ImageSearch) // Stub
	}

	// Start server
	addr := ":" + os.Getenv("API_PORT")
	if addr == ":" {
		addr = ":8080"
	}

	logger.Info("Starting server", zap.String("addr", addr))
	if err := app.Listen(addr); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
