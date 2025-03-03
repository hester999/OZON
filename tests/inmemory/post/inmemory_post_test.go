package post

import (
	"OZON/internal/domain"
	"OZON/internal/repository/memory"
	"context"
	"fmt"
	"github.com/google/uuid"
	"testing"
	"time"
)

func TestInMemoryRepository(t *testing.T) {
	t.Run("CreateAndGetPost", testCreateAndGetPost)
	t.Run("GetPostsWithPagination", testGetPostsWithPagination)
	t.Run("CreateAndGetComment", testCreateAndGetComment)
	t.Run("IsCommentsAllowed", testIsCommentsAllowed)
	t.Run("GetCommentsForPostWithHierarchy", testGetCommentsForPostWithHierarchy)

}

func testCreateAndGetPost(t *testing.T) {
	repo := memory.NewInMemoryRepository()

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

func testGetPostsWithPagination(t *testing.T) {
	repo := memory.NewInMemoryRepository()

	posts := make([]*domain.Post, 3)
	for i := 0; i < 3; i++ {
		posts[i] = &domain.Post{
			ID:            uuid.New(),
			Text:          fmt.Sprintf("Test Post %d", i),
			AllowComments: true,
		}

		now := time.Now().Add(time.Duration(i) * time.Second)
		posts[i].CreatedAt = &now
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

	if retrievedPosts[0].Text != "Test Post 0" || retrievedPosts[1].Text != "Test Post 1" {
		t.Errorf("unexpected post order: %+v", retrievedPosts)
	}

	retrievedPosts, err = repo.GetPosts(context.Background(), 2, 2)
	if err != nil {
		t.Fatalf("failed to get posts: %v", err)
	}
	if len(retrievedPosts) != 1 {
		t.Errorf("expected 1 post, got %d", len(retrievedPosts))
	}
	if retrievedPosts[0].Text != "Test Post 2" {
		t.Errorf("unexpected post: %+v", retrievedPosts[0])
	}
}
func testCreateAndGetComment(t *testing.T) {
	repo := memory.NewInMemoryRepository()

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

func testIsCommentsAllowed(t *testing.T) {

	repo := memory.NewInMemoryRepository()

	post := &domain.Post{
		ID:   uuid.New(),
		Text: "Test Post",
	}
	post, err := repo.CreatePost(context.Background(), post, nil)
	if err != nil {
		t.Fatalf("failed to create post: %v", err)
	}

	allowDefault, err := repo.IsCommentsAllowed(context.Background(), post.ID)

	if !allowDefault {
		t.Fatalf("Comment must be allow")
	}

	notAllowWithParam := &domain.Post{
		ID:   uuid.New(),
		Text: "Comment Post",
	}
	all := false
	pointerAll := &all

	notAllowWithParam, err = repo.CreatePost(context.Background(), post, pointerAll)
	if err != nil {
		t.Fatalf("failed to create post: %v", err)
	}

	notAllowParam, err := repo.IsCommentsAllowed(context.Background(), notAllowWithParam.ID)

	if notAllowParam {
		t.Fatalf("Comment must be disable %v", notAllowParam)
	}

	AllowWithParam := &domain.Post{
		ID:   uuid.New(),
		Text: "No Comments Post",
	}
	*pointerAll = true
	AllowWithParam, err = repo.CreatePost(context.Background(), post, pointerAll)
	if err != nil {
		t.Fatalf("failed to create post: %v", err)
	}

	allowParam, err := repo.IsCommentsAllowed(context.Background(), AllowWithParam.ID)
	if err != nil {
		fmt.Println(err)
		return
	}
	if !allowParam {
		t.Fatalf("Comment must be allow %v", notAllowParam)
	}

}

// testGetCommentsForPostWithHierarchy проверяет получение комментариев с иерархией
func testGetCommentsForPostWithHierarchy(t *testing.T) {
	repo := memory.NewInMemoryRepository()

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
