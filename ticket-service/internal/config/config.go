package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL     string
	RedisURL        string
	MovieServiceURL string
}

func LoadConfig() Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	return Config{
		DatabaseURL:     getEnv("DATABASE_URL"),
		RedisURL:        getEnv("REDIS_URL"),
		MovieServiceURL: getEnv("MOVIE_SERVICE_URL"),
	}
}

func getEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("%s is not set", key)
	}
	return value
}
