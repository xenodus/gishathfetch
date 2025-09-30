package google

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"mtg-price-checker-sg/pkg/config"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/sony/sonyflake"
)

const (
	measurementID = "G-6NRLSYZ9P9"                // Replace with your GA4 Measurement ID
	apiSecretKey  = "GISHATH_MEASUREMENT_API_KEY" // Replace with your Measurement Protocol API secret
	eventName     = "searched_no_result"
)

var (
	measurementIDGenerator *sonyflake.Sonyflake
	measurementAPIBaseUrl  = "https://www.google-analytics.com/mp/collect"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		if os.Getenv("ENV") != config.EnvProd && os.Getenv("ENV") != config.EnvStaging {
			log.Println("No .env file found or error loading .env")
		}
	}
	measurementIDGenerator, err = sonyflake.New(sonyflake.Settings{
		StartTime: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		log.Println("failed to create sonyflake instance")
	}
}

type payload struct {
	ClientID string  `json:"client_id"`
	Events   []event `json:"events"`
}

type event struct {
	Name   string      `json:"name"`
	Params eventParams `json:"params"`
}

type eventParams struct {
	Lgs     string `json:"lgs"`
	Keyword string `json:"q"`
}

func LGSNoResultMeasurement(lgs, keyword string) error {
	apiSecret := os.Getenv(apiSecretKey)

	url := fmt.Sprintf(
		"%s?measurement_id=%s&api_secret=%s",
		measurementAPIBaseUrl, measurementID, apiSecret,
	)

	id, err := measurementIDGenerator.NextID()
	if err != nil {
		return err
	}

	// Construct the payload
	p := payload{
		ClientID: fmt.Sprintf("%d", id), // Use a unique ID per user/session (must be a non-empty string)
		Events: []event{
			{
				Name: eventName,
				Params: eventParams{
					Lgs:     lgs,
					Keyword: keyword,
				},
			},
		},
	}

	// Convert to JSON
	jsonPayload, err := json.Marshal(p)
	if err != nil {
		return err
	}

	// Send the POST request
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
