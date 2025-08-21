package app

import (
	"log"

	ca "github.com/webitel/call_audit/api/call_audit"
	"google.golang.org/grpc"
)

// serviceRegistration holds information for initializing and registering a gRPC service.
type serviceRegistration struct {
	init     func(*App) (interface{}, error)                    // Initialization function for *App
	register func(grpcServer *grpc.Server, service interface{}) // Registration function for gRPC server
	name     string                                             // Service name for logging
}

// RegisterServices initializes and registers all necessary gRPC services.
func RegisterServices(grpcServer *grpc.Server, appInstance *App) {
	services := []serviceRegistration{
		{
			init: func(a *App) (interface{}, error) { return NewLanguageProfileService(a) },
			register: func(s *grpc.Server, svc interface{}) {
				ca.RegisterLanguageProfileServiceServer(s, svc.(ca.LanguageProfileServiceServer))
			},
			name: "LanguageProfile",
		},
		{
			init: func(a *App) (interface{}, error) { return NewCallQuestionnaireRuleService(a) },
			register: func(s *grpc.Server, svc interface{}) {
				ca.RegisterCallQuestionnaireRuleServiceServer(s, svc.(ca.CallQuestionnaireRuleServiceServer))
			},
			name: "CallQuestionnaireRule",
		},
	}

	// Initialize and register each service
	for _, service := range services {
		svc, err := service.init(appInstance)
		if err != nil {
			log.Printf("Error initializing %s service: %v", service.name, err)
			continue
		}
		service.register(grpcServer, svc)
		log.Printf("%s service registered successfully", service.name)
	}
}
