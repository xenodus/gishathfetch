package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"mtg-price-checker-sg/pkg/logger"
	"mtg-price-checker-sg/pkg/webbotauth"
)

func main() {
	logger.Init()

	outPath := flag.String("out", "frontend/public/.well-known/http-message-signatures-directory", "output file path")
	flag.Parse()

	pemData, err := webbotauth.LoadPrivateKeyPEM()
	if err != nil {
		slog.Error("load signing key", "err", err)
		os.Exit(1)
	}

	privateKey, err := webbotauth.ParseEd25519PrivateKeyPEM(pemData)
	if err != nil {
		slog.Error("parse signing key", "err", err)
		os.Exit(1)
	}

	body, err := webbotauth.DirectoryJSON(privateKey)
	if err != nil {
		slog.Error("build signature directory", "err", err)
		os.Exit(1)
	}

	if err := os.MkdirAll(filepath.Dir(*outPath), 0o755); err != nil {
		slog.Error("create output directory", "err", err)
		os.Exit(1)
	}
	if err := os.WriteFile(*outPath, body, 0o644); err != nil {
		slog.Error("write signature directory", "path", *outPath, "err", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "wrote %s (%d bytes)\n", *outPath, len(body))
}
