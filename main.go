package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Le-Chat99/blogAggregator/internal/database"
	"github.com/google/uuid"
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

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

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

func (cfg *apiConfig) postFeeds(w http.ResponseWriter, r *http.Request, u database.User) {
	type Params struct {
		Name string `json:"name"`
		Url  string `json:"url"`
	}
	decoder := json.NewDecoder(r.Body)
	params := Params{}
	err := decoder.Decode(&params)
	if err != nil {
		msg := fmt.Sprintf("Error decoding parameters: %s", err)
		respondWithError(w, http.StatusInternalServerError, msg)
		return
	}

	createfeed := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      params.Name,
		Url:       params.Url,
		UserID:    u.ID,
	}

	feed, err := cfg.DB.CreateFeed(context.Background(), createfeed)
	if err != nil {
		msg := fmt.Sprintf("Error create feed fail: %s", err)
		respondWithError(w, http.StatusInternalServerError, msg)
		return
	}
	respondWithJSON(w, http.StatusCreated, databaseFeedToFeed(feed))
}

func (cfg *apiConfig) getAllFeeds(w http.ResponseWriter, r *http.Request) {
	feeds, err := cfg.DB.GetAllFeeds(context.Background())
	if err != nil {
		msg := fmt.Sprintf("Error get feed fail: %s", err)
		respondWithError(w, http.StatusInternalServerError, msg)
		return
	}
	type AllFeeds struct {
		Feeds []Feed `json:"all_feeds"`
	}
	allFeeds := AllFeeds{}
	for _, f := range feeds {
		allFeeds.Feeds = append(allFeeds.Feeds, databaseFeedToFeed(f))
	}
	respondWithJSON(w, http.StatusOK, allFeeds)
}
