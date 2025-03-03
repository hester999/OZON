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
