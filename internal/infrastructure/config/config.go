package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DBUrl          string
	JWTSecret      string
	JWTExpiration  time.Duration
	Port           string
	ServerClientId string
	RedisAddress   string
	RedisPassword  string
	RedisDB        int
	SMTPHost       string
	SMTPPort       int
	SMTPEmail      string
	SMTPPassword   string
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
	redisDB := 0
	if raw := os.Getenv("REDIS_DB"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			redisDB = parsed
		}
	}
	redisAddress := os.Getenv("REDIS_ADDR")
	if redisAddress == "" {
		redisAddress = "localhost:6379"
	}
	smtpPort := 0
	if raw := os.Getenv("SMTP_PORT"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			smtpPort = parsed
		}
	}
	config := &Config{
		DBUrl:          buildDburl(),
		JWTSecret:      os.Getenv("JWT_SECRET"),
		JWTExpiration:  ttl,
		Port:           os.Getenv("PORT"),
		ServerClientId: os.Getenv("SERVER_CLIENT_ID"),
		RedisAddress:   redisAddress,
		RedisPassword:  os.Getenv("REDIS_PASSWORD"),
		RedisDB:        redisDB,
		SMTPHost:       os.Getenv("SMTP_HOST"),
		SMTPPort:       smtpPort,
		SMTPEmail:      os.Getenv("SMTP_EMAIL"),
		SMTPPassword:   os.Getenv("SMTP_PASSWORD"),
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
