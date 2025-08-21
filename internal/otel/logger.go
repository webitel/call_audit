package logging

import (
	"context"

	"go.opentelemetry.io/otel/sdk/resource"
)

// Setup initializes OpenTelemetry with slog logging and returns a shutdown function
func Setup(service *resource.Resource) func(context.Context) error {
	return nil
}
