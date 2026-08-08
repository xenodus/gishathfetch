package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"mtg-price-checker-sg/handler"
	"mtg-price-checker-sg/pkg/config"
	"mtg-price-checker-sg/pkg/logger"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/joho/godotenv"
)

func init() {
	// load .env file
	err := godotenv.Load()
	logger.Init()
	if err != nil {
		if os.Getenv("ENV") != config.EnvProd {
			slog.Warn("No .env file found or error loading .env")
		}
	}
}

func main() {
	if os.Getenv("ENV") == config.EnvProd {
		lambda.Start(handler.Handle)
	} else {
		start := time.Now()
		ctx := context.Background()
		resp, err := handler.Search(ctx, events.APIGatewayProxyRequest{})
		if err != nil {
			slog.ErrorContext(ctx, "local search failed", "err", err)
		} else {
			slog.InfoContext(ctx, "local search result", "result", resp)
		}
		slog.InfoContext(ctx, "local search duration", "duration", time.Since(start))
	}
}
