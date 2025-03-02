package postgres

import (
	"OZON/internal/domain"
	"OZON/pkg/storage"
	"context"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PostgresRepository struct {
	db storage.DB
}

func NewPostgresRepository(db storage.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (p *PostgresRepository) GetPost(ctx context.Context, id uuid.UUID, commentPage, commentLimit int32) (*domain.Post, error) {
	var post domain.Post
	offset := (commentPage - 1) * commentLimit

	if commentPage <= 0 {
		return nil, fmt.Errorf("comment page must be greater than 0")
	}
	if commentLimit <= 0 {
		return nil, fmt.Errorf("comment limit must be greater than or equal to 0")
	}

	if err := p.db.WithContext(ctx).Where("id = ?", id).Preload("Comments", func(db *gorm.DB) *gorm.DB {
		return db.Where("post_id = ?", id).Order("created_at").Limit(int(commentLimit)).Offset(int(offset))
	}).First(&post).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("post not found")
		}
		return nil, fmt.Errorf("failed to get post: %v", err)
	}

	post.Comments = buildCommentTree(post.Comments, id)
	return &post, nil
}

func (p *PostgresRepository) CreatePost(ctx context.Context, post *domain.Post, allowComments *bool) (*domain.Post, error) {
	if post.ID == uuid.Nil {
		post.ID = uuid.New()
	}

	allow := true
	if allowComments != nil {
		allow = *allowComments
	}
	post.AllowComments = allow

	if err := p.db.WithContext(ctx).Create(post).Error; err != nil {
		return nil, fmt.Errorf("failed to create post: %v", err)
	}

	return post, nil
}

func (p *PostgresRepository) GetPosts(ctx context.Context, page, limit int32) ([]*domain.Post, error) {
	var posts []*domain.Post
	offset := (page - 1) * limit

	var postIDs []uuid.UUID
	if err := p.db.WithContext(ctx).Model(&domain.Post{}).Limit(int(limit)).Offset(int(offset)).Pluck("id", &postIDs).Error; err != nil {
		return nil, fmt.Errorf("failed to get post IDs: %v", err)
	}

	if err := p.db.WithContext(ctx).Limit(int(limit)).Offset(int(offset)).Preload("Comments", func(db *gorm.DB) *gorm.DB {
		return db.Where("post_id IN (?)", postIDs).Order("created_at")
	}).Find(&posts).Error; err != nil {
		return nil, fmt.Errorf("failed to get posts: %v", err)
	}

	for i := range posts {
		posts[i].Comments = buildCommentTree(posts[i].Comments, posts[i].ID)
	}

	return posts, nil
}
func (p *PostgresRepository) IsCommentsAllowed(ctx context.Context, postID uuid.UUID) (bool, error) {
	var post domain.Post
	if err := p.db.WithContext(ctx).Select("allow_comments").Where("id = ?", postID).First(&post).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, fmt.Errorf("post not found")
		}
		return false, fmt.Errorf("failed to check comments allowed: %v", err)
	}
	return post.AllowComments, nil
}

func (p *PostgresRepository) CreateComment(ctx context.Context, comment *domain.Comment) (*domain.Comment, error) {
	comment.ID = uuid.New()

	if err := p.db.WithContext(ctx).Create(comment).Error; err != nil {
		return nil, fmt.Errorf("failed to create comment: %v", err)
	}

	return comment, nil
}

func (p *PostgresRepository) GetCommentsForPost(ctx context.Context, postID uuid.UUID, page, limit int32) ([]*domain.Comment, error) {
	var comments []*domain.Comment
	offset := (page - 1) * limit

	var exists bool
	if err := p.db.WithContext(ctx).Model(&domain.Post{}).Select("count(*) > 0").Where("id = ? ", postID).Find(&exists).Error; err != nil {
		return nil, fmt.Errorf("failed to check post existence: %v", err)
	}
	if !exists {
		return nil, fmt.Errorf("post not found")
	}

	if page <= 0 {
		return nil, fmt.Errorf("page must be greater than 0")
	}
	if limit < 0 {
		return nil, fmt.Errorf("limit must be greater than or equal to 0")
	}

	if err := p.db.WithContext(ctx).Where("post_id = ?", postID).Order("created_at").Limit(int(limit)).Offset(int(offset)).Find(&comments).Error; err != nil {
		return nil, fmt.Errorf("failed to get comments: %v", err)
	}

	return buildCommentTree(comments, postID), nil
}

func buildCommentTree(comments []*domain.Comment, postID uuid.UUID) []*domain.Comment {
	var rootComments []*domain.Comment
	mapComments := make(map[string]*domain.Comment)

	for _, value := range comments {
		mapComments[value.ID.String()] = value
		value.Children = make([]*domain.Comment, 0)
		if value.ParentID == nil && value.PostID == postID {
			rootComments = append(rootComments, value)
		}
	}

	for _, comment := range comments {
		if comment.ParentID != nil {
			parentID := comment.ParentID.String()
			parent, ok := mapComments[parentID]
			if ok {
				parent.Children = append(parent.Children, comment)
			}
		}
	}
	return rootComments
}
