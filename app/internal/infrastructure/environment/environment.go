package environment

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

func LoadEnvironmentVariables() {
	currentEnv := os.Getenv("ENVIRONMENT")
	if currentEnv == "" {
		currentEnv = "local"
	}

	slog.Info(fmt.Sprintf("Using environment: '%s'", currentEnv))

	envFile := fmt.Sprintf("%s.env", currentEnv)
	err := godotenv.Load(envFile)

	if err != nil {
		slog.Warn(fmt.Sprintf("Could not load %s: %v (using existing environment variables)", envFile, err))
	}
}
