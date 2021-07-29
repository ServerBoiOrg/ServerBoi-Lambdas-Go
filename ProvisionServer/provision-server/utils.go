package provision

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func getEnvVar(key string) string {

	env := godotenv.Load(".env")

	if env == nil {
		log.Fatalf("Error loading .env file")
	}

	return os.Getenv(key)
}
