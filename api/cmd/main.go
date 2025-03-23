package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"mtg-price-checker-sg/handler"
	"mtg-price-checker-sg/pkg/config"
)

func main() {
	if os.Getenv("ENV") == config.EnvProd || os.Getenv("ENV") == config.EnvStaging {
		lambda.Start(handler.Search)
	} else {
		start := time.Now()
		log.Println(handler.Search(context.Background(), events.APIGatewayProxyRequest{}))
		log.Println(fmt.Sprintf("Took: %s", time.Since(start)))
	}
}
