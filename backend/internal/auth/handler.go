package auth

import (
	"errors"
	"net/http"

	"sportreg/internal/apperr"
	"sportreg/internal/httpapi"
	"sportreg/internal/user"
)

type Handler struct {
	svc   *Service
	users *user.Repository
}

func NewHandler(svc *Service, users *user.Repository) *Handler {
	return &Handler{svc: svc, users: users}
}

func (h *Handler) Register(mux *http.ServeMux, guard user.Guard) {
	mux.HandleFunc("POST /api/auth/register", h.register)
	mux.HandleFunc("POST /api/auth/login", h.login)
	mux.Handle("GET /api/auth/me", guard(user.RoleUser, user.RoleAdmin)(http.HandlerFunc(h.me)))
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var in RegisterInput
	if !httpapi.DecodeJSON(w, r, &in) {
		return
	}
	created, err := h.svc.Register(r.Context(), in)
	if err != nil {
		httpapi.WriteAppError(w, r, err)
		return
	}
	httpapi.WriteData(w, http.StatusCreated, created, nil)
}

type loginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var in loginInput
	if !httpapi.DecodeJSON(w, r, &in) {
		return
	}
	res, err := h.svc.Login(r.Context(), in.Email, in.Password)
	if err != nil {
		httpapi.WriteAppError(w, r, err)
		return
	}
	httpapi.WriteData(w, http.StatusOK, res, nil)
}

// me возвращает учётную запись владельца токена.
func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	current, _ := user.CurrentFrom(r.Context())
	u, err := h.users.GetByID(r.Context(), current.ID)
	if errors.Is(err, user.ErrNotFound) { // учётную запись удалили, а токен ещё действует
		err = apperr.Unauthorized("Учётная запись не найдена")
	}
	if err != nil {
		httpapi.WriteAppError(w, r, err)
		return
	}
	httpapi.WriteData(w, http.StatusOK, u, nil)
}
