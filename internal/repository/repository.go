package repository

import (
	"OZON/internal/domain"
	"context"
	"github.com/google/uuid"
)

type PostRepository interface {
	GetPost(ctx context.Context, id uuid.UUID, commentPage, commentLimit int32) (*domain.Post, error)
	GetPosts(ctx context.Context, page, limit int32) ([]*domain.Post, error)
	CreatePost(ctx context.Context, post *domain.Post, allowComments *bool) (*domain.Post, error)
	IsCommentsAllowed(ctx context.Context, postID uuid.UUID) (bool, error)
}

type CommentRepository interface {
	CreateComment(ctx context.Context, comment *domain.Comment) (*domain.Comment, error)
	GetCommentsForPost(ctx context.Context, postID uuid.UUID, page, limit int32) ([]*domain.Comment, error)
}
