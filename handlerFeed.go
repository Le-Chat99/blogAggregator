package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Le-Chat99/blogAggregator/internal/database"
	"github.com/google/uuid"
)

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
