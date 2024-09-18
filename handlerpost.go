package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Le-Chat99/blogAggregator/internal/database"
)

func (cfg *apiConfig) getPost(w http.ResponseWriter, r *http.Request, u database.User) {
	limit := 10
	limitStr := r.URL.Query().Get("limit")
	if limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			// Return an error response if limit is not a valid integer
			respondWithError(w, http.StatusBadRequest, "Invalid limit parameter")
			return
		}
	}
	getPostParams := database.GetPostByUserParams{
		UserID: u.ID,
		Limit:  int32(limit),
	}
	posts, err := cfg.DB.GetPostByUser(context.Background(), getPostParams)
	if err != nil {
		msg := fmt.Sprintf("Fail to get posts: %v", err)
		respondWithError(w, http.StatusInternalServerError, msg)
		return
	}
	respondWithJSON(w, http.StatusOK, databasePostsToPosts(posts))
}
