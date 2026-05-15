package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/user/local-service-panel/agent/internal/api"
	"github.com/user/local-service-panel/agent/internal/auth"
	"github.com/user/local-service-panel/agent/internal/config"
	"github.com/user/local-service-panel/agent/internal/cors"
	"github.com/user/local-service-panel/agent/internal/customapp"
	"github.com/user/local-service-panel/agent/internal/db"
	"github.com/user/local-service-panel/agent/internal/db/repository"
	"github.com/user/local-service-panel/agent/internal/events"
	"github.com/user/local-service-panel/agent/internal/logging"
	"github.com/user/local-service-panel/agent/internal/service"
	"github.com/user/local-service-panel/agent/internal/settings"
	"github.com/user/local-service-panel/agent/internal/version"
	"github.com/user/local-service-panel/agent/internal/windowsservice"
	"github.com/user/local-service-panel/agent/internal/autostart"
)

func main() {
	showVersion := flag.Bool("version", false, "Print version and exit")
	dataDir := flag.String("data", "", "Data directory (default: .data/ relative to binary)")
	serviceCmd := flag.String("service", "", "Windows Service command: install, uninstall, run")
	flag.Parse()

	if *showVersion {
		fmt.Printf("local-service-panel-agent %s\n", version.Version)
		os.Exit(0)
	}

	// Handle service install/uninstall commands (no agent startup needed)
	switch *serviceCmd {
	case "install":
		if err := service.Install(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to install service: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Service 'LocalServicePanelAgent' installed successfully.")
		return
	case "uninstall":
		if err := service.Uninstall(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to uninstall service: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Service 'LocalServicePanelAgent' uninstalled successfully.")
		return
	case "run":
		// Run as Windows Service
		if err := service.RunAsService(func(ctx context.Context) error {
			return runAgent(ctx, *dataDir)
		}); err != nil {
			fmt.Fprintf(os.Stderr, "Service error: %v\n", err)
			os.Exit(1)
		}
		return
	case "":
		// Default: run in foreground
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
			<-sigCh
			cancel()
		}()

		if err := runAgent(ctx, *dataDir); err != nil {
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown service command: %s\n", *serviceCmd)
		flag.Usage()
		os.Exit(1)
	}
}

// runAgent initializes all dependencies and starts the HTTP server.
// It blocks until ctx is cancelled, then performs graceful shutdown.
func runAgent(ctx context.Context, dataDir string) error {
	// Determine data directory
	if dataDir == "" {
		dataDir = config.DataDir(".data")
	}

	// Load config
	cfgPath := config.ConfigPath(dataDir)
	cfg, err := config.Load(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config from %s: %v\n", cfgPath, err)
		return err
	}

	// Initialize logger
	logDir := ""
	if dataDir != "" {
		logDir = dataDir + "/logs"
	}
	logger, err := logging.New(cfg.Log.Format, string(cfg.Log.Level), logDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		return err
	}

	logger.Info("Starting Local Service Panel Agent v%s", version.Version)
	logger.Info("Data directory: %s", dataDir)
	logger.Info("Config file: %s", cfgPath)

	// Initialize database
	dbPath := db.DBPath(dataDir)
	database, err := db.Open(dbPath)
	if err != nil {
		logger.Error("Failed to initialize database: %v", err)
		return err
	}
	defer database.Close()
	logger.Info("Database initialized: %s", dbPath)

	// Load or generate auth token
	tokenPath := auth.TokenPath(dataDir)
	token, err := auth.LoadOrCreateToken(tokenPath)
	if err != nil {
		logger.Error("Failed to initialize auth token: %v", err)
		return err
	}
	logger.Info("Auth token loaded from: %s", tokenPath)

	// Create HTTP server with auth middleware
	skipPaths := map[string]bool{"/api/healthz": true}
	devToken := os.Getenv("LOCAL_SERVICE_PANEL_DEV_TOKEN")

	// Initialize providers
	svcProvider := windowsservice.NewProvider()

	// Initialize custom app service
	customeAppRepo := repository.NewManagedTargetRepo(database)
	autoProvider := autostart.NewProvider()
	customAppSvc := customapp.NewService(customeAppRepo, dataDir, autoProvider)

	// Initialize event service
	eventRepo := repository.NewEventLogRepo(database)
	eventSvc := events.NewService(eventRepo)

	// Initialize settings store
	settingsStore, err := settings.NewStore(dataDir)
	if err != nil {
		logger.Error("Failed to initialize settings store: %v", err)
		return err
	}

	// Start HTTP server
	apih := api.NewHandler(svcProvider, customAppSvc, eventSvc, settingsStore)
	router := apih.Router()
	wrappedHandler := cors.Middleware(auth.Middleware(token, skipPaths, devToken, router))
	addr := cfg.Addr()
	logger.Info("Listening on %s", addr)

	srv := &http.Server{
		Addr:         addr,
		Handler:      wrappedHandler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server error: %v", err)
		}
	}()

	// Wait for cancellation (signal in foreground mode, SCM stop in service mode)
	<-ctx.Done()

	logger.Info("Shutting down...")

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server shutdown error: %v", err)
	}

	logger.Info("Agent stopped")
	return nil
}
