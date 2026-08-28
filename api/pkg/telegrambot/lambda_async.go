package telegrambot

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

// LambdaAsyncInvoker enqueues price searches via asynchronous Lambda invoke.
type LambdaAsyncInvoker struct {
	client       *lambda.Client
	functionName string
}

// NewLambdaAsyncInvoker builds an invoker for the given function name.
func NewLambdaAsyncInvoker(ctx context.Context, functionName string) (*LambdaAsyncInvoker, error) {
	functionName = strings.TrimSpace(functionName)
	if functionName == "" {
		return nil, fmt.Errorf("lambda function name is required")
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	return &LambdaAsyncInvoker{
		client:       lambda.NewFromConfig(cfg),
		functionName: functionName,
	}, nil
}

// EnqueuePriceSearch invokes the Lambda asynchronously with a price-run payload.
func (i *LambdaAsyncInvoker) EnqueuePriceSearch(ctx context.Context, chatID int64, query string) error {
	return i.enqueue(ctx, PriceRunAction, chatID, query)
}

// EnqueueCKSearch invokes the Lambda asynchronously with a ck-run payload.
func (i *LambdaAsyncInvoker) EnqueueCKSearch(ctx context.Context, chatID int64, query string) error {
	return i.enqueue(ctx, CKRunAction, chatID, query)
}

func (i *LambdaAsyncInvoker) enqueue(ctx context.Context, action string, chatID int64, query string) error {
	payload, err := json.Marshal(PriceRunEvent{
		Action: action,
		ChatID: chatID,
		Query:  query,
	})
	if err != nil {
		return err
	}

	_, err = i.client.Invoke(ctx, &lambda.InvokeInput{
		FunctionName:   aws.String(i.functionName),
		InvocationType: types.InvocationTypeEvent,
		Payload:        payload,
	})
	return err
}

// NewServiceFromConfig wires clients and optional async invoke for Lambda webhooks.
func NewServiceFromConfig(ctx context.Context, cfg Config, logger *slog.Logger) (*Service, error) {
	if err := cfg.Valid(); err != nil {
		return nil, err
	}

	gishath := NewGishathClient(cfg.GishathAPIBaseURL, cfg.GishathBotToken, cfg.GishathOriginSecret)
	telegram := NewTelegramAPI(cfg.TelegramBotToken)

	var async PriceSearchAsync
	if fn := strings.TrimSpace(cfg.LambdaFunctionName); fn != "" {
		invoker, err := NewLambdaAsyncInvoker(ctx, fn)
		if err != nil {
			return nil, err
		}
		async = invoker
	}

	return NewService(cfg.WebhookSecret, gishath, telegram, async, logger), nil
}

// WebhookStatusForbidden indicates an invalid webhook secret token.
const WebhookStatusForbidden = http.StatusForbidden

// WebhookStatusBadRequest indicates malformed webhook input.
const WebhookStatusBadRequest = http.StatusBadRequest

// WebhookStatusOK indicates the update was accepted.
const WebhookStatusOK = http.StatusOK

// WebhookStatusInternalError indicates an unexpected processing failure.
const WebhookStatusInternalError = http.StatusInternalServerError
