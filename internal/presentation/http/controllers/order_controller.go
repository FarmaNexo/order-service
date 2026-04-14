package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/farmanexo/order-service/internal/application/commands"
	"github.com/farmanexo/order-service/internal/application/queries"
	"github.com/farmanexo/order-service/internal/presentation/dto/requests"
	"github.com/farmanexo/order-service/internal/presentation/dto/responses"
	"github.com/farmanexo/order-service/internal/presentation/http/middlewares"
	"github.com/farmanexo/order-service/internal/shared/common"
	"github.com/farmanexo/order-service/pkg/mediator"
	"go.uber.org/zap"
)

type OrderController struct {
	mediator *mediator.Mediator
	logger   *zap.Logger
}

func NewOrderController(med *mediator.Mediator, logger *zap.Logger) *OrderController {
	return &OrderController{mediator: med, logger: logger}
}

func (c *OrderController) respondJSON(w http.ResponseWriter, response interface{}) {
	statusCode := http.StatusOK
	if resp, ok := response.(interface{ GetHttpStatus() *int }); ok {
		if httpStatus := resp.GetHttpStatus(); httpStatus != nil {
			statusCode = *httpStatus
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// HealthCheck godoc
// @Summary      Health check del servicio
// @Description  Retorna el estado del servicio
// @Tags         Health
// @Produce      json
// @Success      200  {object}  map[string]interface{}  "Servicio saludable"
// @Router       /health [get]
func (c *OrderController) HealthCheck(w http.ResponseWriter, r *http.Request) {
	type HealthResponse struct {
		Status  string `json:"status" example:"healthy"`
		Service string `json:"service" example:"order-service"`
		Version string `json:"version" example:"1.0.0"`
	}

	health := HealthResponse{
		Status:  "healthy",
		Service: "order-service",
		Version: "1.0.0",
	}

	c.respondJSON(w, common.OkResponse(health))
}

// GetCart godoc
// @Summary      Obtener carrito actual
// @Description  Obtiene el carrito de compras del usuario autenticado
// @Tags         Cart
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  common.ApiResponse[responses.CartResponse]
// @Router       /api/v1/cart [get]
func (c *OrderController) GetCart(w http.ResponseWriter, r *http.Request) {
	userID, _ := middlewares.GetUserIDFromContext(r.Context())

	query := queries.GetCartQuery{UserID: userID}
	response, _ := mediator.Send[queries.GetCartQuery, responses.CartResponse](r.Context(), c.mediator, query)
	c.respondJSON(w, response)
}

// AddCartItem godoc
// @Summary      Agregar producto al carrito
// @Description  Agrega un producto de una farmacia al carrito
// @Tags         Cart
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body  requests.AddCartItemRequest  true  "Datos del item"
// @Success      200  {object}  common.ApiResponse[responses.CartResponse]
// @Router       /api/v1/cart/items [post]
func (c *OrderController) AddCartItem(w http.ResponseWriter, r *http.Request) {
	userID, _ := middlewares.GetUserIDFromContext(r.Context())

	var req requests.AddCartItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.respondJSON(w, common.BadRequestResponse[responses.CartResponse]("VAL_001", "Error en formato de datos: "+err.Error()))
		return
	}

	cmd := commands.AddCartItemCommand{
		UserID:     userID,
		ProductID:  req.ProductID,
		PharmacyID: req.PharmacyID,
		Quantity:   req.Quantity,
	}

	response, _ := mediator.Send[commands.AddCartItemCommand, responses.CartResponse](r.Context(), c.mediator, cmd)
	c.respondJSON(w, response)
}

// UpdateCartItem godoc
// @Summary      Actualizar cantidad de un item
// @Description  Actualiza la cantidad de un item en el carrito
// @Tags         Cart
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        item_id  path  string  true  "ID del item"
// @Param        body     body  requests.UpdateCartItemRequest  true  "Nueva cantidad"
// @Success      200  {object}  common.ApiResponse[responses.CartResponse]
// @Router       /api/v1/cart/items/{item_id} [put]
func (c *OrderController) UpdateCartItem(w http.ResponseWriter, r *http.Request) {
	userID, _ := middlewares.GetUserIDFromContext(r.Context())
	itemID := chi.URLParam(r, "item_id")

	var req requests.UpdateCartItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.respondJSON(w, common.BadRequestResponse[responses.CartResponse]("VAL_001", "Error en formato de datos: "+err.Error()))
		return
	}

	cmd := commands.UpdateCartItemCommand{
		UserID:   userID,
		ItemID:   itemID,
		Quantity: req.Quantity,
	}

	response, _ := mediator.Send[commands.UpdateCartItemCommand, responses.CartResponse](r.Context(), c.mediator, cmd)
	c.respondJSON(w, response)
}

// RemoveCartItem godoc
// @Summary      Eliminar item del carrito
// @Description  Elimina un item específico del carrito
// @Tags         Cart
// @Produce      json
// @Security     BearerAuth
// @Param        item_id  path  string  true  "ID del item"
// @Success      200  {object}  common.ApiResponse[responses.CartResponse]
// @Router       /api/v1/cart/items/{item_id} [delete]
func (c *OrderController) RemoveCartItem(w http.ResponseWriter, r *http.Request) {
	userID, _ := middlewares.GetUserIDFromContext(r.Context())
	itemID := chi.URLParam(r, "item_id")

	cmd := commands.RemoveCartItemCommand{UserID: userID, ItemID: itemID}
	response, _ := mediator.Send[commands.RemoveCartItemCommand, responses.CartResponse](r.Context(), c.mediator, cmd)
	c.respondJSON(w, response)
}

// ClearCart godoc
// @Summary      Vaciar carrito
// @Description  Elimina todos los items del carrito
// @Tags         Cart
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  common.ApiResponse[responses.EmptyResponse]
// @Router       /api/v1/cart [delete]
func (c *OrderController) ClearCart(w http.ResponseWriter, r *http.Request) {
	userID, _ := middlewares.GetUserIDFromContext(r.Context())

	cmd := commands.ClearCartCommand{UserID: userID}
	response, _ := mediator.Send[commands.ClearCartCommand, responses.EmptyResponse](r.Context(), c.mediator, cmd)
	c.respondJSON(w, response)
}

// Checkout godoc
// @Summary      Crear órdenes desde carrito
// @Description  Procesa el checkout: crea órdenes separadas por farmacia y procesa el pago
// @Tags         Orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body  requests.CheckoutRequest  true  "Datos del checkout"
// @Success      201  {object}  common.ApiResponse[responses.CheckoutResponse]
// @Router       /api/v1/orders/checkout [post]
func (c *OrderController) Checkout(w http.ResponseWriter, r *http.Request) {
	userID, _ := middlewares.GetUserIDFromContext(r.Context())
	accessToken, _ := middlewares.GetAccessTokenFromContext(r.Context())

	var req requests.CheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.respondJSON(w, common.BadRequestResponse[responses.CheckoutResponse]("VAL_001", "Error en formato de datos: "+err.Error()))
		return
	}

	cmd := commands.CheckoutCommand{
		UserID:            userID,
		DeliveryMethod:    req.DeliveryMethod,
		DeliveryAddressID: req.DeliveryAddressID,
		PaymentMethod:     req.PaymentMethod,
		PaymentDetails:    commands.PaymentDetails{CardToken: req.PaymentDetails.CardToken},
		Notes:             req.Notes,
		AccessToken:       accessToken,
	}

	response, _ := mediator.Send[commands.CheckoutCommand, responses.CheckoutResponse](r.Context(), c.mediator, cmd)
	c.respondJSON(w, response)
}

// ListMyOrders godoc
// @Summary      Listar mis órdenes
// @Description  Lista las órdenes del usuario autenticado con paginación
// @Tags         Orders
// @Produce      json
// @Security     BearerAuth
// @Param        status  query  string  false  "Filtrar por estado"
// @Param        page    query  int     false  "Página" default(1)
// @Param        limit   query  int     false  "Límite" default(10)
// @Success      200  {object}  common.ApiResponse[responses.OrderListResponse]
// @Router       /api/v1/orders [get]
func (c *OrderController) ListMyOrders(w http.ResponseWriter, r *http.Request) {
	userID, _ := middlewares.GetUserIDFromContext(r.Context())
	status := r.URL.Query().Get("status")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	query := queries.ListMyOrdersQuery{
		UserID: userID,
		Status: status,
		Page:   page,
		Limit:  limit,
	}

	response, _ := mediator.Send[queries.ListMyOrdersQuery, responses.OrderListResponse](r.Context(), c.mediator, query)
	c.respondJSON(w, response)
}

// GetOrderDetail godoc
// @Summary      Ver detalle de una orden
// @Description  Obtiene el detalle completo de una orden
// @Tags         Orders
// @Produce      json
// @Security     BearerAuth
// @Param        order_id  path  string  true  "ID de la orden"
// @Success      200  {object}  common.ApiResponse[responses.OrderDetailResponse]
// @Router       /api/v1/orders/{order_id} [get]
func (c *OrderController) GetOrderDetail(w http.ResponseWriter, r *http.Request) {
	userID, _ := middlewares.GetUserIDFromContext(r.Context())
	orderID := chi.URLParam(r, "order_id")

	query := queries.GetOrderDetailQuery{UserID: userID, OrderID: orderID}
	response, _ := mediator.Send[queries.GetOrderDetailQuery, responses.OrderDetailResponse](r.Context(), c.mediator, query)
	c.respondJSON(w, response)
}

// CancelOrder godoc
// @Summary      Cancelar orden
// @Description  Cancela una orden (solo si está en pending_payment o confirmed)
// @Tags         Orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        order_id  path  string  true  "ID de la orden"
// @Param        body      body  requests.CancelOrderRequest  true  "Motivo de cancelación"
// @Success      200  {object}  common.ApiResponse[responses.OrderDetailResponse]
// @Router       /api/v1/orders/{order_id}/cancel [post]
func (c *OrderController) CancelOrder(w http.ResponseWriter, r *http.Request) {
	userID, _ := middlewares.GetUserIDFromContext(r.Context())
	orderID := chi.URLParam(r, "order_id")

	var req requests.CancelOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.respondJSON(w, common.BadRequestResponse[responses.OrderDetailResponse]("VAL_001", "Error en formato de datos: "+err.Error()))
		return
	}

	cmd := commands.CancelOrderCommand{UserID: userID, OrderID: orderID, Reason: req.Reason}
	response, _ := mediator.Send[commands.CancelOrderCommand, responses.OrderDetailResponse](r.Context(), c.mediator, cmd)
	c.respondJSON(w, response)
}

// ListPharmacyOrders godoc
// @Summary      Ver órdenes de mi farmacia
// @Description  Lista las órdenes de la farmacia del usuario (pharmacy_owner)
// @Tags         Pharmacy
// @Produce      json
// @Security     BearerAuth
// @Param        status  query  string  false  "Filtrar por estado"
// @Param        page    query  int     false  "Página" default(1)
// @Param        limit   query  int     false  "Límite" default(10)
// @Success      200  {object}  common.ApiResponse[responses.OrderListResponse]
// @Router       /api/v1/orders/pharmacy [get]
func (c *OrderController) ListPharmacyOrders(w http.ResponseWriter, r *http.Request) {
	userID, _ := middlewares.GetUserIDFromContext(r.Context())
	status := r.URL.Query().Get("status")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	// In production, resolve pharmacy_id from user's ownership
	// For now, use user_id as pharmacy_id placeholder
	query := queries.ListPharmacyOrdersQuery{
		UserID:     userID,
		PharmacyID: userID, // TODO: resolve from pharmacy ownership
		Status:     status,
		Page:       page,
		Limit:      limit,
	}

	response, _ := mediator.Send[queries.ListPharmacyOrdersQuery, responses.OrderListResponse](r.Context(), c.mediator, query)
	c.respondJSON(w, response)
}

// UpdateOrderStatus godoc
// @Summary      Actualizar estado de orden
// @Description  Actualiza el estado de una orden (pharmacy_owner)
// @Tags         Pharmacy
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        order_id  path  string  true  "ID de la orden"
// @Param        body      body  requests.UpdateOrderStatusRequest  true  "Nuevo estado"
// @Success      200  {object}  common.ApiResponse[responses.OrderDetailResponse]
// @Router       /api/v1/orders/pharmacy/{order_id}/status [put]
func (c *OrderController) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	userID, _ := middlewares.GetUserIDFromContext(r.Context())
	orderID := chi.URLParam(r, "order_id")

	var req requests.UpdateOrderStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.respondJSON(w, common.BadRequestResponse[responses.OrderDetailResponse]("VAL_001", "Error en formato de datos: "+err.Error()))
		return
	}

	cmd := commands.UpdateOrderStatusCommand{
		UserID:     userID,
		OrderID:    orderID,
		Status:     req.Status,
		Notes:      req.Notes,
		PharmacyID: userID, // TODO: resolve from pharmacy ownership
	}

	response, _ := mediator.Send[commands.UpdateOrderStatusCommand, responses.OrderDetailResponse](r.Context(), c.mediator, cmd)
	c.respondJSON(w, response)
}

// ListAllOrders godoc
// @Summary      Ver todas las órdenes (admin)
// @Description  Lista todas las órdenes con filtros (admin)
// @Tags         Admin
// @Produce      json
// @Security     BearerAuth
// @Param        user_id      query  string  false  "Filtrar por usuario"
// @Param        pharmacy_id  query  string  false  "Filtrar por farmacia"
// @Param        status       query  string  false  "Filtrar por estado"
// @Param        date_from    query  string  false  "Fecha desde"
// @Param        date_to      query  string  false  "Fecha hasta"
// @Param        page         query  int     false  "Página" default(1)
// @Param        limit        query  int     false  "Límite" default(10)
// @Success      200  {object}  common.ApiResponse[responses.OrderListResponse]
// @Router       /api/v1/orders/admin [get]
func (c *OrderController) ListAllOrders(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))

	query := queries.ListAllOrdersQuery{
		UserID:     q.Get("user_id"),
		PharmacyID: q.Get("pharmacy_id"),
		Status:     q.Get("status"),
		DateFrom:   q.Get("date_from"),
		DateTo:     q.Get("date_to"),
		Page:       page,
		Limit:      limit,
	}

	response, _ := mediator.Send[queries.ListAllOrdersQuery, responses.OrderListResponse](r.Context(), c.mediator, query)
	c.respondJSON(w, response)
}

// GetOrderStats godoc
// @Summary      Estadísticas de órdenes (admin)
// @Description  Obtiene estadísticas de órdenes
// @Tags         Admin
// @Produce      json
// @Security     BearerAuth
// @Param        date_from  query  string  false  "Fecha desde"
// @Param        date_to    query  string  false  "Fecha hasta"
// @Success      200  {object}  common.ApiResponse[responses.OrderStatsResponse]
// @Router       /api/v1/orders/admin/stats [get]
func (c *OrderController) GetOrderStats(w http.ResponseWriter, r *http.Request) {
	query := queries.GetOrderStatsQuery{
		DateFrom: r.URL.Query().Get("date_from"),
		DateTo:   r.URL.Query().Get("date_to"),
	}

	response, _ := mediator.Send[queries.GetOrderStatsQuery, responses.OrderStatsResponse](r.Context(), c.mediator, query)
	c.respondJSON(w, response)
}
