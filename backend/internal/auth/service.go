// Package auth — регистрация, вход по паролю, JWT и проверка ролей.
package auth

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"golang.org/x/crypto/bcrypt"

	"sportreg/internal/apperr"
	"sportreg/internal/observability"
	"sportreg/internal/user"
)

// UserStore — то, что сервису нужно от хранилища пользователей.
type UserStore interface {
	Create(ctx context.Context, u user.User) (user.User, error)
	GetByEmail(ctx context.Context, email string) (user.User, error)
}

// dummyHash — bcrypt-хеш случайной строки для входа с неизвестным email.
var dummyHash = func() string {
	h, _ := bcrypt.GenerateFromPassword([]byte("timing-equalizer-not-a-password"), bcrypt.DefaultCost)
	return string(h)
}()

type Service struct {
	users  UserStore
	tokens *Tokens
}

func NewService(users UserStore, tokens *Tokens) *Service {
	return &Service{users: users, tokens: tokens}
}

// LoginResult — ответ на успешный вход.
type LoginResult struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      user.User `json:"user"`
}

// Register создаёт пользователя с ролью USER. Роль ADMIN через API получить нельзя.
func (s *Service) Register(ctx context.Context, in RegisterInput) (user.User, error) {
	in = in.normalize()
	if errs := ValidateRegistration(in); len(errs) > 0 {
		return user.User{}, apperr.Validation(errs)
	}
	return s.create(ctx, in, user.RoleUser)
}

func (s *Service) create(ctx context.Context, in RegisterInput, role user.Role) (user.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return user.User{}, err
	}
	created, err := s.users.Create(ctx, user.User{Email: in.Email, FullName: in.FullName, PasswordHash: string(hash), Role: role})
	if errors.Is(err, user.ErrEmailTaken) {
		return user.User{}, apperr.Conflict("EMAIL_TAKEN", "Пользователь с таким email уже зарегистрирован")
	}
	if err != nil {
		return user.User{}, err
	}
	slog.InfoContext(ctx, "user created", "new_user_id", created.ID, "role", created.Role)
	return created, nil
}

// Login проверяет пароль и выдаёт JWT. Для неизвестного email и неверного пароля
// ответ одинаковый, чтобы по нему нельзя было узнать, зарегистрирован ли адрес.
func (s *Service) Login(ctx context.Context, email, password string) (LoginResult, error) {
	in := RegisterInput{Email: email}.normalize()
	u, err := s.users.GetByEmail(ctx, in.Email)
	if err != nil && !errors.Is(err, user.ErrNotFound) {
		return LoginResult{}, err
	}
	reason := "wrong_password"
	if err != nil {
		// Сравнение с фиктивным хешем выравнивает время ответа: иначе по нему
		// можно было бы узнать, что такого email нет
		reason, u.PasswordHash = "unknown_email", dummyHash
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil || reason == "unknown_email" {
		observability.RecordLogin(false)
		slog.WarnContext(ctx, "login failed", "email", MaskEmail(in.Email), "reason", reason, "result", "invalid_credentials")
		return LoginResult{}, apperr.Unauthorized("Неверный email или пароль")
	}
	token, expires, err := s.tokens.Issue(u)
	if err != nil {
		return LoginResult{}, err
	}
	observability.RecordLogin(true)
	slog.InfoContext(ctx, "login succeeded", "login_user_id", u.ID, "role", u.Role, "result", "success")
	return LoginResult{Token: token, ExpiresAt: expires, User: u}, nil
}

// EnsureAdmin создаёт первого администратора из настроек, если такого email ещё нет.
func (s *Service) EnsureAdmin(ctx context.Context, email, password string) error {
	in := RegisterInput{Email: email, FullName: "Администратор", Password: password}.normalize()
	if _, err := s.users.GetByEmail(ctx, in.Email); err == nil {
		return nil
	} else if !errors.Is(err, user.ErrNotFound) {
		return err
	}
	if errs := ValidateRegistration(in); len(errs) > 0 {
		return apperr.Validation(errs)
	}
	_, err := s.create(ctx, in, user.RoleAdmin)
	return err
}
