// internal/shared/constants/message_codes.go
package constants

// MessageCode contiene todos los códigos de respuesta del sistema
type MessageCode string

const (
	// Success codes
	CodeSuccess        MessageCode = "SUCCESS_001"
	CodeCreatedSuccess MessageCode = "SUCCESS_002"
	CodeUpdatedSuccess MessageCode = "SUCCESS_003"
	CodeDeletedSuccess MessageCode = "SUCCESS_004"

	// Order domain codes
	CodeCartRetrieved       MessageCode = "ORD_001"
	CodeCartItemAdded       MessageCode = "ORD_002"
	CodeCartItemUpdated     MessageCode = "ORD_003"
	CodeCartItemRemoved     MessageCode = "ORD_004"
	CodeCartCleared         MessageCode = "ORD_005"
	CodeCheckoutCompleted   MessageCode = "ORD_006"
	CodeOrderRetrieved      MessageCode = "ORD_007"
	CodeOrdersListed        MessageCode = "ORD_008"
	CodeOrderCancelled      MessageCode = "ORD_009"
	CodeOrderStatusUpdated  MessageCode = "ORD_010"
	CodeOrderStatsRetrieved MessageCode = "ORD_011"
	CodePaymentProcessed    MessageCode = "ORD_012"
	CodePaymentFailed       MessageCode = "ORD_013"

	// Validation errors
	CodeValidationError    MessageCode = "VAL_001"
	CodeRequiredField      MessageCode = "VAL_006"
	CodeInvalidQuantity    MessageCode = "VAL_007"
	CodeInvalidPage        MessageCode = "VAL_011"
	CodeInvalidStatus      MessageCode = "VAL_012"
	CodeInvalidTransition  MessageCode = "VAL_013"

	// Authentication errors
	CodeUnauthorized MessageCode = "AUTH_ERR_001"
	CodeInvalidToken MessageCode = "AUTH_ERR_002"
	CodeTokenExpired MessageCode = "AUTH_ERR_003"
	CodeForbidden    MessageCode = "AUTH_ERR_005"

	// Business errors
	CodeProductNotFound       MessageCode = "BUS_001"
	CodeResourceNotFound      MessageCode = "BUS_003"
	CodeCartEmpty             MessageCode = "BUS_020"
	CodeCartItemNotFound      MessageCode = "BUS_021"
	CodeInsufficientStock     MessageCode = "BUS_022"
	CodeOrderNotFound         MessageCode = "BUS_023"
	CodeOrderNotCancellable   MessageCode = "BUS_024"
	CodeOrderNotOwned         MessageCode = "BUS_025"
	CodeInvalidDeliveryMethod MessageCode = "BUS_026"
	CodeInvalidPaymentMethod  MessageCode = "BUS_027"
	CodeAddressNotFound       MessageCode = "BUS_028"
	CodePharmacyNotFound      MessageCode = "BUS_029"

	// Rate limiting
	CodeRateLimitExceeded MessageCode = "RATE_001"

	// System errors
	CodeInternalError        MessageCode = "SYS_001"
	CodeDatabaseError        MessageCode = "SYS_002"
	CodeServiceUnavailable   MessageCode = "SYS_003"
	CodeCacheError           MessageCode = "SYS_005"
	CodeExternalServiceError MessageCode = "SYS_006"
	CodePaymentGatewayError  MessageCode = "SYS_007"
)

// MessageDescription contiene las descripciones predefinidas
var MessageDescription = map[MessageCode]string{
	CodeSuccess:        "Operación exitosa",
	CodeCreatedSuccess: "Recurso creado exitosamente",
	CodeUpdatedSuccess: "Recurso actualizado exitosamente",
	CodeDeletedSuccess: "Recurso eliminado exitosamente",

	CodeCartRetrieved:       "Carrito obtenido exitosamente",
	CodeCartItemAdded:       "Producto agregado al carrito exitosamente",
	CodeCartItemUpdated:     "Cantidad actualizada exitosamente",
	CodeCartItemRemoved:     "Producto eliminado del carrito exitosamente",
	CodeCartCleared:         "Carrito vaciado exitosamente",
	CodeCheckoutCompleted:   "Checkout completado exitosamente",
	CodeOrderRetrieved:      "Orden obtenida exitosamente",
	CodeOrdersListed:        "Órdenes listadas exitosamente",
	CodeOrderCancelled:      "Orden cancelada exitosamente",
	CodeOrderStatusUpdated:  "Estado de orden actualizado exitosamente",
	CodeOrderStatsRetrieved: "Estadísticas obtenidas exitosamente",
	CodePaymentProcessed:    "Pago procesado exitosamente",
	CodePaymentFailed:       "Error procesando el pago",

	CodeValidationError:   "Error de validación",
	CodeRequiredField:     "Campo requerido",
	CodeInvalidQuantity:   "Cantidad inválida",
	CodeInvalidPage:       "Paginación inválida",
	CodeInvalidStatus:     "Estado inválido",
	CodeInvalidTransition: "Transición de estado no permitida",

	CodeUnauthorized: "No autorizado",
	CodeInvalidToken: "Token inválido",
	CodeTokenExpired: "Token expirado",
	CodeForbidden:    "No tiene permisos para esta acción",

	CodeProductNotFound:       "Producto no encontrado",
	CodeResourceNotFound:      "Recurso no encontrado",
	CodeCartEmpty:             "El carrito está vacío",
	CodeCartItemNotFound:      "Item no encontrado en el carrito",
	CodeInsufficientStock:     "Stock insuficiente",
	CodeOrderNotFound:         "Orden no encontrada",
	CodeOrderNotCancellable:   "La orden no puede ser cancelada en su estado actual",
	CodeOrderNotOwned:         "No tiene permisos sobre esta orden",
	CodeInvalidDeliveryMethod: "Método de entrega inválido",
	CodeInvalidPaymentMethod:  "Método de pago inválido",
	CodeAddressNotFound:       "Dirección de entrega no encontrada",
	CodePharmacyNotFound:      "Farmacia no encontrada",

	CodeRateLimitExceeded: "Demasiadas solicitudes. Intente nuevamente más tarde",

	CodeInternalError:        "Error interno del servidor",
	CodeDatabaseError:        "Error de base de datos",
	CodeServiceUnavailable:   "Servicio no disponible",
	CodeCacheError:           "Error en servicio de caché",
	CodeExternalServiceError: "Error en servicio externo",
	CodePaymentGatewayError:  "Error en pasarela de pago",
}

// GetDescription retorna la descripción del código
func GetDescription(code MessageCode) string {
	if desc, ok := MessageDescription[code]; ok {
		return desc
	}
	return "Descripción no disponible"
}
