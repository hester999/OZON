package config

import (
	"OZON/internal/cli"
	"OZON/internal/repository"
	"OZON/internal/repository/memory"
	postgres "OZON/internal/repository/postrges"

	"OZON/pkg/storage"
	"fmt"
)

type Config struct {
	PostRepository    repository.PostRepository
	CommentRepository repository.CommentRepository
}

func NewConfig(db storage.DB) (*Config, error) {

	repo := cli.ProcessFlag(db)

	switch r := repo.(type) {
	case *memory.InMemoryRepository:
		return &Config{
			PostRepository:    r,
			CommentRepository: r,
		}, nil
	case *postgres.PostgresRepository:
		return &Config{
			PostRepository:    r,
			CommentRepository: r,
		}, nil
	default:
		return nil, fmt.Errorf("invalid repository type returned from cli.ProcessFlag: %T", r)
	}
}
