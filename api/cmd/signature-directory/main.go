package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"mtg-price-checker-sg/pkg/config"
	"mtg-price-checker-sg/pkg/webbotauth"
)

func main() {
	outPath := flag.String("out", "frontend/public/.well-known/http-message-signatures-directory", "output file path")
	flag.Parse()

	pemData := strings.TrimSpace(os.Getenv(config.WebBotAuthPrivateKeyEnv))
	if pemData == "" {
		log.Fatalf("%s is required", config.WebBotAuthPrivateKeyEnv)
	}

	privateKey, err := webbotauth.ParseEd25519PrivateKeyPEM(pemData)
	if err != nil {
		log.Fatalf("parse %s: %v", config.WebBotAuthPrivateKeyEnv, err)
	}

	body, err := webbotauth.DirectoryJSON(privateKey)
	if err != nil {
		log.Fatalf("build signature directory: %v", err)
	}

	if err := os.MkdirAll(filepath.Dir(*outPath), 0o755); err != nil {
		log.Fatalf("create output directory: %v", err)
	}
	if err := os.WriteFile(*outPath, body, 0o644); err != nil {
		log.Fatalf("write %s: %v", *outPath, err)
	}

	fmt.Fprintf(os.Stderr, "wrote %s (%d bytes)\n", *outPath, len(body))
}
