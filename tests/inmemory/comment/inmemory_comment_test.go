package comment

import (
	"OZON/internal/domain"
	"OZON/internal/repository/memory"
	"context"
	"fmt"
	"github.com/google/uuid"
	"testing"
	"time"
)

// TestGetCommentsForPostPagination содержит тесты для пагинации комментариев
func TestGetCommentsForPostPagination(t *testing.T) {
	t.Run("PaginationWithLimit10", testPaginationWithLimit10)
	t.Run("PaginationWithMultiplePages", testPaginationWithMultiplePages)
	t.Run("InvalidPage", testInvalidPage)
	t.Run("InvalidLimit", testInvalidLimit)
	t.Run("NoComments", testNoComments)

}

// testPaginationWithLimit10 проверяет пагинацию с лимитом 10 комментариев
func testPaginationWithLimit10(t *testing.T) {
	repo := memory.NewInMemoryRepository()

	// Создание тестового поста
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

	// Создание 15 корневых комментариев с разными временами для сортировки
	comments := make([]*domain.Comment, 15)
	for i := 0; i < 15; i++ {
		// Создаём время и берём его адрес
		now := time.Now().Add(time.Duration(i) * time.Second)
		comments[i] = &domain.Comment{
			ID:        uuid.New(),
			PostID:    postID,
			Text:      fmt.Sprintf("Comment %d", i),
			CreatedAt: &now, // Берем адрес переменной now
		}
		_, err := repo.CreateComment(context.Background(), comments[i])
		if err != nil {
			t.Fatalf("failed to create comment %d: %v", i, err)
		}
	}

	// Проверка первой страницы (page=1, limit=10)
	commentsList, err := repo.GetCommentsForPost(context.Background(), postID, 1, 10)
	if err != nil {
		t.Fatalf("failed to get comments: %v", err)
	}
	if len(commentsList) != 10 {
		t.Errorf("expected 10 comments, got %d", len(commentsList))
	}
	// Проверяем порядок (от ранних к поздним по CreatedAt)
	for i := 0; i < 9; i++ {
		if commentsList[i].CreatedAt == nil || commentsList[i+1].CreatedAt == nil {
			t.Errorf("CreatedAt should not be nil for comments")
			continue
		}
		if !commentsList[i].CreatedAt.Before(*commentsList[i+1].CreatedAt) {
			t.Errorf("comments not sorted correctly: %v should be before %v", commentsList[i].Text, commentsList[i+1].Text)
		}
	}

	// Проверка второй страницы (page=2, limit=10)
	commentsList, err = repo.GetCommentsForPost(context.Background(), postID, 2, 10)
	if err != nil {
		t.Fatalf("failed to get comments: %v", err)
	}
	if len(commentsList) != 5 { // 15 - 10 = 5 оставшихся комментариев
		t.Errorf("expected 5 comments, got %d", len(commentsList))
	}
	// Проверяем порядок (от ранних к поздним по CreatedAt)
	for i := 0; i < 4; i++ {
		if commentsList[i].CreatedAt == nil || commentsList[i+1].CreatedAt == nil {
			t.Errorf("CreatedAt should not be nil for comments")
			continue
		}
		if !commentsList[i].CreatedAt.Before(*commentsList[i+1].CreatedAt) {
			t.Errorf("comments not sorted correctly: %v should be before %v", commentsList[i].Text, commentsList[i+1].Text)
		}
	}
}

// testPaginationWithMultiplePages проверяет пагинацию с несколькими страницами
func testPaginationWithMultiplePages(t *testing.T) {
	repo := memory.NewInMemoryRepository()

	// Создание тестового поста
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

	comments := make([]*domain.Comment, 25)
	for i := 0; i < 25; i++ {

		now := time.Now().Add(time.Duration(i) * time.Second)
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

	// Проверка страниц (page=1, limit=10; page=2, limit=10; page=3, limit=10)
	for p := 1; p <= 3; p++ {
		commentsList, err := repo.GetCommentsForPost(context.Background(), postID, int32(p), 10)
		if err != nil {
			t.Fatalf("failed to get comments for page %d: %v", p, err)
		}
		expectedLen := 10
		if p == 3 {
			expectedLen = 5
		}
		if len(commentsList) != expectedLen {
			t.Errorf("for page %d, expected %d comments, got %d", p, expectedLen, len(commentsList))
		}

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

// testInvalidPage проверяет обработку некорректного page
func testInvalidPage(t *testing.T) {
	repo := memory.NewInMemoryRepository()

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
	_, err = repo.GetCommentsForPost(context.Background(), postID, 0, 10)
	if err == nil || err.Error() != "page must be greater than 0" {
		t.Errorf("expected error 'page must be greater than 0', got %v", err)
	}

	// Проверка большой page (за пределами данных)
	_, err = repo.GetCommentsForPost(context.Background(), postID, 10, 10)
	if err != nil && err.Error() != "no posts found" {
		t.Errorf("expected error 'no posts found' for large page, got %v", err)
	}
}

// testInvalidLimit проверяет обработку некорректного limit
func testInvalidLimit(t *testing.T) {
	repo := memory.NewInMemoryRepository()

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

	// Проверка limit <= 0
	_, err = repo.GetCommentsForPost(context.Background(), postID, 1, 0)
	if err == nil || err.Error() != "limit must be greater than 0" {
		t.Errorf("expected error 'limit must be greater than 0', got %v", err)
	}

	// Проверка limit < 0
	_, err = repo.GetCommentsForPost(context.Background(), postID, 1, -1)
	if err == nil || err.Error() != "limit must be greater than 0" {
		t.Errorf("expected error 'limit must be greater than 0', got %v", err)
	}
}

// testNoComments проверяет поведение при отсутствии комментариев
func testNoComments(t *testing.T) {
	repo := memory.NewInMemoryRepository()

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
	comments, err := repo.GetCommentsForPost(context.Background(), postID, 1, 10)
	if err != nil {
		t.Fatalf("failed to get comments: %v", err)
	}
	if len(comments) != 0 {
		t.Errorf("expected 0 comments, got %d", len(comments))
	}
}
