// cmd/server/main.go
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/farmanexo/order-service/internal/application/commands"
	"github.com/farmanexo/order-service/internal/application/handlers"
	"github.com/farmanexo/order-service/internal/application/postprocessors"
	"github.com/farmanexo/order-service/internal/application/preprocessors"
	"github.com/farmanexo/order-service/internal/application/validators"
	"github.com/farmanexo/order-service/internal/infrastructure/cache"
	"github.com/farmanexo/order-service/internal/infrastructure/clients"
	"github.com/farmanexo/order-service/internal/infrastructure/messaging"
	"github.com/farmanexo/order-service/internal/infrastructure/payment"
	"github.com/farmanexo/order-service/internal/infrastructure/persistence/postgres"
	"github.com/farmanexo/order-service/internal/infrastructure/security"
	"github.com/farmanexo/order-service/internal/presentation/dto/responses"
	"github.com/farmanexo/order-service/internal/presentation/http/controllers"
	"github.com/farmanexo/order-service/internal/presentation/http/middlewares"
	"github.com/farmanexo/order-service/internal/presentation/http/routes"
	"github.com/farmanexo/order-service/pkg/config"
	"github.com/farmanexo/order-service/pkg/logger"
	"github.com/farmanexo/order-service/pkg/mediator"

	// Swagger docs
	_ "github.com/farmanexo/order-service/docs"

	"go.uber.org/zap"
	pgdriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// @title           FarmaNexo Order Service API
// @version         1.0
// @description     Servicio de gestión de órdenes y carrito de compras para FarmaNexo - Microservicio con CQRS y Clean Architecture
// @termsOfService  https://farmanexo.pe/terms

// @contact.name    FarmaNexo API Support
// @contact.url     https://farmanexo.pe/support
// @contact.email   support@farmanexo.pe

// @license.name    Apache 2.0
// @license.url     http://www.apache.org/licenses/LICENSE-2.0.html

// @host            localhost:4011
// @BasePath        /api/v1

// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 JWT Authorization header using the Bearer scheme. Example: "Bearer {token}"

// @tag.name         Cart
// @tag.description  Endpoints de carrito de compras

// @tag.name         Orders
// @tag.description  Endpoints de órdenes

// @tag.name         Pharmacy
// @tag.description  Endpoints para encargados de farmacia

// @tag.name         Admin
// @tag.description  Endpoints de administración

// @tag.name         Health
// @tag.description  Endpoints de salud del servicio

func main() {
	env := getEnvironment()
	cfg, err := config.LoadConfig(env)
	if err != nil {
		panic(fmt.Sprintf("Error cargando configuración: %v", err))
	}

	zapLogger, err := logger.NewLogger(cfg.Environment, cfg.Log.Encoding, cfg.Log.Level)
	if err != nil {
		panic(fmt.Sprintf("Error inicializando logger: %v", err))
	}
	defer zapLogger.Sync()

	zapLogger.Info("Iniciando Order Service",
		zap.String("environment", cfg.Environment),
		zap.Int("port", cfg.Server.Port),
	)

	db := initDatabase(cfg, zapLogger)

	zapLogger.Info("Auto-migration deshabilitado - Usar migraciones manuales")

	// ========================================
	// REPOSITORIOS
	// ========================================
	cartRepo := postgres.NewCartRepository(db, zapLogger)
	orderRepo := postgres.NewOrderRepository(db, zapLogger)
	orderItemRepo := postgres.NewOrderItemRepository(db, zapLogger)
	statusHistRepo := postgres.NewOrderStatusHistoryRepository(db, zapLogger)

	// ========================================
	// SERVICIOS
	// ========================================
	jwtService := security.NewJWTService(cfg.JWT.Secret, zapLogger)

	// SQS Event Publisher
	eventPublisher, err := messaging.NewSQSEventPublisher(cfg.AWS, cfg.SQS, zapLogger)
	if err != nil {
		zapLogger.Fatal("Error inicializando SQS EventPublisher", zap.Error(err))
	}

	// Redis Cache
	redisClient, err := cache.NewRedisClient(cfg.Redis, zapLogger)
	if err != nil {
		zapLogger.Fatal("Error inicializando Redis", zap.Error(err))
	}
	defer redisClient.Close()

	_ = cache.NewRedisCacheService(redisClient, zapLogger)

	// HTTP Clients
	pharmacyClient := clients.NewPharmacyClient(cfg.Services.PharmacyService.BaseURL, zapLogger)
	_ = clients.NewCatalogClient(cfg.Services.CatalogService.BaseURL, zapLogger)
	userClient := clients.NewUserClient(cfg.Services.UserService.BaseURL, zapLogger)

	// Payment Service
	paymentService := payment.NewMockPaymentService(cfg.Payment.SuccessRate, zapLogger)

	// ========================================
	// SQS CONSUMER
	// ========================================
	if cfg.Consumer.Enabled {
		sqsConsumer, err := messaging.NewSQSConsumer(
			cfg.AWS, cfg.SQS, cfg.Consumer,
			cartRepo, zapLogger,
		)
		if err != nil {
			zapLogger.Fatal("Error inicializando SQS Consumer", zap.Error(err))
		}

		consumerCtx, consumerCancel := context.WithCancel(context.Background())
		defer consumerCancel()
		sqsConsumer.Start(consumerCtx)
		zapLogger.Info("SQS Consumer iniciado para catalog/pharmacy events")
	}

	// ========================================
	// MEDIATOR
	// ========================================
	med := mediator.NewMediator()

	// ========================================
	// HANDLERS
	// ========================================

	// Cart handlers
	getCartHandler := handlers.NewGetCartHandler(cartRepo, zapLogger)
	mediator.RegisterHandler(med, getCartHandler)

	addCartItemHandler := handlers.NewAddCartItemHandler(cartRepo, pharmacyClient, zapLogger)
	mediator.RegisterHandler(med, addCartItemHandler)

	updateCartItemHandler := handlers.NewUpdateCartItemHandler(cartRepo, zapLogger)
	mediator.RegisterHandler(med, updateCartItemHandler)

	removeCartItemHandler := handlers.NewRemoveCartItemHandler(cartRepo, zapLogger)
	mediator.RegisterHandler(med, removeCartItemHandler)

	clearCartHandler := handlers.NewClearCartHandler(cartRepo, zapLogger)
	mediator.RegisterHandler(med, clearCartHandler)

	// Order handlers
	checkoutHandler := handlers.NewCheckoutHandler(
		cartRepo, orderRepo, orderItemRepo, statusHistRepo,
		pharmacyClient, userClient, paymentService, eventPublisher,
		cfg.Delivery, zapLogger,
	)
	mediator.RegisterHandler(med, checkoutHandler)

	cancelOrderHandler := handlers.NewCancelOrderHandler(
		orderRepo, statusHistRepo, paymentService, eventPublisher, zapLogger,
	)
	mediator.RegisterHandler(med, cancelOrderHandler)

	updateOrderStatusHandler := handlers.NewUpdateOrderStatusHandler(
		orderRepo, statusHistRepo, eventPublisher, zapLogger,
	)
	mediator.RegisterHandler(med, updateOrderStatusHandler)

	// Query handlers
	getOrderDetailHandler := handlers.NewGetOrderDetailHandler(orderRepo, zapLogger)
	mediator.RegisterHandler(med, getOrderDetailHandler)

	listMyOrdersHandler := handlers.NewListMyOrdersHandler(orderRepo, zapLogger)
	mediator.RegisterHandler(med, listMyOrdersHandler)

	listPharmacyOrdersHandler := handlers.NewListPharmacyOrdersHandler(orderRepo, zapLogger)
	mediator.RegisterHandler(med, listPharmacyOrdersHandler)

	listAllOrdersHandler := handlers.NewListAllOrdersHandler(orderRepo, zapLogger)
	mediator.RegisterHandler(med, listAllOrdersHandler)

	getOrderStatsHandler := handlers.NewGetOrderStatsHandler(orderRepo, zapLogger)
	mediator.RegisterHandler(med, getOrderStatsHandler)

	// ========================================
	// VALIDATORS
	// ========================================
	addCartItemValidator := validators.NewAddCartItemValidator()
	mediator.RegisterValidator[commands.AddCartItemCommand, responses.CartResponse](med, addCartItemValidator)

	checkoutValidator := validators.NewCheckoutValidator()
	mediator.RegisterValidator[commands.CheckoutCommand, responses.CheckoutResponse](med, checkoutValidator)

	cancelOrderValidator := validators.NewCancelOrderValidator()
	mediator.RegisterValidator[commands.CancelOrderCommand, responses.OrderDetailResponse](med, cancelOrderValidator)

	updateOrderStatusValidator := validators.NewUpdateOrderStatusValidator()
	mediator.RegisterValidator[commands.UpdateOrderStatusCommand, responses.OrderDetailResponse](med, updateOrderStatusValidator)

	// ========================================
	// PREPROCESSORS Y POSTPROCESSORS
	// ========================================
	sanitizePreProcessor := preprocessors.NewSanitizeInputPreProcessor(zapLogger)
	med.RegisterPreProcessor(sanitizePreProcessor)

	auditPostProcessor := postprocessors.NewLogAuditPostProcessor(zapLogger)
	med.RegisterPostProcessor(auditPostProcessor)

	zapLogger.Info("Mediator configurado",
		zap.Int("handlers", 13),
		zap.Int("validators", 4),
		zap.Int("preprocessors", 1),
		zap.Int("postprocessors", 1),
	)

	// ========================================
	// MIDDLEWARES
	// ========================================
	authMiddleware := middlewares.NewAuthMiddleware(jwtService, zapLogger)

	// ========================================
	// CONTROLADORES Y RUTAS
	// ========================================
	orderController := controllers.NewOrderController(med, zapLogger)
	router := routes.SetupRoutes(orderController, authMiddleware)

	// ========================================
	// SERVIDOR HTTP
	// ========================================
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	go func() {
		zapLogger.Info("Servidor HTTP iniciado",
			zap.String("address", server.Addr),
			zap.String("swagger_url", fmt.Sprintf("http://localhost:%d/swagger/index.html", cfg.Server.Port)),
		)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zapLogger.Fatal("Error iniciando servidor", zap.Error(err))
		}
	}()

	// ========================================
	// GRACEFUL SHUTDOWN
	// ========================================
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	zapLogger.Info("Iniciando graceful shutdown...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		zapLogger.Error("Error en shutdown", zap.Error(err))
	}

	zapLogger.Info("Servidor detenido exitosamente")
}

func getEnvironment() string {
	env := os.Getenv("ENV")
	if env == "" {
		env = "local"
	}
	return env
}

func initDatabase(cfg *config.Config, log *zap.Logger) *gorm.DB {
	gormLogLevel := gormlogger.Silent
	if cfg.IsDevelopment() {
		gormLogLevel = gormlogger.Info
	}

	gormLogger := gormlogger.Default.LogMode(gormLogLevel)

	db, err := gorm.Open(pgdriver.Open(cfg.Database.GetDSN()), &gorm.Config{
		Logger: gormLogger,
	})

	if err != nil {
		log.Fatal("Error conectando a PostgreSQL",
			zap.Error(err),
			zap.String("host", cfg.Database.Host),
			zap.Int("port", cfg.Database.Port),
		)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Error obteniendo SQL DB", zap.Error(err))
	}

	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	log.Info("Conexión a PostgreSQL establecida",
		zap.String("host", cfg.Database.Host),
		zap.String("database", cfg.Database.DBName),
	)

	return db
}
