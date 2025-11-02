package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/sage-x-project/sage-payment-agent-for-demo/config"
	"github.com/sage-x-project/sage-payment-agent-for-demo/handlers"
	"github.com/sage-x-project/sage-payment-agent-for-demo/logger"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Set log level
	logger.SetLogLevel(cfg.LogLevel)

	// Log startup information
	logger.LogStartup(cfg.AgentPort, cfg.IsSAGEEnabled())

	// Create payment handler
	paymentHandler := handlers.NewPaymentHandler(cfg)

	// Setup routes
	http.HandleFunc("/payment", paymentHandler.HandlePayment)
	http.HandleFunc("/process", paymentHandler.HandlePayment)  // Alias for /payment
	http.HandleFunc("/health", paymentHandler.HandleHealth)
	http.HandleFunc("/status", paymentHandler.HandleStatus)

	// Root endpoint
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "SAGE Payment Agent v%s\n", cfg.AgentVersion)
		fmt.Fprintf(w, "SAGE Protocol: %v\n", cfg.IsSAGEEnabled())
		fmt.Fprintf(w, "\nEndpoints:\n")
		fmt.Fprintf(w, "  POST /payment - Process payment\n")
		fmt.Fprintf(w, "  GET  /health  - Health check\n")
		fmt.Fprintf(w, "  GET  /status  - Detailed status\n")
	})

	// Start HTTP server in goroutine
	serverAddr := ":" + cfg.AgentPort
	go func() {
		logger.Info("HTTP server listening on %s", serverAddr)
		if err := http.ListenAndServe(serverAddr, nil); err != nil {
			logger.Error("Server error: %v", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	logger.Info("Shutting down gracefully...")
	logger.Info("Payment Agent stopped")
}
