package observability

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"
)

// В production уровень логов не опускается ниже INFO
func TestNewLogger_ProductionLevel(t *testing.T) {
	tests := []struct {
		name       string
		production bool
		level      slog.Level
		wantRaised bool
		wantDebug  bool
	}{
		{"production, debug поднимается до info", true, slog.LevelDebug, true, false},
		{"production, info", true, slog.LevelInfo, false, false},
		{"development, debug", false, slog.LevelDebug, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger, raised := NewLogger(&buf, tt.production, tt.level)
			logger.Debug("debug record")
			if raised != tt.wantRaised || (buf.Len() > 0) != tt.wantDebug {
				t.Errorf("raised=%v, записан DEBUG: %v", raised, buf.Len() > 0)
			}
		})
	}
}

// Записи внутри запроса получают request_id, метод, путь и пользователя
func TestLogger_AddsRequestFields(t *testing.T) {
	var buf bytes.Buffer
	logger, _ := NewLogger(&buf, false, slog.LevelInfo)
	ctx := withInfo(context.Background(), &RequestInfo{ID: "abc123", Method: "POST", Path: "/api/applications", UserID: 7})
	logger.InfoContext(ctx, "application submitted")

	var rec map[string]any
	if err := json.Unmarshal(buf.Bytes(), &rec); err != nil {
		t.Fatal(err)
	}
	for key, want := range map[string]any{"request_id": "abc123", "method": "POST", "path": "/api/applications", "user_id": float64(7)} {
		if rec[key] != want {
			t.Errorf("%s = %v, ожидалось %v (запись %s)", key, rec[key], want, buf.String())
		}
	}
}
