package auth

import (
	"context"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"sportreg/internal/apperr"
	"sportreg/internal/user"
)

// fakeUsers — хранилище пользователей в памяти для unit-тестов сервиса.
type fakeUsers struct {
	byEmail map[string]user.User
}

func newFakeUsers() *fakeUsers {
	return &fakeUsers{byEmail: map[string]user.User{}}
}

func (f *fakeUsers) Create(_ context.Context, u user.User) (user.User, error) {
	if _, ok := f.byEmail[u.Email]; ok {
		return user.User{}, user.ErrEmailTaken
	}
	u.ID = int64(len(f.byEmail) + 1)
	f.byEmail[u.Email] = u
	return u, nil
}

func (f *fakeUsers) GetByEmail(_ context.Context, email string) (user.User, error) {
	u, ok := f.byEmail[email]
	if !ok {
		return user.User{}, user.ErrNotFound
	}
	return u, nil
}

func newTestService() (*Service, *fakeUsers) {
	users := newFakeUsers()
	return NewService(users, NewTokens(testSecret, time.Hour)), users
}

func wantAppErr(t *testing.T, err error, kind apperr.Kind, code string) apperr.Error {
	t.Helper()
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != kind || appErr.Code != code {
		t.Fatalf("ошибка %v, ожидалась %s (вид %d)", err, code, kind)
	}
	return *appErr
}

func TestService_Register(t *testing.T) {
	svc, users := newTestService()
	ctx := context.Background()

	created, err := svc.Register(ctx, RegisterInput{Email: " Ivan@Mail.RU ", FullName: "Иван", Password: "Sport2026"})
	if err != nil {
		t.Fatal(err)
	}
	if created.Role != user.RoleUser || created.Email != "ivan@mail.ru" {
		t.Errorf("создан %+v: ожидалась роль USER и email в нижнем регистре", created)
	}
	stored := users.byEmail["ivan@mail.ru"]
	if stored.PasswordHash == "Sport2026" || bcrypt.CompareHashAndPassword([]byte(stored.PasswordHash), []byte("Sport2026")) != nil {
		t.Error("пароль должен храниться только в виде bcrypt-хеша")
	}

	// тот же email в другом регистре — уже занят
	_, err = svc.Register(ctx, RegisterInput{Email: "IVAN@mail.ru", FullName: "Иван", Password: "Sport2026"})
	wantAppErr(t, err, apperr.KindConflict, "EMAIL_TAKEN")

	_, err = svc.Register(ctx, RegisterInput{Email: "anna@", FullName: "Анна", Password: "123"})
	if appErr := wantAppErr(t, err, apperr.KindValidation, "VALIDATION_ERROR"); len(appErr.Fields) != 2 {
		t.Errorf("ожидались ошибки email и password: %v", appErr.Fields)
	}
}

func TestService_Login(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()
	if _, err := svc.Register(ctx, RegisterInput{Email: "ivan@mail.ru", FullName: "Иван", Password: "Sport2026"}); err != nil {
		t.Fatal(err)
	}

	res, err := svc.Login(ctx, "IVAN@mail.ru", "Sport2026")
	if err != nil {
		t.Fatal(err)
	}
	current, err := svc.tokens.Parse(res.Token)
	if err != nil || current.Role != user.RoleUser || current.ID != res.User.ID {
		t.Errorf("токен: %+v, %v", current, err)
	}

	// неверный пароль и неизвестный email дают одинаковую ошибку
	errWrong := wantAppErr(t, func() error { _, err := svc.Login(ctx, "ivan@mail.ru", "Sport2027"); return err }(), apperr.KindUnauthorized, "UNAUTHORIZED")
	errUnknown := wantAppErr(t, func() error { _, err := svc.Login(ctx, "nobody@mail.ru", "Sport2026"); return err }(), apperr.KindUnauthorized, "UNAUTHORIZED")
	if errWrong.Message != errUnknown.Message {
		t.Errorf("сообщения различаются: %q и %q — по ним можно узнать, зарегистрирован ли email", errWrong.Message, errUnknown.Message)
	}
}

func TestService_EnsureAdmin(t *testing.T) {
	svc, users := newTestService()
	ctx := context.Background()
	for range 2 { // повторный запуск сервера не создаёт второго администратора
		if err := svc.EnsureAdmin(ctx, "Admin@SportReg.local", "Admin12345"); err != nil {
			t.Fatal(err)
		}
	}
	if len(users.byEmail) != 1 || users.byEmail["admin@sportreg.local"].Role != user.RoleAdmin {
		t.Errorf("пользователи: %+v", users.byEmail)
	}
	err := svc.EnsureAdmin(ctx, "root@sportreg.local", "weak")
	wantAppErr(t, err, apperr.KindValidation, "VALIDATION_ERROR")
}
