package main

import (
	"os"

	"cashfac-test/internal/app"
	"cashfac-test/internal/platform/logger"
)

func main() {
	application, err := app.New()
	if err != nil {
		logger.Error("BOOT", "failed to build application", logger.F("error", err))
		os.Exit(1)
	}

	if err := application.Run(); err != nil {
		logger.Error("BOOT", "server stopped", logger.F("error", err))
		os.Exit(1)
	}
}
