package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Le-Chat99/blogAggregator/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) userAdd(w http.ResponseWriter, r *http.Request) {
	type Params struct {
		Name string `json:"name"`
	}
	decoder := json.NewDecoder(r.Body)
	params := Params{}
	err := decoder.Decode(&params)
	if err != nil {
		msg := fmt.Sprintf("Error decoding parameters: %s", err)
		respondWithError(w, http.StatusInternalServerError, msg)
		return
	}
	user := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      params.Name,
	}
	createdUser, err := cfg.DB.CreateUser(r.Context(), user)
	if err != nil {
		msg := fmt.Sprintf("Error create user fail: %s", err)
		respondWithError(w, http.StatusInternalServerError, msg)
		return
	}
	respondWithJSON(w, http.StatusCreated, databaseUserToUser(createdUser))
}

func (cfg *apiConfig) getUser(w http.ResponseWriter, r *http.Request, u database.User) {
	respondWithJSON(w, http.StatusOK, databaseUserToUser(u))
}
