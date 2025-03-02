package handlers

// THIS CODE WILL BE UPDATED WITH SCHEMA CHANGES. PREVIOUS IMPLEMENTATION FOR SCHEMA CHANGES WILL BE KEPT IN THE COMMENT SECTION. IMPLEMENTATION FOR UNCHANGED SCHEMA WILL BE KEPT.

import (
	"OZON/graph"
	"OZON/graph/model"
	"OZON/internal/domain"
	"OZON/internal/usecases"
	"context"
	"fmt"
	"github.com/google/uuid"
	"time"
)

type Resolver struct {
	postUsecase    *usecases.PostUsecase
	commentUsecase *usecases.CommentUsecase
}

func NewResolver(postUsecase *usecases.PostUsecase, commentUsecase *usecases.CommentUsecase) *Resolver {
	return &Resolver{
		postUsecase:    postUsecase,
		commentUsecase: commentUsecase,
	}
}

// CreatePost is the resolver for the createPost field.
func (r *mutationResolver) CreatePost(ctx context.Context, text string, allowComments *bool) (*model.Post, error) {
	if text == "" {
		return nil, fmt.Errorf("post text cannot be empty")
	}
	if len(text) > 10000 {
		return nil, fmt.Errorf("post text too long")
	}

	domainPost, err := r.postUsecase.CreatePost(ctx, text, allowComments)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &model.Post{
		ID:            domainPost.ID.String(),
		Text:          domainPost.Text,
		AllowComments: domainPost.AllowComments,
		Comments:      convertComments(domainPost.Comments, nil, nil),
	}, nil
}

// CreateComment is the resolver for the createComment field.
func (r *mutationResolver) CreateComment(ctx context.Context, postID string, text string, parentID *string) (*model.Comment, error) {
	if postID == "" {
		return nil, fmt.Errorf("post ID cannot be empty")
	}
	uuidPostID, err := uuid.Parse(postID)
	if err != nil {
		return nil, fmt.Errorf("invalid post ID format: %v", err)
	}
	if text == "" {
		return nil, fmt.Errorf("comment text cannot be empty")
	}
	if len(text) > 2000 {
		return nil, fmt.Errorf("comment text exceeds 2000 characters")
	}
	var uuidParentID *uuid.UUID
	if parentID != nil {
		pid, err := uuid.Parse(*parentID)
		if err != nil {
			return nil, fmt.Errorf("invalid parent ID format: %v", err)
		}
		uuidParentID = &pid
	}

	domainComment := &domain.Comment{
		PostID:   uuidPostID,
		Text:     text,
		ParentID: uuidParentID,
	}
	createdComment, err := r.commentUsecase.CreateComment(ctx, domainComment)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	var parentIDStr *string
	if createdComment.ParentID != nil {
		str := createdComment.ParentID.String()
		parentIDStr = &str
	}
	return &model.Comment{
		ID:        createdComment.ID.String(),
		Text:      createdComment.Text,
		PostID:    createdComment.PostID.String(),
		ParentID:  parentIDStr,
		CreatedAt: createdComment.CreatedAt.Format(time.RFC3339),
		Children:  convertComments(createdComment.Children, nil, nil),
	}, nil
}

// GetPost is the resolver for the getPost field.
func (r *queryResolver) GetPost(ctx context.Context, id string, commentPage *int, commentLimit *int) (*model.Post, error) {
	if id == "" {
		return nil, fmt.Errorf("post ID cannot be empty")
	}
	uuidID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid post ID format: %v", err)
	}

	limit := int32(10)
	page := int32(1)
	if commentLimit != nil {
		limit = int32(*commentLimit)
	}
	if commentPage != nil {
		page = int32(*commentPage)
	}
	if page <= 0 {
		return nil, fmt.Errorf("comment page must be greater than 0")
	}
	if limit <= 0 {
		return nil, fmt.Errorf("comment limit must be greater than 0")
	}

	domainPost, err := r.postUsecase.GetPost(ctx, uuidID.String(), page, limit)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &model.Post{
		ID:            domainPost.ID.String(),
		Text:          domainPost.Text,
		AllowComments: domainPost.AllowComments,
		Comments:      convertComments(domainPost.Comments, nil, nil),
	}, nil
}

// GetPosts is the resolver for the getPosts field.
func (r *queryResolver) GetPosts(ctx context.Context, page *int, limit *int) ([]*model.Post, error) {
	lim := int32(10)
	pa := int32(1)
	if limit != nil {
		lim = int32(*limit)
	}
	if page != nil {
		pa = int32(*page)
	}
	if pa <= 0 {
		return nil, fmt.Errorf("page must be greater than 0")
	}
	if lim <= 0 {
		return nil, fmt.Errorf("limit must be greater than 0")
	}

	domainPosts, err := r.postUsecase.GetPosts(ctx, pa, lim)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	var posts []*model.Post
	for _, dp := range domainPosts {
		posts = append(posts, &model.Post{
			ID:            dp.ID.String(),
			Text:          dp.Text,
			AllowComments: dp.AllowComments,
			Comments:      convertComments(dp.Comments, nil, nil),
		})
	}
	return posts, nil
}

func convertComments(domainComments []*domain.Comment, page *int, limit *int) []*model.Comment {
	var comments []*model.Comment

	start := 0
	end := len(domainComments)
	if page != nil && limit != nil {
		start = (*page - 1) * (*limit)
		end = start + *limit
		if start < 0 {
			start = 0
		}
		if end > len(domainComments) {
			end = len(domainComments)
		}
	}

	for _, v := range domainComments[start:end] {
		var parentID *string
		if v.ParentID != nil {
			parentIDStr := v.ParentID.String()
			parentID = &parentIDStr
		}
		comments = append(comments, &model.Comment{
			ID:        v.ID.String(),
			Text:      v.Text,
			PostID:    v.PostID.String(),
			ParentID:  parentID,
			CreatedAt: v.CreatedAt.Format(time.RFC3339),
			Children:  convertComments(v.Children, page, limit),
		})
	}
	return comments
}

// NewComment is the resolver for the newComment field.
func (r *subscriptionResolver) NewComment(ctx context.Context, postID string) (<-chan *model.Comment, error) {

	return nil, nil
}

func convertComment(domainComment *domain.Comment, page *int, limit *int) *model.Comment {
	var parentID *string
	if domainComment.ParentID != nil {
		parentIDStr := domainComment.ParentID.String()
		parentID = &parentIDStr
	}
	return &model.Comment{
		ID:        domainComment.ID.String(),
		Text:      domainComment.Text,
		PostID:    domainComment.PostID.String(),
		ParentID:  parentID,
		CreatedAt: domainComment.CreatedAt.Format(time.RFC3339),
		Children:  convertComments(domainComment.Children, page, limit),
	}
}

// Mutation returns graph.MutationResolver implementation.
func (r *Resolver) Mutation() graph.MutationResolver { return &mutationResolver{r} }

// Query returns graph.QueryResolver implementation.
func (r *Resolver) Query() graph.QueryResolver { return &queryResolver{r} }

// Subscription returns graph.SubscriptionResolver implementation.
func (r *Resolver) Subscription() graph.SubscriptionResolver { return &subscriptionResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type subscriptionResolver struct{ *Resolver }
