package editorapi

import (
	"net/http"
	"sort"
	"strings"

	"github.com/jamesread/greyvar/datlib/texdefs"
)

type tiledefSummary struct {
	Name        string `json:"name"`
	Texture     string `json:"texture"`
	Traversable bool   `json:"traversable"`
	HasFile     bool   `json:"hasFile"`
}

type tiledefPayload struct {
	Texture     string `json:"texture"`
	Traversable bool   `json:"traversable"`
}

var defaultTiledefs = []tiledefSummary{
	{Name: "water", Texture: "water.png", Traversable: false, HasFile: false},
	{Name: "barrier", Texture: "barrier.png", Traversable: false, HasFile: false},
}

func (s *Server) handleTiledefsCollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listTiledefs(w, r)
	case http.MethodPost:
		s.createTiledef(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) handleTiledefItem(w http.ResponseWriter, r *http.Request) {
	name, ok := nameFromPath("/api/tiledefs/", r.URL.Path)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid tiledef name")
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.getTiledef(w, r, name)
	case http.MethodPut:
		s.updateTiledef(w, r, name)
	case http.MethodDelete:
		s.deleteTiledef(w, r, name)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) listTiledefs(w http.ResponseWriter, r *http.Request) {
	byName := map[string]tiledefSummary{}

	for _, def := range defaultTiledefs {
		byName[def.Name] = def
	}

	loaded, err := texdefs.ListTiledefs()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	for _, def := range loaded {
		name := strings.TrimSuffix(def.Filename, ".yaml")
		traversable := true
		if def.Traversable != nil {
			traversable = *def.Traversable
		}

		byName[name] = tiledefSummary{
			Name:        name,
			Texture:     def.TextureName,
			Traversable: traversable,
			HasFile:     true,
		}
	}

	out := make([]tiledefSummary, 0, len(byName))
	for _, item := range byName {
		out = append(out, item)
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].Name < out[j].Name
	})

	writeJSON(w, http.StatusOK, out)
}

func (s *Server) getTiledef(w http.ResponseWriter, r *http.Request, name string) {
	def, err := texdefs.ReadTiledef(name)
	if err != nil {
		for _, fallback := range defaultTiledefs {
			if fallback.Name == name {
				writeJSON(w, http.StatusOK, map[string]any{
					"name":        fallback.Name,
					"texture":     fallback.Texture,
					"traversable": fallback.Traversable,
					"hasFile":     false,
				})
				return
			}
		}

		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	traversable := true
	if def.Traversable != nil {
		traversable = *def.Traversable
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"name":        strings.TrimSuffix(def.Filename, ".yaml"),
		"texture":     def.TextureName,
		"traversable": traversable,
		"hasFile":     true,
	})
}

func (s *Server) createTiledef(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name string `json:"name"`
		tiledefPayload
	}

	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	name := strings.TrimSpace(body.Name)
	if name == "" {
		writeError(w, http.StatusBadRequest, "name required")
		return
	}

	def := payloadToTiledef(body.tiledefPayload)
	if err := texdefs.WriteTiledef(name, def); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"name": name})
}

func (s *Server) updateTiledef(w http.ResponseWriter, r *http.Request, name string) {
	var body tiledefPayload
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	def := payloadToTiledef(body)
	if err := texdefs.WriteTiledef(name, def); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"name": name})
}

func (s *Server) deleteTiledef(w http.ResponseWriter, r *http.Request, name string) {
	if err := texdefs.DeleteTiledef(name); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"deleted": name})
}

func payloadToTiledef(body tiledefPayload) *texdefs.TileDefinition {
	traversable := body.Traversable
	return &texdefs.TileDefinition{
		Texture:     strings.TrimSpace(body.Texture),
		Traversable: &traversable,
	}
}
