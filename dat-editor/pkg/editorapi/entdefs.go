package editorapi

import (
	"net/http"
	"regexp"
	"sort"
	"strings"

	"github.com/jamesread/greyvar/datlib/entdefs"
)

var entdefNamePattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`)

type entdefSummary struct {
	Name         string `json:"name"`
	Title        string `json:"title"`
	InitialState string `json:"initialState"`
	Solid        bool   `json:"solid"`
	StateCount   int    `json:"stateCount"`
}

type entdefPayload struct {
	Title        string                         `json:"title"`
	InitialState string                         `json:"initialState"`
	Solid        bool                           `json:"solid"`
	States       map[string]entdefs.EntityState `json:"states"`
}

func (s *Server) handleEntdefsCollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listEntdefs(w, r)
	case http.MethodPost:
		s.createEntdef(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) handleEntdefItem(w http.ResponseWriter, r *http.Request) {
	name, ok := nameFromPath("/api/entdefs/", r.URL.Path)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid entdef name")
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.getEntdef(w, r, name)
	case http.MethodPut:
		s.updateEntdef(w, r, name)
	case http.MethodDelete:
		s.deleteEntdef(w, r, name)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) listEntdefs(w http.ResponseWriter, r *http.Request) {
	names, err := entdefs.ListEntdefs()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	sort.Strings(names)

	summaries := make([]entdefSummary, 0, len(names))
	for _, name := range names {
		def, err := entdefs.ReadEntdef(name)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		summaries = append(summaries, entdefSummary{
			Name:         name,
			Title:        def.Title,
			InitialState: def.InitialState,
			Solid:        def.Solid,
			StateCount:   len(def.States),
		})
	}

	writeJSON(w, http.StatusOK, summaries)
}

func (s *Server) getEntdef(w http.ResponseWriter, r *http.Request, name string) {
	def, err := entdefs.ReadEntdef(name)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"name":         name,
		"title":        def.Title,
		"initialState": def.InitialState,
		"solid":        def.Solid,
		"states":       def.States,
	})
}

func (s *Server) createEntdef(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name string `json:"name"`
		entdefPayload
	}

	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	name := strings.TrimSpace(body.Name)
	if !entdefNamePattern.MatchString(name) {
		writeError(w, http.StatusBadRequest, "invalid entdef name")
		return
	}

	if _, err := entdefs.ReadEntdef(name); err == nil {
		writeError(w, http.StatusConflict, "entdef already exists")
		return
	}

	def, err := payloadToEntdef(body.entdefPayload, name)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := entdefs.WriteEntdef(name, def); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"name": name})
}

func (s *Server) updateEntdef(w http.ResponseWriter, r *http.Request, name string) {
	var body entdefPayload
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	def, err := payloadToEntdef(body, name)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := entdefs.WriteEntdef(name, def); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"name": name})
}

func (s *Server) deleteEntdef(w http.ResponseWriter, r *http.Request, name string) {
	if err := entdefs.DeleteEntdef(name); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"deleted": name})
}

func payloadToEntdef(body entdefPayload, name string) (*entdefs.EntityDefinition, error) {
	title := strings.TrimSpace(body.Title)
	if title == "" {
		title = name
	}

	states := body.States
	if states == nil {
		states = map[string]entdefs.EntityState{}
	}

	for key, state := range states {
		if state.Name == "" {
			state.Name = key
			states[key] = state
		}
	}

	return &entdefs.EntityDefinition{
		Title:        title,
		InitialState: strings.TrimSpace(body.InitialState),
		Solid:        body.Solid,
		States:       states,
	}, nil
}
