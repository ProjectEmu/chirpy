package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "Hits: %d", cfg.fileserverHits.Load())
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	cfg.fileserverHits.Store(0)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte("Hits counter reset to 0"))
}

func main() {
	// Step 1: Create a new http.ServeMux
	mux := http.NewServeMux()

	// Step 2: Create the apiConfig instance
	apiCfg := &apiConfig{}

	// Step 3: Add the readiness endpoint handler (GET only)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Step 4: Update the file server path to "/app/" and use StripPrefix, wrapped with middleware
	fileServer := http.FileServer(http.Dir("."))
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", fileServer)))

	// Step 5: Add the metrics handler (GET only)
	mux.HandleFunc("/metrics", apiCfg.handlerMetrics)

	// Step 6: Add the reset handler (POST only)
	mux.HandleFunc("/reset", apiCfg.handlerReset)

	// Step 7: Create the http.Server struct
	server := &http.Server{
		Addr:    ":8080", // Listen on port 8080
		Handler: mux,     // Use the ServeMux with the handlers
	}

	// Step 8: Start the server using ListenAndServe
	err := server.ListenAndServe()
	if err != nil {
		panic(err) // Panic if the server fails to start
	}
}
