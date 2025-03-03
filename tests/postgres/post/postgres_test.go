package post

import (
	"OZON/internal/domain"
	pg "OZON/internal/repository/postrges"
	"OZON/pkg/storage"
	"context"
	_ "database/sql"
	"fmt"
	"github.com/google/uuid"
	_ "gorm.io/driver/postgres"
	"gorm.io/gorm"
	_ "gorm.io/gorm"
	"testing"
)

func TestPostgresRepository(t *testing.T) {
	db, err := storage.NewPostgresDB()
	if err != nil {
		fmt.Println(err)
		return
	}
	PostgresRepo := pg.NewPostgresRepository(*db)

	if err := db.Exec("TRUNCATE TABLE posts, comments RESTART IDENTITY CASCADE").Error; err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}

	defer func() {
		if err := db.Exec("TRUNCATE TABLE posts, comments RESTART IDENTITY CASCADE").Error; err != nil {
			t.Fatalf("failed to truncate tables in defer: %v", err)
		}
	}()

	t.Run("CreateAndGetPost", testCreateAndGetPost(PostgresRepo))
	t.Run("GetPostsWithPagination", testGetPostsWithPagination(PostgresRepo, db.DB))
	t.Run("CreateAndGetComment", testCreateAndGetComment(PostgresRepo))
	t.Run("IsCommentsAllowed", testIsCommentsAllowed(PostgresRepo))
	t.Run("GetCommentsForPostWithHierarchy", testGetCommentsForPostWithHierarchy(PostgresRepo)) // Исправлено имя теста
}
func testCreateAndGetPost(repo *pg.PostgresRepository) func(*testing.T) {
	return func(t *testing.T) {

		post := &domain.Post{
			ID:            uuid.New(),
			Text:          "Test Post",
			AllowComments: true,
		}
		createdPost, err := repo.CreatePost(context.Background(), post, nil)
		if err != nil {
			t.Fatalf("failed to create post: %v", err)
		}

		retrievedPost, err := repo.GetPost(context.Background(), post.ID, 1, 10)
		if err != nil {
			t.Fatalf("failed to get post: %v", err)
		}

		if retrievedPost.ID != createdPost.ID || retrievedPost.Text != "Test Post" || retrievedPost.AllowComments != true {
			t.Errorf("unexpected post data: got %+v, want %+v", retrievedPost, createdPost)
		}
		if len(retrievedPost.Comments) != 0 {
			t.Errorf("expected no comments, got %d", len(retrievedPost.Comments))
		}
	}
}

func testGetPostsWithPagination(repo *pg.PostgresRepository, db *gorm.DB) func(*testing.T) {
	return func(t *testing.T) {
		if err := db.Exec("TRUNCATE TABLE posts, comments RESTART IDENTITY CASCADE").Error; err != nil {
			t.Fatalf("failed to truncate tables: %v", err)
		}
		posts := make([]*domain.Post, 3)
		for i := 0; i < 3; i++ {
			posts[i] = &domain.Post{
				ID:            uuid.New(),
				Text:          fmt.Sprintf("Test Post %d", i),
				AllowComments: true,
			}

			_, err := repo.CreatePost(context.Background(), posts[i], nil)
			if err != nil {
				t.Fatalf("failed to create post %d: %v", i, err)
			}

		}

		retrievedPosts, err := repo.GetPosts(context.Background(), 1, 2)
		if err != nil {
			t.Fatalf("failed to get posts: %v", err)
		}

		if len(retrievedPosts) != 2 {
			t.Errorf("expected 2 posts, got %d", len(retrievedPosts))
		}

		if len(retrievedPosts) > 0 {
			if retrievedPosts[0].Text != "Test Post 0" || retrievedPosts[1].Text != "Test Post 1" {
				t.Errorf("unexpected post order: %+v", retrievedPosts)
			}
		}

		retrievedPosts, err = repo.GetPosts(context.Background(), 2, 2)
		if err != nil {
			t.Fatalf("failed to get posts: %v", err)
		}
		if len(retrievedPosts) != 1 {
			t.Errorf("expected 1 post, got %d", len(retrievedPosts))
		}
		if len(retrievedPosts) > 0 {

			if retrievedPosts[0].Text != "Test Post 2" {
				t.Errorf("unexpected post: %+v", retrievedPosts[0])
			}
		}
	}
}

// testCreateAndGetComment проверяет создание и получение комментария
func testCreateAndGetComment(repo *pg.PostgresRepository) func(*testing.T) {
	return func(t *testing.T) {

		post := &domain.Post{
			ID:            uuid.New(),
			Text:          "Test Post",
			AllowComments: true,
		}
		_, err := repo.CreatePost(context.Background(), post, nil)
		if err != nil {
			t.Fatalf("failed to create post: %v", err)
		}

		comment := &domain.Comment{
			PostID: post.ID,
			Text:   "Test Comment",
		}
		createdComment, err := repo.CreateComment(context.Background(), comment)
		if err != nil {
			t.Fatalf("failed to create comment: %v", err)
		}

		retrievedPost, err := repo.GetPost(context.Background(), post.ID, 1, 10)
		if err != nil {
			t.Fatalf("failed to get post: %v", err)
		}

		if len(retrievedPost.Comments) != 1 {
			t.Errorf("expected 1 comment, got %d", len(retrievedPost.Comments))
		}
		if retrievedPost.Comments[0].ID != createdComment.ID || retrievedPost.Comments[0].Text != "Test Comment" {
			t.Errorf("unexpected comment: %+v", retrievedPost.Comments[0])
		}
	}
}

// testIsCommentsAllowed проверяет проверку разрешения комментариев
func testIsCommentsAllowed(repo *pg.PostgresRepository) func(*testing.T) {
	return func(t *testing.T) {

		post := &domain.Post{
			ID:   uuid.New(),
			Text: "Test Post",
		}
		post, err := repo.CreatePost(context.Background(), post, nil)
		if err != nil {
			t.Fatalf("failed to create post: %v", err)
		}

		allowed, err := repo.IsCommentsAllowed(context.Background(), post.ID)
		if err != nil {
			t.Fatalf("failed to check comments allowed: %v", err)
		}
		if !allowed {
			t.Errorf("expected comments to be allowed, got false")
		}

		noCommentsPost := &domain.Post{
			ID:   uuid.New(),
			Text: "No Comments Post",
		}
		off := false
		pointerOff := &off
		noCommentsPost, err = repo.CreatePost(context.Background(), noCommentsPost, pointerOff)
		if err != nil {
			t.Fatalf("failed to create post: %v", err)
		}

		notAllowed, err := repo.IsCommentsAllowed(context.Background(), noCommentsPost.ID)
		if err != nil {
			t.Fatalf("failed to check comments allowed: %v", err)
		}
		if !notAllowed {

			t.Errorf("expected comments to be not allowed, got true")
		}
	}
}

// testGetCommentsForPostWithHierarchy проверяет получение комментариев с иерархией
func testGetCommentsForPostWithHierarchy(repo *pg.PostgresRepository) func(*testing.T) {
	return func(t *testing.T) {

		post := &domain.Post{
			ID:            uuid.New(),
			Text:          "Test Post",
			AllowComments: true,
		}
		_, err := repo.CreatePost(context.Background(), post, nil)
		if err != nil {
			t.Fatalf("failed to create post: %v", err)
		}

		rootComment := &domain.Comment{
			PostID: post.ID,
			Text:   "Root Comment",
		}
		_, err = repo.CreateComment(context.Background(), rootComment)
		if err != nil {
			t.Fatalf("failed to create root comment: %v", err)
		}

		nestedComment := &domain.Comment{
			PostID:   post.ID,
			Text:     "Nested Comment",
			ParentID: &rootComment.ID,
		}
		_, err = repo.CreateComment(context.Background(), nestedComment)
		if err != nil {
			t.Fatalf("failed to create nested comment: %v", err)
		}

		comments, err := repo.GetCommentsForPost(context.Background(), post.ID, 1, 10)
		if err != nil {
			t.Fatalf("failed to get comments: %v", err)
		}

		if len(comments) != 1 {
			t.Errorf("expected 1 root comment, got %d", len(comments))
		}
		if comments[0].Text != "Root Comment" || len(comments[0].Children) != 1 {
			t.Errorf("unexpected comment hierarchy: %+v", comments[0])
		}
		if comments[0].Children[0].Text != "Nested Comment" {
			t.Errorf("unexpected nested comment: %+v", comments[0].Children[0])
		}
	}
}
