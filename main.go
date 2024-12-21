package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

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

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handleValidateChirp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
		return
	}

	var chirp struct {
		Body string `json:"body"`
	}
	err := json.NewDecoder(r.Body).Decode(&chirp)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	if len(chirp.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp too long")
		return
	}

	cleanedBody := cleanProfanity(chirp.Body)

	respondWithJSON(w, http.StatusOK, map[string]string{"cleaned_body": cleanedBody})
}

func (cfg *apiConfig) handleMetrics(w http.ResponseWriter, _ *http.Request) {
	hits := cfg.fileserverHits.Load()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	html := `
	<html>
	  <body>
		<h1>Welcome, Chirpy Admin</h1>
		<p>Chirpy has been visited %d times!</p>
	  </body>
	</html>
	`
	_, err := fmt.Fprintf(w, html, hits)
	if err != nil {
		return
	}
}

func (cfg *apiConfig) handleReset(w http.ResponseWriter, _ *http.Request) {
	cfg.fileserverHits.Store(0)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("Hits counter reset to 0"))
	if err != nil {
		return
	}
}

func main() {
	apiCfg := &apiConfig{}
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
	mux.Handle("POST /api/validate_chirp", http.HandlerFunc(apiCfg.handleValidateChirp))

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
