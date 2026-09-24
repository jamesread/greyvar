package editorapi

import (
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func (s *Server) handleTextures(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	kind := strings.TrimPrefix(r.URL.Path, "/api/textures/")
	kind = strings.TrimSuffix(kind, "/")

	var subdir string
	switch kind {
	case "tiles":
		subdir = filepath.Join("img", "textures", "tiles")
	case "entities":
		subdir = filepath.Join("img", "textures", "entities")
	default:
		writeError(w, http.StatusBadRequest, "unknown texture kind")
		return
	}

	dir := filepath.Join(s.resRoot, subdir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if strings.HasSuffix(strings.ToLower(name), ".png") {
			files = append(files, name)
		}
	}

	sort.Strings(files)
	writeJSON(w, http.StatusOK, files)
}
