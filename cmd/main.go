package cmd

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	conf "github.com/webitel/call_audit/config"
	"github.com/webitel/call_audit/internal/app"
)

func Run() {

	// Load configuration
	config, appErr := conf.LoadConfig()
	if appErr != nil {
		slog.Error("call_audit.main.configuration_error", slog.String("error", appErr.Error()))
		return
	}

	// Initialize the application
	application, appErr := app.New(config)
	if appErr != nil {
		slog.Error("call_audit.main.application_initialization_error", slog.String("error", appErr.Error()))
		return
	}

	// Initialize signal handling for graceful shutdown
	initSignals(application)

	// Log the configuration
	slog.Debug("call_audit.main.configuration_loaded",
		slog.String("data_source", config.Database.Url),
		slog.String("consul", config.Consul.Address),
		slog.String("grpc_address", config.Consul.Address),
		slog.String("consul_id", config.Consul.Id),
	)

	// Start the application
	slog.Info("call_audit.main.starting_application")
	startErr := application.Start()
	if startErr != nil {
		slog.Error("call_audit.main.application_start_error", slog.String("error", startErr.Error()))
	} else {
		slog.Info("call_audit.main.application_started_successfully")
	}

}

func initSignals(application *app.App) {
	slog.Info("call_audit.main.initializing_stop_signals", slog.String("main", "initializing_stop_signals"))
	sigchnl := make(chan os.Signal, 1)
	signal.Notify(sigchnl)

	go func() {
		for {
			s := <-sigchnl
			handleSignals(s, application)
		}
	}()
}

func handleSignals(signal os.Signal, application *app.App) {
	if signal == syscall.SIGTERM || signal == syscall.SIGINT || signal == syscall.SIGKILL {
		err := application.Stop()
		if err != nil {
			return
		}
		slog.Info(
			"call_audit.main.received_kill_signal",
			slog.String(
				"signal",
				signal.String(),
			),
			slog.String(
				"status",
				"service gracefully stopped",
			),
		)
		os.Exit(0)
	}
}
