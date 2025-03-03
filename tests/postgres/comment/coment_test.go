package comment

import (
	"OZON/internal/domain"
	pg "OZON/internal/repository/postrges"
	"OZON/pkg/storage"
	"context"
	"github.com/google/uuid"
	"testing"
)

// TestCommentRepository содержит тесты для CommentRepository
func TestCommentRepository(t *testing.T) {
	db, err := storage.NewPostgresDB()
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	repo := pg.NewPostgresRepository(*db)

	if err := db.Exec("TRUNCATE TABLE posts, comments RESTART IDENTITY CASCADE").Error; err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}

	defer func() {
		if err := db.Exec("TRUNCATE TABLE posts, comments RESTART IDENTITY CASCADE").Error; err != nil {
			t.Fatalf("failed to truncate tables in defer: %v", err)
		}
	}()

	t.Run("CreateAndGetComment", testCreateAndGetComment(repo))
}

// testCreateAndGetComment проверяет создание и получение комментария
func testCreateAndGetComment(repo *pg.PostgresRepository) func(*testing.T) {
	return func(t *testing.T) {

		postID := uuid.New()
		post := &domain.Post{
			ID:            postID,
			Text:          "Test Post",
			AllowComments: true,
		}
		_, err := repo.CreatePost(context.Background(), post, nil)
		if err != nil {
			t.Fatalf("failed to create post: %v", err)
		}

		comment := &domain.Comment{
			PostID:   postID,
			Text:     "Test Comment",
			ParentID: nil, // Корневой комментарий
		}
		createdComment, err := repo.CreateComment(context.Background(), comment)
		if err != nil {
			t.Fatalf("failed to create comment: %v", err)
		}

		// Получение поста
		retrievedPost, err := repo.GetPost(context.Background(), postID, 1, 10)
		if err != nil {
			t.Fatalf("failed to get post: %v", err)
		}

		if len(retrievedPost.Comments) != 1 {
			t.Errorf("expected 1 comment, got %d", len(retrievedPost.Comments))
		}
		if retrievedPost.Comments[0].ID != createdComment.ID || retrievedPost.Comments[0].Text != "Test Comment" {
			t.Errorf("unexpected comment: %+v, want %+v", retrievedPost.Comments[0], createdComment)
		}

		// Проверка через GetCommentsForPost
		comments, err := repo.GetCommentsForPost(context.Background(), postID, 1, 10)
		if err != nil {
			t.Fatalf("failed to get comments: %v", err)
		}
		if len(comments) != 1 {
			t.Errorf("expected 1 comment, got %d", len(comments))
		}
		if comments[0].ID != createdComment.ID || comments[0].Text != "Test Comment" {
			t.Errorf("unexpected comment: %+v, want %+v", comments[0], createdComment)
		}
	}
}
