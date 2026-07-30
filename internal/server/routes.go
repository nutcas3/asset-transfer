package server

import (
	"net/http"

	httpdelivery "asset-transfer-app/internal/delivery/http"
	"asset-transfer-app/internal/infrastructure/gateway"
	"asset-transfer-app/internal/infrastructure/persistence"
	"asset-transfer-app/internal/usecases"
)

func (s *Server) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()

	// Initialize infrastructure layer (implementations)
	transferRepo := persistence.NewInMemoryTransferRepository()
	gatewayRepo := gateway.NewStubGateway()

	// Wrap gateway with circuit breaker
	circuitBreaker := gateway.NewCircuitBreaker(gatewayRepo)

	// Initialize use case layer (business logic)
	transferUseCase := usecases.NewTransferUseCase(transferRepo, circuitBreaker)

	// Initialize delivery layer (HTTP handlers)
	handler := httpdelivery.NewHandler(transferUseCase)

	// Register transfer routes
	mux.HandleFunc("/v1/transfers", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handler.CreateTransfer(w, r)
		} else {
			http.NotFound(w, r)
		}
	})

	mux.HandleFunc("/v1/transfers/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handler.GetTransfer(w, r)
		} else {
			http.NotFound(w, r)
		}
	})

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	return mux
}
