package usecases

import (
	"OZON/internal/domain"
	"OZON/internal/repository"
	"context"
	"fmt"
	"github.com/google/uuid"
)

type CommentUsecase struct {
	postRepo    repository.PostRepository
	commentRepo repository.CommentRepository
}

func NewCommentUsecase(postRepo repository.PostRepository, commentRepo repository.CommentRepository) *CommentUsecase {
	return &CommentUsecase{
		postRepo:    postRepo,
		commentRepo: commentRepo,
	}
}

func (u *CommentUsecase) CreateComment(ctx context.Context, comment *domain.Comment) (*domain.Comment, error) {
	if comment.PostID == uuid.Nil {
		return nil, fmt.Errorf("post ID cannot be empty")
	}
	if comment.Text == "" {
		return nil, fmt.Errorf("comment text cannot be empty")
	}
	if len(comment.Text) > 2000 {
		return nil, fmt.Errorf("comment text exceeds 2000 characters")
	}
	if comment.ParentID != nil {
		if _, err := uuid.Parse(comment.ParentID.String()); err != nil {
			return nil, fmt.Errorf("invalid parent ID format: %v", err)
		}
		_, err := u.commentRepo.GetCommentsForPost(ctx, comment.PostID, 1, 1)
		if err != nil {
			return nil, fmt.Errorf("parent comment not found or invalid: %v", err)
		}
	}

	allowed, err := u.postRepo.IsCommentsAllowed(ctx, comment.PostID)
	if err != nil || !allowed {
		return nil, fmt.Errorf("comments not allowed or post not found: %v", err)
	}

	return u.commentRepo.CreateComment(ctx, comment)
}

func (u *CommentUsecase) GetCommentsForPost(ctx context.Context, postID string, page, limit int32) ([]*domain.Comment, error) {
	if postID == "" {
		return nil, fmt.Errorf("post ID cannot be empty")
	}
	uuidPostID, err := uuid.Parse(postID)
	if err != nil {
		return nil, fmt.Errorf("invalid post ID format: %v", err)
	}
	if page <= 0 {
		return nil, fmt.Errorf("page must be greater than 0")
	}
	if limit < 0 {
		return nil, fmt.Errorf("limit must be greater than or equal to 0")
	}

	comments, err := u.commentRepo.GetCommentsForPost(ctx, uuidPostID, page, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments: %v", err)
	}
	return comments, nil
}
