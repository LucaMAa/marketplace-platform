package config

import (
	"fmt"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"marketplace-platform/models"
)

type Config struct {
	DBHost     string
	DBUser     string
	DBPassword string
	DBName     string
	DBPort     string
	RedisAddr  string
	ESUrl      string
	JWTSecret  string
	Port       string
}

func Load() *Config {
	return &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBUser:     getEnv("DB_USER", "marketplace"),
		DBPassword: getEnv("DB_PASSWORD", "marketplace"),
		DBName:     getEnv("DB_NAME", "marketplace"),
		DBPort:     getEnv("DB_PORT", "5432"),
		RedisAddr:  getEnv("REDIS_ADDR", "localhost:6379"),
		ESUrl:      getEnv("ES_URL", "http://localhost:9200"),
		JWTSecret:  getEnv("JWT_SECRET", "change-me-in-production"),
		Port:       getEnv("PORT", "8080"),
	}
}

func InitDB(cfg *Config) *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// AutoMigrate all entities
	if err := db.AutoMigrate(
		&models.User{},
		&models.Merchant{},
		&models.Product{},
		&models.ProductCategory{},
		&models.Request{},
		&models.ChatMessage{},
		&models.Inventory{},
		&models.MerchantNotification{},
		&models.RequestMerchant{},
	); err != nil {
		log.Fatalf("automigrate failed: %v", err)
	}

	log.Println("Database migrated successfully")
	return db
}

func InitRedis(cfg *Config) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
	})
	return client
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
