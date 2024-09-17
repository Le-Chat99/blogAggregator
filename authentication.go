package main

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/Le-Chat99/blogAggregator/internal/database"
)

type authedHandler func(http.ResponseWriter, *http.Request, database.User)

func (cfg *apiConfig) middlewareAuth(handler authedHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKey, err := GetBearerKey(r.Header)
		if err != nil {
			msg := fmt.Sprintf("Error get apikey: %s", err)
			respondWithError(w, http.StatusUnauthorized, msg)
			return
		}
		user, err := cfg.DB.GetUserByAPIKey(r.Context(), apiKey)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				respondWithError(w, http.StatusUnauthorized, "Invalid API key")
			} else {
				respondWithError(w, http.StatusInternalServerError, "Error retrieving user")
			}
			return
		}
		handler(w, r, user)
	}
}
