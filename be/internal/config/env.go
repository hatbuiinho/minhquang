package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// LoadDotEnv loads the monorepo root env first while keeping direct backend runs working.
func LoadDotEnv() {
	for _, file := range []string{"../.env", ".env"} {
		if err := godotenv.Load(file); err != nil && !os.IsNotExist(err) {
			log.Printf("load %s: %v", file, err)
		}
	}
}
