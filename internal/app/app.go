package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/VoroniakPavlo/call_audit/auth"
	"github.com/VoroniakPavlo/call_audit/auth/manager/webitel_app"
	"github.com/VoroniakPavlo/call_audit/internal/errors"

	conf "github.com/VoroniakPavlo/call_audit/config"
	cerror "github.com/VoroniakPavlo/call_audit/internal/errors"
	"github.com/VoroniakPavlo/call_audit/internal/server"
	"github.com/VoroniakPavlo/call_audit/internal/store"
	"github.com/VoroniakPavlo/call_audit/internal/store/postgres"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	AnonymousName = "Anonymous"
)

func NewBadRequestError(err error) errors.AppError {
	return errors.NewBadRequestError("app.process_api.validation.error", err.Error())
}

type App struct {
	config         *conf.AppConfig
	Store          store.Store
	server         *server.Server
	exitChan       chan error
	storageConn    *grpc.ClientConn
	sessionManager auth.Manager
	webitelAppConn *grpc.ClientConn
	shutdown       func(ctx context.Context) error
	log            *slog.Logger
	rabbitExitChan chan cerror.AppError
	engineConn     *grpc.ClientConn
}

func New(config *conf.AppConfig) (*App, error) {
	// --------- App Initialization ---------
	app := &App{config: config}
	var err error

	// --------- DB Initialization ---------
	if config.Database == nil {
		return nil, cerror.NewInternalError("internal.internal.new.database_config.bad_arguments", "error creating store, config is nil")
	}
	app.Store = BuildDatabase(config.Database)

	// --------- Webitel App gRPC Connection ---------
	app.webitelAppConn, err = grpc.NewClient(fmt.Sprintf("consul://%s/go.webitel.app?wait=14s", config.Consul.Address),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		return nil, cerror.NewInternalError("internal.internal.new_app.grpc_conn.error", err.Error())
	}

	// --------- Session Manager Initialization ---------
	app.sessionManager, err = webitel_app.New(app.webitelAppConn)
	if err != nil {
		return nil, err
	}

	// --------- gRPC Server Initialization ---------
	s, err := server.BuildServer(app.config.Consul, app.sessionManager, app.exitChan)
	if err != nil {
		return nil, err
	}
	app.server = s

	// --------- Service Registration ---------
	RegisterServices(app.server.Server, app)

	// --------- Storage gRPC Connection ---------
	app.storageConn, err = grpc.NewClient(fmt.Sprintf("consul://%s/store?wait=14s", config.Consul.Address),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, cerror.NewInternalError("internal.internal.new_app.grpc_conn.error", err.Error())
	}

	return app, nil
}

func BuildDatabase(config *conf.DatabaseConfig) store.Store {
	return postgres.New(config)
}

func (a *App) Start() error { // Change return type to standard error
	err := a.Store.Open()
	if err != nil {
		return err
	}

	// Jobs execution
	StartJobs(a)

	// * run grpc server
	go a.server.Start()
	return <-a.exitChan
}

func (a *App) Stop() error { // Change return type to standard error
	// close massive modules
	a.server.Stop()
	// close store connection
	a.Store.Close()
	// close grpc connections
	err := a.storageConn.Close()
	if err != nil {
		return err
	}
	err = a.webitelAppConn.Close()
	if err != nil {
		return err
	}

	// ----- Call the shutdown function for OTel ----- //
	if a.shutdown != nil {
		err := a.shutdown(context.Background())
		if err != nil {
			return err
		}
	}

	return nil
}
