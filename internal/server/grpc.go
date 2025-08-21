package server

import (
	"net"

	"github.com/VoroniakPavlo/call_audit/auth"
	conf "github.com/VoroniakPavlo/call_audit/config"
	grpcerr "github.com/VoroniakPavlo/call_audit/internal/errors"
	"github.com/VoroniakPavlo/call_audit/internal/server/interceptor"
	"github.com/VoroniakPavlo/call_audit/registry"
	"github.com/VoroniakPavlo/call_audit/registry/consul"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	Server   *grpc.Server
	listener net.Listener
	config   *conf.ConsulConfig
	exitChan chan error
	registry registry.ServiceRegistrator
}

// BuildServer constructs and configures a new gRPC server
func BuildServer(config *conf.ConsulConfig, authManager auth.Manager, exitChan chan error) (*Server, error) {

	// Create a new gRPC server
	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptor.OuterInterceptor(),
			interceptor.AuthUnaryServerInterceptor(authManager),
		),
	)

	// Open a TCP listener on the configured address
	listener, err := net.Listen("tcp", config.PublicAddress)
	if err != nil {
		return nil, grpcerr.NewInternalError("server.build.listen.error", err.Error())
	}

	// Initialize Consul service registry
	reg, err := consul.NewConsulRegistry(config)
	if err != nil {
		return nil, grpcerr.NewInternalError("server.build.consul_registry.error", err.Error())
	}

	// Register gRPC reflection for debugging
	reflection.Register(s)

	return &Server{
		Server:   s,
		listener: listener,
		exitChan: exitChan,
		config:   config,
		registry: reg,
	}, nil
}

// Start registers and starts the gRPC server
func (s *Server) Start() {
	if err := s.registry.Register(); err != nil {
		s.exitChan <- err
		return
	}
	if err := s.Server.Serve(s.listener); err != nil {
		s.exitChan <- grpcerr.NewInternalError("server.start.serve.error", err.Error())
	}
}

// Stop deregisters the service and gracefully stops the gRPC server
func (s *Server) Stop() {
	if err := s.registry.Deregister(); err != nil {
		s.exitChan <- err
		return
	}
	s.Server.Stop()
}
