package user

import (
	"net/http"

	"sportreg/internal/httpapi"
)

// Handler — список учётных записей для администратора.
type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Register(mux *http.ServeMux, guard Guard) {
	mux.Handle("GET /api/users", guard(RoleAdmin)(http.HandlerFunc(h.list)))
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	list, err := h.repo.List(r.Context())
	if err != nil {
		httpapi.WriteAppError(w, r, err)
		return
	}
	httpapi.WriteData(w, http.StatusOK, list, map[string]int{"count": len(list)})
}
