// Package observability — телеметрия бэкенда: структурированные логи и метрики Prometheus.
package observability

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"io"
	"log/slog"
)

// RequestInfo — сведения о запросе, которые нужны логам и метрикам. Объект создаётся
// в начале обработки и дополняется по ходу: auth записывает пользователя, маршрутизатор — шаблон маршрута.
type RequestInfo struct {
	ID     string
	Method string
	Path   string
	UserID int64
	Role   string
	Route  string
}

type infoKey struct{}

func withInfo(ctx context.Context, info *RequestInfo) context.Context {
	return context.WithValue(ctx, infoKey{}, info)
}

// Info возвращает сведения о текущем запросе или nil вне HTTP-запроса.
func Info(ctx context.Context) *RequestInfo {
	info, _ := ctx.Value(infoKey{}).(*RequestInfo)
	return info
}

func newRequestID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// NewLogger создаёт JSON-логгер. В production уровень не опускается ниже INFO:
// DEBUG-записи там — лишний шум и риск утечки данных. Второй результат — true,
// если уровень пришлось поднять.
func NewLogger(w io.Writer, production bool, level slog.Level) (*slog.Logger, bool) {
	raised := false
	if production && level < slog.LevelInfo {
		level = slog.LevelInfo
		raised = true
	}
	handler := slog.NewJSONHandler(w, &slog.HandlerOptions{Level: level})
	return slog.New(contextHandler{handler}), raised
}

// contextHandler добавляет к каждой записи внутри HTTP-запроса request_id, user_id, метод и путь,
// чтобы все записи одного запроса можно было найти по одному идентификатору.
type contextHandler struct {
	slog.Handler
}

func (h contextHandler) Handle(ctx context.Context, rec slog.Record) error {
	if info := Info(ctx); info != nil {
		if !hasAttr(rec, "method") {
			rec.AddAttrs(slog.String("method", info.Method), slog.String("path", info.Path))
		}
		rec.AddAttrs(slog.String("request_id", info.ID))
		if info.UserID != 0 {
			rec.AddAttrs(slog.Int64("user_id", info.UserID))
		}
	}
	return h.Handler.Handle(ctx, rec)
}

func (h contextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return contextHandler{h.Handler.WithAttrs(attrs)}
}

func (h contextHandler) WithGroup(name string) slog.Handler {
	return contextHandler{h.Handler.WithGroup(name)}
}

func hasAttr(rec slog.Record, key string) bool {
	found := false
	rec.Attrs(func(a slog.Attr) bool {
		found = a.Key == key
		return !found
	})
	return found
}
