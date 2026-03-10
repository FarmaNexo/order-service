package middlewares

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/farmanexo/order-service/pkg/mediator"
)

func CorrelationID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		correlationID := r.Header.Get("X-Correlation-ID")
		if correlationID == "" {
			correlationID = uuid.New().String()
		}

		ctx := mediator.WithValue(r.Context(), mediator.CorrelationKey, correlationID)
		w.Header().Set("X-Correlation-ID", correlationID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
