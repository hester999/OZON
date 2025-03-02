package usecases

import (
	"OZON/internal/domain"
	"OZON/internal/repository"
	"context"
	"fmt"
	"github.com/google/uuid"
)

type PostUsecase struct {
	postRepo    repository.PostRepository
	commentRepo repository.CommentRepository
}

func NewPostUsecase(postRepo repository.PostRepository, commentRepo repository.CommentRepository) *PostUsecase {
	return &PostUsecase{
		postRepo:    postRepo,
		commentRepo: commentRepo,
	}
}

func (u *PostUsecase) GetPost(ctx context.Context, id string, commentPage, commentLimit int32) (*domain.Post, error) {
	if id == "" {
		return nil, fmt.Errorf("post ID cannot be empty")
	}
	uuidID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid post ID format: %v", err)
	}
	if commentPage <= 0 {
		return nil, fmt.Errorf("comment page must be greater than 0")
	}
	if commentLimit <= 0 {
		return nil, fmt.Errorf("comment limit must be greater than or equal to 0")
	}

	post, err := u.postRepo.GetPost(ctx, uuidID, commentPage, commentLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to get post: %v", err)
	}

	post.Comments, err = u.commentRepo.GetCommentsForPost(ctx, post.ID, commentPage, commentLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments for post: %v", err)
	}

	return post, nil
}

func (u *PostUsecase) GetPosts(ctx context.Context, page, limit int32) ([]*domain.Post, error) {
	if page <= 0 {
		return nil, fmt.Errorf("page must be greater than 0")
	}
	if limit <= 0 {
		return nil, fmt.Errorf("limit must be greater than or equal to 0")
	}

	posts, err := u.postRepo.GetPosts(ctx, page, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get posts: %v", err)
	}

	for i := range posts {
		if posts[i] != nil {
			posts[i].Comments, err = u.commentRepo.GetCommentsForPost(ctx, posts[i].ID, 1, 10)
			if err != nil {
				return nil, fmt.Errorf("failed to get comments for post %s: %v", posts[i].ID.String(), err)
			}
		}
	}

	return posts, nil
}

func (u *PostUsecase) CreatePost(ctx context.Context, text string, allowComments *bool) (*domain.Post, error) {
	if text == "" {
		return nil, fmt.Errorf("post text cannot be empty")
	}
	if len(text) > 10000 {
		return nil, fmt.Errorf("post text too long")
	}

	allow := true
	if allowComments != nil {
		allow = *allowComments
	}

	post := &domain.Post{
		Text:          text,
		AllowComments: allow,
	}
	createdPost, err := u.postRepo.CreatePost(ctx, post, &allow)
	if err != nil {
		return nil, fmt.Errorf("failed to create post: %v", err)
	}
	return createdPost, nil
}

func (u *PostUsecase) IsCommentsAllowed(ctx context.Context, postID string) (bool, error) {
	if postID == "" {
		return false, fmt.Errorf("post ID cannot be empty")
	}
	uuidPostID, err := uuid.Parse(postID)
	if err != nil {
		return false, fmt.Errorf("invalid post ID format: %v", err)
	}

	allowed, err := u.postRepo.IsCommentsAllowed(ctx, uuidPostID)
	if err != nil {
		return false, fmt.Errorf("failed to check comments allowed: %v", err)
	}
	return allowed, nil
}
