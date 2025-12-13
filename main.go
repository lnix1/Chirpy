package main

import (
	"net/http"
	"encoding/json"
	"log"
	"fmt"
	"strings"
	"sync/atomic"
)

func main() {
	const filepathRoot = "."
	const port = "8080"

	counter := apiConfig{fileserverHits: atomic.Int32{}}

	mux := http.NewServeMux()
	mux.Handle("/app/", http.StripPrefix("/app", counter.middlewareMetricsInc(http.FileServer(http.Dir(filepathRoot)))))
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("POST /api/validate_chirp", handlerValidateChirp)
	mux.HandleFunc("POST /admin/reset", counter.middlewareMetricsReset)
	mux.HandleFunc("GET /admin/metrics", counter.middlewareMetricsHits)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}

type apiConfig struct {
	fileserverHits	atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsHits(w http.ResponseWriter, r *http.Request) {
	currentHits := cfg.fileserverHits.Load()
	pageData := fmt.Sprintf(`
		<html>
  			<body>
    				<h1>Welcome, Chirpy Admin</h1>
    				<p>Chirpy has been visited %d times!</p>
  			</body>
		</html>`, currentHits)

	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(pageData))
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)

		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) middlewareMetricsReset(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
}

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	type chirpToValidate struct {
		Body	string `json:"body"`
	}

	type responseBody struct {
		Error 	string `json:"error"`
		Valid	bool   `json:"valid"`
		CleanedBody string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(r.Body)
	incomingChirp := chirpToValidate{}
	err := decoder.Decode(&incomingChirp)
	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		resBody := responseBody{Error: "something went wrong"}
		data, err := json.Marshal(resBody)
		if err != nil {
			log.Printf("Error marshalling response body: %s", err)
		}
		log.Printf("Error decoding chirp: %w", err)
		w.Write(data)
		w.WriteHeader(500)
		return
	}

	if len(incomingChirp.Body) > 140 {
		resBody := responseBody{Error: "Chirp is too long"}
		data, err := json.Marshal(resBody)
		if err != nil {
			log.Printf("Error marshalling response body: %s", err)
		}
		w.WriteHeader(400)
		w.Write(data)
		return
	}

	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}
	cleaned := getCleanedBody(incomingChirp.Body, badWords)

	resBody := responseBody{Valid: true, CleanedBody: cleaned}

	data, err := json.Marshal(resBody)
	if err != nil {
		log.Printf("Error marshalling response body: %s", err)
	}
	w.WriteHeader(200)
	w.Write(data)
	return
}

func getCleanedBody(body string, badWords map[string]struct{}) string {
	words := strings.Split(body, " ")
	for i, word := range words {
		loweredWord := strings.ToLower(word)
		if _, ok := badWords[loweredWord]; ok {
			words[i] = "****"
		}
	}
	cleaned := strings.Join(words, " ")
	return cleaned
}
