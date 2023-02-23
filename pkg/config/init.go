package config

import (
	"log"
	"os"
	"time"
)

type AuthService struct {
	SecretSignature     string
	AccessTokenExpired  time.Duration //todo convert to expiration
	RefreshTokenExpired time.Duration
}

type Config struct {
	Loglevel    string
	ServerHost  string
	ServerPort  string
	Dbase       string
	Srv         AuthService
}

func getEnv(key, defaultKey string) string {
	if envVal, exists := os.LookupEnv(key); exists {
		return envVal
	}
	return defaultKey
}

func parseDurationString(key, defaultKey string) time.Duration {
	var dur time.Duration
	value := getEnv(key, defaultKey)
	dur, err := time.ParseDuration(value)
	if err != nil{
		log.Fatalf("Failed to parse DURATION", err.Error())
	}
	return dur
}

func New() *Config {
	auths := AuthService{
		SecretSignature: getEnv("SIGN_SUGNATURE", "mysignature"),
		AccessTokenExpired: parseDurationString("ACCESS_DURATION", "5m"),
		RefreshTokenExpired: parseDurationString("REFRESH_DURATION", "1h"),

	}
	return &Config{
		Loglevel: getEnv("LOG_LEVEL", "INFO"),
		ServerHost: getEnv("SERVER_HOST", "127.0.0.1"),
		ServerPort: getEnv("SERVER_PORT", ":8080"),
		Dbase:      getEnv("DATABSE", "postgres"),
		Srv: auths,
		
	}
}