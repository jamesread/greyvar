package editorapi

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	datlib "github.com/jamesread/greyvar/datlib/common"
)

func ResolveResDir(explicit string) string {
	if explicit != "" {
		return explicit
	}

	if env := os.Getenv("GREYVAR_RES"); env != "" {
		return env
	}

	candidates := []string{
		"../res",
		"../../res",
	}

	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			abs, _ := filepath.Abs(candidate)
			return abs
		}
	}

	return "../res"
}

func DatDirLabel() string {
	return datlib.DatDir()
}

type Server struct {
	resRoot string
	mux     *http.ServeMux
}

func NewServer(resRoot string) http.Handler {
	s := &Server{
		resRoot: resRoot,
		mux:     http.NewServeMux(),
	}

	s.mux.HandleFunc("/api/health", s.handleHealth)
	s.mux.HandleFunc("/api/entdefs", s.handleEntdefsCollection)
	s.mux.HandleFunc("/api/entdefs/", s.handleEntdefItem)
	s.mux.HandleFunc("/api/tiledefs", s.handleTiledefsCollection)
	s.mux.HandleFunc("/api/tiledefs/", s.handleTiledefItem)
	s.mux.HandleFunc("/api/textures/", s.handleTextures)
	s.mux.Handle("/res/", http.StripPrefix("/res/", http.FileServer(http.Dir(resRoot))))

	return corsMiddleware(s.mux)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
		"datDir": datlib.DatDir(),
		"resDir": s.resRoot,
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func readJSON(r *http.Request, dst any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(dst)
}

func nameFromPath(prefix, path string) (string, bool) {
	name := strings.TrimPrefix(path, prefix)
	name = strings.TrimSuffix(name, "/")
	if name == "" || strings.Contains(name, "/") {
		return "", false
	}
	return name, true
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
