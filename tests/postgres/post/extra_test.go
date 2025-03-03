// post/postgres_test.go

package post

import (
	"OZON/internal/domain"
	pg "OZON/internal/repository/postrges"
	"OZON/pkg/storage"
	"context"
	"fmt"
	"github.com/google/uuid"
	"testing"
	"time"
)

// TestGetCommentsForPostPagination содержит тесты для пагинации комментариев
func TestGetCommentsForPostPagination(t *testing.T) {
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

	t.Run("PaginationWithLimit5", testPaginationWithLimit5(repo))
	t.Run("PaginationWithMultiplePages", testPaginationWithMultiplePages(repo))
	t.Run("InvalidPage", testInvalidPage(repo))
	t.Run("InvalidLimit", testInvalidLimit(repo))
	t.Run("NoComments", testNoComments(repo))

}

// testPaginationWithLimit5 проверяет пагинацию с лимитом 5 комментариев
func testPaginationWithLimit5(repo *pg.PostgresRepository) func(*testing.T) {
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

		comments := make([]*domain.Comment, 8)
		for i := 0; i < 8; i++ {
			now := time.Now().Add(time.Duration(i) * time.Millisecond) // Разные времена для сортировки
			comments[i] = &domain.Comment{
				ID:        uuid.New(),
				PostID:    postID,
				Text:      fmt.Sprintf("Comment %d", i),
				CreatedAt: &now,
			}
			_, err := repo.CreateComment(context.Background(), comments[i])
			if err != nil {
				t.Fatalf("failed to create comment %d: %v", i, err)
			}
		}

		// Проверка первой страницы (page=1, limit=5)
		commentsList, err := repo.GetCommentsForPost(context.Background(), postID, 1, 5)
		if err != nil {
			t.Fatalf("failed to get comments: %v", err)
		}
		if len(commentsList) != 5 {
			t.Errorf("expected 5 comments, got %d", len(commentsList))
		}

		for i := 0; i < 4; i++ { // 5 комментариев, проверяем 4 пары
			if commentsList[i].CreatedAt == nil || commentsList[i+1].CreatedAt == nil {
				t.Errorf("CreatedAt should not be nil for comments")
				continue
			}
			if !commentsList[i].CreatedAt.Before(*commentsList[i+1].CreatedAt) {
				t.Errorf("comments not sorted correctly: %v should be before %v", commentsList[i].Text, commentsList[i+1].Text)
			}
		}

		// Проверка второй страницы (page=2, limit=5)
		commentsList, err = repo.GetCommentsForPost(context.Background(), postID, 2, 5)
		if err != nil {
			t.Fatalf("failed to get comments: %v", err)
		}
		if len(commentsList) != 3 { // 8 - 5 = 3 оставшихся комментария
			t.Errorf("expected 3 comments, got %d", len(commentsList))
		}
		// Проверяем порядок (от ранних к поздним по CreatedAt)
		for i := 0; i < 2; i++ { // 3 комментария, проверяем 2 пары
			if commentsList[i].CreatedAt == nil || commentsList[i+1].CreatedAt == nil {
				t.Errorf("CreatedAt should not be nil for comments")
				continue
			}
			if !commentsList[i].CreatedAt.Before(*commentsList[i+1].CreatedAt) {
				t.Errorf("comments not sorted correctly: %v should be before %v", commentsList[i].Text, commentsList[i+1].Text)
			}
		}
	}
}

// testPaginationWithMultiplePages проверяет пагинацию с несколькими страницами
func testPaginationWithMultiplePages(repo *pg.PostgresRepository) func(*testing.T) {
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

		// Создание 10 корневых комментариев
		comments := make([]*domain.Comment, 10)
		for i := 0; i < 10; i++ {
			now := time.Now().Add(time.Duration(i) * time.Millisecond)
			comments[i] = &domain.Comment{
				ID:        uuid.New(),
				PostID:    postID,
				Text:      fmt.Sprintf("Comment %d", i),
				CreatedAt: &now,
			}
			_, err := repo.CreateComment(context.Background(), comments[i])
			if err != nil {
				t.Fatalf("failed to create comment %d: %v", i, err)
			}
		}

		// Проверка страниц (page=1, limit=4; page=2, limit=4; page=3, limit=4)
		for p := 1; p <= 3; p++ {
			commentsList, err := repo.GetCommentsForPost(context.Background(), postID, int32(p), 4)
			if err != nil {
				t.Fatalf("failed to get comments for page %d: %v", p, err)
			}
			expectedLen := 4
			if p == 3 {
				expectedLen = 2
			}
			if len(commentsList) != expectedLen {
				t.Errorf("for page %d, expected %d comments, got %d", p, expectedLen, len(commentsList))
			}
			// Проверяем порядок (от ранних к поздним)
			for i := 0; i < len(commentsList)-1; i++ {
				if commentsList[i].CreatedAt == nil || commentsList[i+1].CreatedAt == nil {
					t.Errorf("CreatedAt should not be nil for comments")
					continue
				}
				if !commentsList[i].CreatedAt.Before(*commentsList[i+1].CreatedAt) {
					t.Errorf("comments not sorted correctly on page %d: %v should be before %v", p, commentsList[i].Text, commentsList[i+1].Text)
				}
			}
		}
	}
}

// testInvalidPage проверяет обработку некорректного page
func testInvalidPage(repo *pg.PostgresRepository) func(*testing.T) {
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

		// Проверка page <= 0
		_, err = repo.GetCommentsForPost(context.Background(), postID, 0, 5)
		if err == nil || err.Error() != "page must be greater than 0" {
			t.Errorf("expected error 'page must be greater than 0', got %v", err)
		}
	}
}

// testInvalidLimit проверяет обработку некорректного limit
func testInvalidLimit(repo *pg.PostgresRepository) func(*testing.T) {
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

		// Проверка limit < 0
		_, err = repo.GetCommentsForPost(context.Background(), postID, 1, -1)
		if err == nil || err.Error() != "limit must be greater than or equal to 0" {
			t.Errorf("expected error 'limit must be greater than or equal to 0', got %v", err)
		}
	}
}

// testNoComments проверяет поведение при отсутствии комментариев
func testNoComments(repo *pg.PostgresRepository) func(*testing.T) {
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

		// Проверка пагинации без комментариев
		comments, err := repo.GetCommentsForPost(context.Background(), postID, 1, 5)
		if err != nil {
			t.Fatalf("failed to get comments: %v", err)
		}
		if len(comments) != 0 {
			t.Errorf("expected 0 comments, got %d", len(comments))
		}
	}
}
