package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/farmanexo/order-service/internal/presentation/http/controllers"
	"github.com/farmanexo/order-service/internal/presentation/http/middlewares"
)

func SetupRoutes(
	orderController *controllers.OrderController,
	authMiddleware *middlewares.AuthMiddleware,
) *chi.Mux {
	r := chi.NewRouter()

	// Global middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "https://farmanexo.pe"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Correlation-ID"},
		ExposedHeaders:   []string{"Link", "X-Correlation-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Use(middlewares.CorrelationID)

	// Swagger
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:4011/swagger/doc.json"),
	))

	// Health
	r.Get("/health", orderController.HealthCheck)
	r.Get("/", orderController.HealthCheck)

	// API v1
	r.Route("/api/v1", func(r chi.Router) {
		// All order endpoints require authentication
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequireAuth)

			// Cart endpoints
			r.Route("/cart", func(r chi.Router) {
				r.Get("/", orderController.GetCart)
				r.Delete("/", orderController.ClearCart)
				r.Post("/items", orderController.AddCartItem)
				r.Put("/items/{item_id}", orderController.UpdateCartItem)
				r.Delete("/items/{item_id}", orderController.RemoveCartItem)
			})

			// Order endpoints (user)
			r.Route("/orders", func(r chi.Router) {
				r.Get("/", orderController.ListMyOrders)
				r.Post("/checkout", orderController.Checkout)
				r.Get("/{order_id}", orderController.GetOrderDetail)
				r.Post("/{order_id}/cancel", orderController.CancelOrder)
			})

			// Pharmacy endpoints (pharmacy_owner)
			r.Route("/pharmacy", func(r chi.Router) {
				r.Use(authMiddleware.RequirePharmacyOwner)

				r.Get("/orders", orderController.ListPharmacyOrders)
				r.Put("/orders/{order_id}/status", orderController.UpdateOrderStatus)
			})

			// Admin endpoints
			r.Route("/admin", func(r chi.Router) {
				r.Use(authMiddleware.RequireAdmin)

				r.Get("/orders", orderController.ListAllOrders)
				r.Get("/orders/stats", orderController.GetOrderStats)
			})
		})
	})

	// API v2 (future)
	r.Route("/api/v2", func(r chi.Router) {
		r.Get("/status", orderController.HealthCheck)
	})

	return r
}
