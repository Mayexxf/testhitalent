package config

import (
	"fmt"
	"os"
)

type Config struct {
	DB        DB
	Migration Migration
}

type DB struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
	ServerPort string
}

type Migration struct {
	Dir string
}

func Load() *Config {
	return &Config{
		DB: DB{
			DBHost:     getEnv("DB_HOST", "localhost"),
			DBPort:     getEnv("DB_PORT", "5432"),
			DBUser:     getEnv("DB_USER", "postgres"),
			DBPassword: getEnv("DB_PASSWORD", "postgres"),
			DBName:     getEnv("DB_NAME", "orgstructure"),
			DBSSLMode:  getEnv("DB_SSLMODE", "disable"),
			ServerPort: getEnv("SERVER_PORT", "8080"),
		},
		Migration: Migration{
			Dir: "./migrations",
		},
	}
}

func (c *Config) GetDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DB.DBHost, c.DB.DBPort, c.DB.DBUser, c.DB.DBPassword, c.DB.DBName, c.DB.DBSSLMode)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
