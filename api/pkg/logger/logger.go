package logger

import (
	"context"
	"log/slog"
	"os"

	"mtg-price-checker-sg/pkg/config"

	"github.com/aws/aws-lambda-go/lambdacontext"
)

// Init configures the process-wide default slog logger.
func Init() {
	slog.SetDefault(New())
}

// New returns a logger suitable for the current environment.
func New() *slog.Logger {
	if os.Getenv("ENV") == config.EnvProd {
		return lambdacontext.NewLogger()
	}
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
}

// From returns the default logger. Use context-aware methods (InfoContext, etc.)
// so Lambda request IDs are attached when running in production.
func From(_ context.Context) *slog.Logger {
	return slog.Default()
}
