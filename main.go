package main

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Le-Chat99/blogAggregator/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	DB *database.Queries
}

func main() {
	godotenv.Load()
	port := os.Getenv("PORT")
	dbURL := os.Getenv("CONN")

	mux := http.NewServeMux()
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	dbQueries := database.New(db)
	cfg := apiConfig{
		DB: dbQueries,
	}

	mux.HandleFunc("GET /v1/healthz", healthz)
	mux.HandleFunc("GET /v1/err", errHfunc)

	mux.HandleFunc("POST /v1/users", cfg.userAdd)
	mux.HandleFunc("GET /v1/users", cfg.middlewareAuth(cfg.getUser))

	mux.HandleFunc("POST /v1/feeds", cfg.middlewareAuth(cfg.postFeeds))
	mux.HandleFunc("GET /v1/feeds", cfg.getAllFeeds)

	mux.HandleFunc("POST /v1/feed_follows", cfg.middlewareAuth(cfg.postFeedFollow))
	mux.HandleFunc("GET /v1/feed_follows", cfg.middlewareAuth(cfg.getFeedFollow))
	mux.HandleFunc("DELETE /v1/feed_follows/{id}", cfg.deleteFeedFollow)

	mux.HandleFunc("GET /v1/posts?limit={limit}", cfg.middlewareAuth(cfg.getPost))

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	const collectionConcurrency = 10
	const collectionInterval = time.Minute
	go FecthNTime(cfg.DB, collectionConcurrency, collectionInterval)

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(srv.ListenAndServe())
}

func healthz(w http.ResponseWriter, r *http.Request) {
	type stat struct {
		Status string `json:"status"`
	}
	respondWithJSON(w, 200, stat{Status: "ok"})
}

func errHfunc(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, 500, "Internal Server Error")
}

func GetBearerKey(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("not auth header included in request")
	}
	splitAuth := strings.Split(authHeader, " ")
	if len(splitAuth) < 2 || splitAuth[0] != "ApiKey" {
		return "", errors.New("malformed authorization header")
	}

	return splitAuth[1], nil
}
