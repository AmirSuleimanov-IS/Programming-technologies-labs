package application

import (
	"net/http"
	"strconv"

	"sportreg/internal/httpapi"
	"sportreg/internal/user"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(mux *http.ServeMux, guard user.Guard) {
	anyone := guard(user.RoleUser, user.RoleAdmin)
	admin := guard(user.RoleAdmin)
	mux.Handle("GET /api/applications/my", anyone(http.HandlerFunc(h.listMine)))
	mux.Handle("POST /api/applications", anyone(http.HandlerFunc(h.submit)))
	mux.Handle("PATCH /api/applications/{id}", anyone(http.HandlerFunc(h.changeStatus)))
	mux.Handle("GET /api/applications", admin(http.HandlerFunc(h.list)))
}

func (h *Handler) listMine(w http.ResponseWriter, r *http.Request) {
	actor, _ := user.CurrentFrom(r.Context())
	list, err := h.svc.ListMine(r.Context(), actor)
	if err != nil {
		httpapi.WriteAppError(w, r, err)
		return
	}
	httpapi.WriteData(w, http.StatusOK, list, map[string]int{"count": len(list)})
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := Filter{Status: Status(q.Get("status"))}
	switch f.Status {
	case "", StatusPending, StatusApproved, StatusRejected, StatusCancelled:
	default:
		httpapi.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "status: допустимо pending, approved, rejected или cancelled", nil)
		return
	}
	if raw := q.Get("competition_id"); raw != "" {
		id, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || id <= 0 {
			httpapi.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "competition_id должен быть положительным числом", nil)
			return
		}
		f.CompetitionID = id
	}
	list, err := h.svc.List(r.Context(), f)
	if err != nil {
		httpapi.WriteAppError(w, r, err)
		return
	}
	httpapi.WriteData(w, http.StatusOK, list, map[string]int{"count": len(list)})
}

type submitInput struct {
	CompetitionID int64 `json:"competition_id"`
}

func (h *Handler) submit(w http.ResponseWriter, r *http.Request) {
	var in submitInput
	if !httpapi.DecodeJSON(w, r, &in) {
		return
	}
	actor, _ := user.CurrentFrom(r.Context())
	created, err := h.svc.Submit(r.Context(), actor, in.CompetitionID)
	if err != nil {
		httpapi.WriteAppError(w, r, err)
		return
	}
	httpapi.WriteData(w, http.StatusCreated, created, nil)
}

type statusInput struct {
	Status Status `json:"status"`
}

func (h *Handler) changeStatus(w http.ResponseWriter, r *http.Request) {
	id, ok := httpapi.PathID(w, r, "id")
	if !ok {
		return
	}
	var in statusInput
	if !httpapi.DecodeJSON(w, r, &in) {
		return
	}
	actor, _ := user.CurrentFrom(r.Context())
	updated, err := h.svc.ChangeStatus(r.Context(), actor, id, in.Status)
	if err != nil {
		httpapi.WriteAppError(w, r, err)
		return
	}
	httpapi.WriteData(w, http.StatusOK, updated, nil)
}
