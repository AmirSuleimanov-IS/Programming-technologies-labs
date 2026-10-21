package competition

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"sportreg/internal/cache"
	"sportreg/internal/httpapi"
	"sportreg/internal/user"
)

// Handler — REST-контроллер каталога соревнований.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Register подключает маршруты: смотреть каталог могут USER и ADMIN, менять — только ADMIN.
func (h *Handler) Register(mux *http.ServeMux, guard user.Guard) {
	anyone := guard(user.RoleUser, user.RoleAdmin)
	admin := guard(user.RoleAdmin)
	mux.Handle("GET /api/competitions", anyone(http.HandlerFunc(h.list)))
	mux.Handle("GET /api/competitions/{id}", anyone(http.HandlerFunc(h.get)))
	mux.Handle("POST /api/competitions", admin(http.HandlerFunc(h.create)))
	mux.Handle("PUT /api/competitions/{id}", admin(http.HandlerFunc(h.replace)))
	mux.Handle("PATCH /api/competitions/{id}", admin(http.HandlerFunc(h.patch)))
	mux.Handle("DELETE /api/competitions/{id}", admin(http.HandlerFunc(h.delete)))
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := Filter{Query: q.Get("q"), Sport: q.Get("sport")}
	if !utf8.ValidString(f.Query) || !utf8.ValidString(f.Sport) {
		httpapi.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Параметры запроса должны быть в кодировке UTF-8", nil)
		return
	}
	list, src, err := h.svc.List(r.Context(), f)
	if err != nil {
		httpapi.WriteAppError(w, r, err)
		return
	}
	httpapi.WriteData(w, http.StatusOK, list, withSource(w, src, map[string]any{"count": len(list)}))
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, ok := httpapi.PathID(w, r, "id")
	if !ok {
		return
	}
	c, src, err := h.svc.Get(r.Context(), id)
	if err != nil {
		httpapi.WriteAppError(w, r, err)
		return
	}
	httpapi.WriteData(w, http.StatusOK, c, withSource(w, src, map[string]any{}))
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

// withSource показывает клиенту, откуда пришли данные: заголовок X-Cache (HIT, MISS, BYPASS)
// и поле meta.cache в теле ответа.
func withSource(w http.ResponseWriter, src cache.Source, meta map[string]any) map[string]any {
	if src == "" {
		if len(meta) == 0 {
			return nil
		}
		return meta
	}
	w.Header().Set("X-Cache", strings.ToUpper(string(src)))
	meta["cache"] = src
	return meta
}
