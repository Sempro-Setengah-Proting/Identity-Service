package config

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DBUrl          string
	JWTSecret      string
	JWTExpiration  time.Duration
	Port           string
	ServerClientId string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	ttl := 30 * time.Minute
	if raw := os.Getenv("JWT_EXPIRATION"); raw != "" {
		if parsed, err := time.ParseDuration(raw); err == nil {
			ttl = parsed
		}
	}
	config := &Config{
		DBUrl:          buildDburl(),
		JWTSecret:      os.Getenv("JWT_SECRET"),
		JWTExpiration:  ttl,
		Port:           os.Getenv("PORT"),
		ServerClientId: os.Getenv("SERVER_CLIENT_ID"),
	}

	return config
}

func buildDburl() string {

	return "host=" + os.Getenv("DB_HOST") +
		" user=" + os.Getenv("DB_USER") +
		" password=" + os.Getenv("DB_PASSWORD") +
		" dbname=" + os.Getenv("DB_NAME") +
		" port=" + os.Getenv("DB_PORT")
}
