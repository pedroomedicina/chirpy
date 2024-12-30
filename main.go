package main

import (
	"database/sql"
	"encoding/json"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/pedroomedicina/chirpy/internal/database"
	"log"
	"net/http"
	"os"
	"strings"
)

func respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	err := json.NewEncoder(w).Encode(map[string]string{"error": message})
	if err != nil {
		return
	}
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	err := json.NewEncoder(w).Encode(payload)
	if err != nil {
		return
	}
}

func cleanProfanity(text string) string {
	profaneWords := []string{"kerfuffle", "sharbert", "fornax"}
	words := strings.Split(text, " ")
	for i, word := range words {
		lowerWord := strings.ToLower(word)
		for _, profane := range profaneWords {
			if lowerWord == profane {
				words[i] = "****"
				break
			}
		}
	}

	return strings.Join(words, " ")
}

func main() {
	err := godotenv.Load()
	if err != nil {
		return
	}
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}

	apiCfg := &apiConfig{
		db:        db,
		dbQueries: database.New(db),
		platform:  os.Getenv("PLATFORM"),
		jwtSecret: os.Getenv("JWT_SECRET"),
		polkaKey:  os.Getenv("POLKA_KEY"),
	}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("OK"))
		if err != nil {
			return
		}
	})

	fileServer := http.FileServer(http.Dir("public"))
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", fileServer)))
	mux.Handle("GET /admin/metrics", http.HandlerFunc(apiCfg.handleMetrics))
	mux.Handle("POST /admin/reset", http.HandlerFunc(apiCfg.handleReset))

	mux.Handle("POST /api/chirps", http.HandlerFunc(apiCfg.handleCreateChirp))
	mux.Handle("POST /api/users", http.HandlerFunc(apiCfg.handleCreateUser))
	mux.Handle("PUT /api/users", http.HandlerFunc(apiCfg.handleUpdateUser))

	mux.Handle("GET /api/chirps", http.HandlerFunc(apiCfg.handleGetAllChirps))
	mux.Handle("GET /api/chirps/{id}", http.HandlerFunc(apiCfg.handleGetChirpByID))
	mux.Handle("DELETE /api/chirps/{id}", http.HandlerFunc(apiCfg.handleDeleteChirp))

	mux.Handle("POST /api/login", http.HandlerFunc(apiCfg.handleLogin))
	mux.Handle("POST /api/refresh", http.HandlerFunc(apiCfg.handleRefresh))
	mux.Handle("POST /api/revoke", http.HandlerFunc(apiCfg.handleRevoke))
	mux.Handle("POST /api/polka/webhooks", http.HandlerFunc(apiCfg.handlePolkaWebHook))

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	err = server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
