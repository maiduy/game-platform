package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	route "game-platform/internal/services/masterconfig/api/http"
	"game-platform/internal/platform/app"
	"game-platform/internal/platform/server"
)

func main() {
	// Create context that listens for the interrupt signal from the OS
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Initialize shared application resources
	application, err := app.Initialize()
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}
	defer application.Close()

	// Get port from environment
	port := os.Getenv("MASTERCONFIG_PORT")
	if port == "" {
		port = "4002"
	}

	// Create server with masterconfig routes
	srv := server.NewServer(application, port, "micro-masterconfig-service", func(router *gin.Engine, app *app.App) {
		route.InitMasterConfigRoutes(app.Logger, app.MongoConn, router, app.DynamicConfig)
	})

	// Start server in a goroutine
	go func() {
		if err := srv.Start(); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-ctx.Done()

	// Restore default behavior on the interrupt signal
	stop()
	log.Println("Shutting down gracefully, press Ctrl+C again to force")

	// Create a deadline to wait for
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown the server
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("MasterConfig service exited successfully")
}
