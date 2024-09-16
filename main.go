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
	mux.HandleFunc("POST /v1/feed_follows", cfg.middlewareAuth(cfg.postFeedFollow))
	mux.HandleFunc("GET /v1/feed_follows", cfg.middlewareAuth(cfg.getFeedFollow))
	mux.HandleFunc("DELETE /v1/feed_follows/{id}", cfg.deleteFeedFollow)

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

	createFeedFollowed := database.CreateFollowParams{
		ID:        uuid.New(),
		FeedID:    createfeed.ID,
		UserID:    u.ID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_, err = cfg.DB.CreateFollow(context.Background(), createFeedFollowed)
	if err != nil {
		msg := fmt.Sprintf("Error create follow fail: %s", err)
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

func (cfg *apiConfig) postFeedFollow(w http.ResponseWriter, r *http.Request, u database.User) {
	type Params struct {
		FeedID uuid.UUID `json:"feed_id"`
	}
	decoder := json.NewDecoder(r.Body)
	params := Params{}
	err := decoder.Decode(&params)
	if err != nil {
		msg := fmt.Sprintf("Error decoding parameters: %s", err)
		respondWithError(w, http.StatusInternalServerError, msg)
		return
	}
	createFeedFollowed := database.CreateFollowParams{
		ID:        uuid.New(),
		FeedID:    params.FeedID,
		UserID:    u.ID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	feedfollowed, err := cfg.DB.CreateFollow(context.Background(), createFeedFollowed)
	if err != nil {
		msg := fmt.Sprintf("Error create follow fail: %s", err)
		respondWithError(w, http.StatusInternalServerError, msg)
		return
	}
	respondWithJSON(w, http.StatusCreated, databaseFeedFollowedToFeedFollowed((feedfollowed)))
}

func (cfg *apiConfig) deleteFeedFollow(w http.ResponseWriter, r *http.Request) {
	sid := r.PathValue("id")
	id, err := uuid.Parse(sid)
	if err != nil {
		msg := fmt.Sprintf("Invalid id input: %s", err)
		respondWithError(w, http.StatusPartialContent, msg)
		return
	}
	cfg.DB.DeleteFollow(context.Background(), id)
}

func (cfg *apiConfig) getFeedFollow(w http.ResponseWriter, r *http.Request, u database.User) {
	follows, err := cfg.DB.GetFollowByAPIKey(context.Background(), u.ID)
	if err != nil {
		msg := fmt.Sprintf("Error create follow fail: %s", err)
		respondWithError(w, http.StatusInternalServerError, msg)
		return
	}
	type FollowedFeeds struct {
		FollowedFeeds []FeedFollowed `json:"followed_feeds"`
	}
	var Followlist FollowedFeeds
	for _, follow := range follows {
		Followlist.FollowedFeeds = append(Followlist.FollowedFeeds, databaseFeedFollowedToFeedFollowed(follow))
	}

	respondWithJSON(w, http.StatusOK, Followlist)
}
