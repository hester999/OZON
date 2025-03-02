package memory

import (
	"OZON/internal/domain"
	"context"
	"fmt"
	"github.com/google/uuid"
	"sync"
	"time"
)

type InMemoryRepository struct {
	posts    sync.Map
	comments sync.Map
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{}
}

func (r *InMemoryRepository) GetPost(ctx context.Context, id uuid.UUID, commentPage, commentLimit int32) (*domain.Post, error) {
	if commentPage <= 0 {
		return nil, fmt.Errorf("comment page must be greater than 0")
	}
	if commentLimit <= 0 {
		return nil, fmt.Errorf("comment limit must be greater than or equal to 0")
	}

	if v, ok := r.posts.Load(id); ok {
		post, ok := v.(*domain.Post)
		if !ok {
			return nil, fmt.Errorf("invalid post type")
		}

		postCopy := *post
		postCopy.Comments = r.getCommentsForPost(post.ID, commentPage, commentLimit)
		return &postCopy, nil
	}
	return nil, fmt.Errorf("post not found")
}

func (r *InMemoryRepository) GetPosts(ctx context.Context, page, limit int32) ([]*domain.Post, error) {
	if page <= 0 {
		return nil, fmt.Errorf("page must be greater than 0")
	}
	if limit <= 0 {
		return nil, fmt.Errorf("limit must be greater than or equal to 0")
	}

	var allPosts []*domain.Post
	r.posts.Range(func(key, value interface{}) bool {
		if post, ok := value.(*domain.Post); ok {
			allPosts = append(allPosts, post)
		}
		return true
	})

	offset := (page - 1) * limit
	if offset < 0 {
		offset = 0
	}
	if offset > int32(len(allPosts)) {
		return nil, fmt.Errorf("page out of range")
	}
	end := offset + limit
	if end > int32(len(allPosts)) {
		end = int32(len(allPosts))
	}

	paginatedPosts := make([]*domain.Post, 0, limit)
	for _, post := range allPosts[offset:end] {
		postCopy := *post
		postCopy.Comments = r.getCommentsForPost(post.ID, offset, limit)
		paginatedPosts = append(paginatedPosts, &postCopy)
	}

	if len(paginatedPosts) == 0 {
		return nil, fmt.Errorf("no posts found")
	}
	return paginatedPosts, nil
}

func (r *InMemoryRepository) CreatePost(ctx context.Context, post *domain.Post, allowComments *bool) (*domain.Post, error) {
	if post.ID == uuid.Nil {
		post.ID = uuid.New()
	}

	allow := true
	if allowComments != nil {
		allow = *allowComments
	}
	post.AllowComments = allow
	post.Comments = nil

	r.posts.Store(post.ID, post)
	return post, nil
}

func (r *InMemoryRepository) IsCommentsAllowed(ctx context.Context, postID uuid.UUID) (bool, error) {

	if v, ok := r.posts.Load(postID); ok {
		allow, ok := v.(*domain.Post)
		if !ok {
			return false, fmt.Errorf("invalid post type")
		}
		return allow.AllowComments, nil
	}
	return false, fmt.Errorf("post not found")
}

func (r *InMemoryRepository) CreateComment(ctx context.Context, comment *domain.Comment) (*domain.Comment, error) {
	comment.ID = uuid.New()
	if comment.CreatedAt == nil {
		now := time.Now()
		comment.CreatedAt = &now
	}
	if val, ok := r.posts.Load(comment.PostID); ok {
		post, ok := val.(*domain.Post)
		if !ok {
			return nil, fmt.Errorf("invalid post type")
		}
		if !post.AllowComments {
			return nil, fmt.Errorf("comments not allowed for this post")
		}
		if comment.ParentID != nil {
			if parent, ok := r.comments.Load(*comment.ParentID); ok {
				if _, ok := parent.(*domain.Comment); ok {
				} else {
					return nil, fmt.Errorf("parent not found or invalid type")
				}
			} else {
				return nil, fmt.Errorf("parent not found")
			}
		}
		r.comments.Store(comment.ID, comment)
		return comment, nil
	}
	return nil, fmt.Errorf("post not found")
}

func (r *InMemoryRepository) GetCommentsForPost(ctx context.Context, postID uuid.UUID, page, limit int32) ([]*domain.Comment, error) {
	return r.getCommentsForPost(postID, page, limit), nil
}

func (r *InMemoryRepository) getCommentsForPost(postID uuid.UUID, page, limit int32) []*domain.Comment {
	var rootComments []*domain.Comment
	offset := (page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	var count int
	r.comments.Range(func(key, value interface{}) bool {
		if comment, ok := value.(*domain.Comment); ok {
			if comment.PostID == postID && comment.ParentID == nil {
				if count >= int(offset) && count < int(offset)+int(limit) {
					commentCopy := *comment
					commentCopy.Children = make([]*domain.Comment, 0)
					rootComments = append(rootComments, &commentCopy)
				}
				count++
			}
		}
		return true
	})

	for i, root := range rootComments {
		rootComments[i].Children = r.buildCommentTreeForChildren(root.ID)
	}

	return rootComments
}

func (r *InMemoryRepository) buildCommentTreeForChildren(parentID uuid.UUID) []*domain.Comment {
	var children []*domain.Comment

	r.comments.Range(func(key, value interface{}) bool {
		if comment, ok := value.(*domain.Comment); ok {
			if comment.ParentID != nil && comment.ParentID.String() == parentID.String() {
				commentCopy := *comment
				commentCopy.Children = r.buildCommentTreeForChildren(comment.ID)
				children = append(children, &commentCopy)
			}
		}
		return true
	})

	return children
}
