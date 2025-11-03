package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/sage-x-project/sage-payment-agent-for-demo/blockchain"
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

	// Initialize agent (keys and blockchain)
	agentInit, err := blockchain.NewAgentInitializer(cfg)
	if err != nil {
		logger.Error("Failed to initialize agent: %v", err)
		os.Exit(1)
	}
	defer agentInit.Cleanup()

	// Print agent info
	fmt.Println(agentInit.GetAgentInfo())

	// Register on blockchain if auto-register is enabled
	if err := agentInit.RegisterIfNeeded(); err != nil {
		logger.Error("Failed to register agent: %v", err)
		// Don't exit - continue without blockchain registration
	}

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
