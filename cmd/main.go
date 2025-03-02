//package main
//
//import (
//	"OZON/graph"
//	"OZON/internal/handlers"
//	"OZON/internal/repository/memory"
//	//postgres "OZON/internal/repository/postrges"
//	"OZON/internal/usecases"
//	//"OZON/pkg/storage"
//	"fmt"
//	"github.com/99designs/gqlgen/graphql/handler"
//	"github.com/99designs/gqlgen/graphql/playground"
//	"log"
//	"net/http"
//)
//
//func main() {
//
//	//db, err := storage.NewPostgresDB()
//	//if err != nil {
//	//	fmt.Println(err)
//	//	return
//	//}
//	//PostgresRepo := postgres.NewPostgresRepository(*db)
//	mem := memory.NewInMemoryRepository()
//	NewPost := usecases.NewPostUsecase(mem, mem)
//	NewComment := usecases.NewCommentUsecase(mem, mem)
//	NewResolver := handlers.NewResolver(NewPost, NewComment)
//	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: NewResolver}))
//
//	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
//	http.Handle("/query", srv)
//	log.Fatal(http.ListenAndServe(":8080", nil))
//	fmt.Println(123)
//}

// cmd/server/main.go

// cmd/server/main.go

package main

import (
	"OZON/graph"
	"OZON/internal/config"
	"OZON/internal/handlers"
	"OZON/internal/usecases"
	"OZON/pkg/storage"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"log"
	"net/http"
)

func main() {
	// Подключаемся к PostgreSQL (если понадобится для PostgreSQL)
	db, err := storage.NewPostgresDB()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	cfg, err := config.NewConfig(*db)
	if err != nil {
		log.Fatalf("failed to initialize config: %v", err)
	}

	postUsecase := usecases.NewPostUsecase(cfg.PostRepository, cfg.CommentRepository)
	commentUsecase := usecases.NewCommentUsecase(cfg.PostRepository, cfg.CommentRepository)

	resolver := handlers.NewResolver(postUsecase, commentUsecase)

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)
	log.Fatal(http.ListenAndServe(":8080", nil))

}
