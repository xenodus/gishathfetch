package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"mtg-price-checker-sg/handler"
	"mtg-price-checker-sg/pkg/config"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/joho/godotenv"
)

func init() {
	// load .env file
	err := godotenv.Load()
	if err != nil {
		if os.Getenv("ENV") != config.EnvProd && os.Getenv("ENV") != config.EnvStaging {
			log.Println("No .env file found or error loading .env")
		}
	}
}

func main() {
	if os.Getenv("ENV") == config.EnvProd || os.Getenv("ENV") == config.EnvStaging {
		lambda.Start(handler.Search)
	} else {
		start := time.Now()
		log.Println(handler.Search(context.Background(), events.APIGatewayProxyRequest{}))
		log.Println(fmt.Sprintf("Took: %s", time.Since(start)))
	}
}
