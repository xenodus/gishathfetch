package handler

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"mtg-price-checker-sg/controller/ckprice"
	"mtg-price-checker-sg/pkg/config"
	"mtg-price-checker-sg/store/ckprices"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdatypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

const ckPriceRefreshRunAction = "ck-price-refresh-run"

type CKRefreshResponse struct {
	Refreshed int `json:"refreshed"`
}

type CKRefreshAcceptedResponse struct {
	Status string `json:"status"`
}

var (
	newCKRefreshStoreFunc = func(ctx context.Context) (ckprices.Store, error) {
		return ckprices.NewDynamoDBStore(ctx)
	}
	refreshCKPricesFunc = ckprice.RefreshPrices
	enqueueCKPriceRefreshFunc = enqueueCKPriceRefresh
)

func CKPriceRefresh(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var apiRes events.APIGatewayProxyResponse

	if request.HTTPMethod == http.MethodOptions {
		return optionsResponse(request.Headers["origin"])
	}

	if request.HTTPMethod != http.MethodPost {
		return errorResponse(apiRes, "", "method not allowed", http.StatusMethodNotAllowed)
	}

	if !isValidRefreshAPIKey(request.Headers["x-api-key"]) {
		return errorResponse(apiRes, "", "unauthorized", http.StatusUnauthorized)
	}

	if err := enqueueCKPriceRefreshFunc(ctx); err != nil {
		log.Printf("ck price refresh enqueue failed: %v", err)
		return errorResponse(apiRes, "", "err refreshing card kingdom prices", http.StatusInternalServerError)
	}

	return jsonResponse(apiRes, "", http.StatusAccepted, CKRefreshAcceptedResponse{Status: "accepted"})
}

func runCKPriceRefresh(ctx context.Context) error {
	store, err := newCKRefreshStoreFunc(ctx)
	if err != nil {
		return err
	}

	count, err := refreshCKPricesFunc(ctx, store)
	if err != nil {
		log.Printf("ck price mtgjson: %v", err)
		return err
	}

	log.Printf("refreshed %d card kingdom prices", count)
	return nil
}

func enqueueCKPriceRefresh(ctx context.Context) error {
	functionName := strings.TrimSpace(os.Getenv("AWS_LAMBDA_FUNCTION_NAME"))
	if functionName == "" {
		return runCKPriceRefresh(ctx)
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(config.AWSRegion))
	if err != nil {
		return err
	}

	payload, err := json.Marshal(map[string]string{"action": ckPriceRefreshRunAction})
	if err != nil {
		return err
	}

	_, err = lambda.NewFromConfig(cfg).Invoke(ctx, &lambda.InvokeInput{
		FunctionName:   aws.String(functionName),
		InvocationType: lambdatypes.InvocationTypeEvent,
		Payload:        payload,
	})
	return err
}

func isValidRefreshAPIKey(provided string) bool {
	expected := strings.TrimSpace(os.Getenv(config.CKRefreshAPIKeyEnv))
	if expected == "" {
		return os.Getenv("ENV") != config.EnvProd
	}
	return subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) == 1
}
