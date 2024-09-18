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
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    u.ID,
		FeedID:    params.FeedID,
	}
	feedfollowed, err := cfg.DB.CreateFollow(r.Context(), createFeedFollowed)
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
	follows, err := cfg.DB.GetFollowByAPIKey(r.Context(), u.ID)
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
