package competition

import (
	"context"
	"net/http"
	"strconv"
	"unicode/utf8"

	"sportreg/internal/httpapi"
)

// Handler — REST-контроллер каталога соревнований.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/competitions", h.list)
	mux.HandleFunc("GET /api/competitions/{id}", h.get)
	mux.HandleFunc("POST /api/competitions", h.create)
	mux.HandleFunc("PUT /api/competitions/{id}", h.replace)
	mux.HandleFunc("PATCH /api/competitions/{id}", h.patch)
	mux.HandleFunc("DELETE /api/competitions/{id}", h.delete)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := Filter{Query: q.Get("q"), Sport: q.Get("sport")}
	if !utf8.ValidString(f.Query) || !utf8.ValidString(f.Sport) {
		httpapi.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Параметры запроса должны быть в кодировке UTF-8", nil)
		return
	}
	list, err := h.svc.List(r.Context(), f)
	if err != nil {
		httpapi.WriteAppError(w, r, err)
		return
	}
	httpapi.WriteData(w, http.StatusOK, list, map[string]int{"count": len(list)})
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, ok := httpapi.PathID(w, r, "id")
	if !ok {
		return
	}
	c, err := h.svc.Get(r.Context(), id)
	if err != nil {
		httpapi.WriteAppError(w, r, err)
		return
	}
	httpapi.WriteData(w, http.StatusOK, c, nil)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var in Input
	if !httpapi.DecodeJSON(w, r, &in) {
		return
	}
	c, err := h.svc.Create(r.Context(), in)
	if err != nil {
		httpapi.WriteAppError(w, r, err)
		return
	}
	w.Header().Set("Location", "/api/competitions/"+strconv.FormatInt(c.ID, 10))
	httpapi.WriteData(w, http.StatusCreated, c, nil)
}

func (h *Handler) replace(w http.ResponseWriter, r *http.Request) {
	h.update(w, r, h.svc.Replace)
}

func (h *Handler) patch(w http.ResponseWriter, r *http.Request) {
	h.update(w, r, h.svc.Patch)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request, apply func(ctx context.Context, id int64, in Input) (Competition, error)) {
	id, ok := httpapi.PathID(w, r, "id")
	if !ok {
		return
	}
	var in Input
	if !httpapi.DecodeJSON(w, r, &in) {
		return
	}
	c, err := apply(r.Context(), id, in)
	if err != nil {
		httpapi.WriteAppError(w, r, err)
		return
	}
	httpapi.WriteData(w, http.StatusOK, c, nil)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	id, ok := httpapi.PathID(w, r, "id")
	if !ok {
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		httpapi.WriteAppError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
