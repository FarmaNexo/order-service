// internal/shared/constants/http_status.go
package constants

// HTTPStatusCode define los códigos HTTP estándar
type HTTPStatusCode int

const (
	StatusOK        HTTPStatusCode = 200
	StatusCreated   HTTPStatusCode = 201
	StatusAccepted  HTTPStatusCode = 202
	StatusNoContent HTTPStatusCode = 204

	StatusBadRequest          HTTPStatusCode = 400
	StatusUnauthorized        HTTPStatusCode = 401
	StatusForbidden           HTTPStatusCode = 403
	StatusNotFound            HTTPStatusCode = 404
	StatusMethodNotAllowed    HTTPStatusCode = 405
	StatusConflict            HTTPStatusCode = 409
	StatusUnprocessableEntity HTTPStatusCode = 422
	StatusTooManyRequests     HTTPStatusCode = 429

	StatusInternalServerError HTTPStatusCode = 500
	StatusNotImplemented      HTTPStatusCode = 501
	StatusBadGateway          HTTPStatusCode = 502
	StatusServiceUnavailable  HTTPStatusCode = 503
	StatusGatewayTimeout      HTTPStatusCode = 504
)

func (h HTTPStatusCode) Int() int {
	return int(h)
}

func (h HTTPStatusCode) String() string {
	descriptions := map[HTTPStatusCode]string{
		StatusOK:                  "OK",
		StatusCreated:             "Created",
		StatusAccepted:            "Accepted",
		StatusNoContent:           "No Content",
		StatusBadRequest:          "Bad Request",
		StatusUnauthorized:        "Unauthorized",
		StatusForbidden:           "Forbidden",
		StatusNotFound:            "Not Found",
		StatusMethodNotAllowed:    "Method Not Allowed",
		StatusConflict:            "Conflict",
		StatusUnprocessableEntity: "Unprocessable Entity",
		StatusTooManyRequests:     "Too Many Requests",
		StatusInternalServerError: "Internal Server Error",
		StatusNotImplemented:      "Not Implemented",
		StatusBadGateway:          "Bad Gateway",
		StatusServiceUnavailable:  "Service Unavailable",
		StatusGatewayTimeout:      "Gateway Timeout",
	}

	if desc, ok := descriptions[h]; ok {
		return desc
	}
	return "Unknown Status"
}

// MessageType define tipos de mensajes
type MessageType string

const (
	MessageTypeInformation MessageType = "INFORMATION"
	MessageTypeSuccess     MessageType = "SUCCESS"
	MessageTypeWarning     MessageType = "WARNING"
	MessageTypeError       MessageType = "ERROR"
)
