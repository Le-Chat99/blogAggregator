package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
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
		UpdatedAt: time.Now(),
		Name:      params.Name,
	}
	createdUser, err := cfg.DB.CreateUser(context.Background(), user)
	if err != nil {
		msg := fmt.Sprintf("Error create user fail: %s", err)
		respondWithError(w, http.StatusInternalServerError, msg)
		return
	}
	respondWithJSON(w, http.StatusCreated, createdUser)
}
