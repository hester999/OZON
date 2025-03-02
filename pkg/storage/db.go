package storage

import (
	"OZON/internal/domain"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

type DB struct {
	*gorm.DB
}

func NewPostgresDB() (*DB, error) {
	connString := os.Getenv("POSTGRES_CONNECTION_STRING")
	if connString == "" {
		connString = "postgres://admin:123@localhost:5432/ozon?sslmode=disable"
	}
	customLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // Логгер для вывода
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      logger.Info,
			Colorful:      true,
		},
	)
	db, err := gorm.Open(postgres.Open(connString), &gorm.Config{
		Logger: customLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(&domain.Post{}, &domain.Comment{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %v", err)
	}

	return &DB{db.Debug()}, nil
}
