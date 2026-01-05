package config

import (
	"fmt"
	"log"

	"starter-kit-grpc-golang/internal/models"
	"starter-kit-grpc-golang/pkg/logger"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DB Global Database instance
var DB *gorm.DB

// ConnectDB initializes the database
func ConnectDB(cfg *Config) {
	var err error
	var dsn string
	var dialector gorm.Dialector

	if cfg.Database.Driver == "sqlite" {
		// SQLite setup (File based)
		dsn = cfg.Database.Name + ".db"
		dialector = sqlite.Open(dsn)
		logger.Log.Info("Using SQLite database", "file", dsn)
	} else {
		// PostgreSQL setup
		dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
			cfg.Database.Host,
			cfg.Database.User,
			cfg.Database.Password,
			cfg.Database.Name,
			cfg.Database.Port,
			cfg.Database.SSLMode,
		)
		dialector = postgres.Open(dsn)
		logger.Log.Info("Using PostgreSQL database", "host", cfg.Database.Host)
	}

	DB, err = gorm.Open(dialector, &gorm.Config{
		// Logger: logger.Default.LogMode(logger.Info), // Uncomment to see raw SQL
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto Migrate (Create Tables)
	err = DB.AutoMigrate(&models.User{}, &models.Token{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	logger.Log.Info("Database connected and migrated successfully")
}