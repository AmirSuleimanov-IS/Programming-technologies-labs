// Package cache — кэш в Redis по стратегии cache-aside с логированием и метриками каждой операции.
package cache

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"math/rand/v2"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

// Source — откуда пришли данные: из кэша, из БД после промаха или из БД в обход
// недоступного Redis. Пустая строка — кэш выключен.
type Source string

const (
	SourceHit    Source = "hit"
	SourceMiss   Source = "miss"
	SourceBypass Source = "bypass"
)

// pause — сколько не обращаться к Redis после ошибки. Без паузы каждый запрос
// ждал бы таймаута соединения с недоступным Redis.
const pause = 5 * time.Second

var operations = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "cache_operations_total",
	Help: "Операции с кэшем: hit, miss, set, invalidate, bypass, error",
}, []string{"cache", "result"})

// Client — обёртка над Redis. nil-клиент означает «кэш выключен»: Load просто вызывает загрузку.
type Client struct {
	rdb       *redis.Client
	group     singleflight.Group
	downUntil atomic.Int64 // UnixNano: до этого момента чтение идёт мимо Redis
}

func New(rdb *redis.Client) *Client {
	return &Client{rdb: rdb}
}

// event пишет в лог и метрики одну операцию: name — логическое имя ресурса, key — ключ Redis.
func event(ctx context.Context, name, key, result string, err error) {
	operations.WithLabelValues(name, result).Inc()
	attrs := []any{"cache", name, "key", key, "result", result}
	switch {
	case err != nil:
		slog.ErrorContext(ctx, "cache error", append(attrs, "error", err.Error())...)
	case result == "bypass":
		slog.DebugContext(ctx, "cache bypass", attrs...) // во время паузы — на каждый запрос, поэтому DEBUG
	default:
		slog.InfoContext(ctx, "cache "+result, attrs...)
	}
}

// fail фиксирует ошибку Redis и включает паузу (простой circuit breaker).
func (c *Client) fail(ctx context.Context, name, key string, err error) {
	c.downUntil.Store(time.Now().Add(pause).UnixNano())
	event(ctx, name, key, "error", err)
}

func (c *Client) paused() bool {
	return time.Now().UnixNano() < c.downUntil.Load()
}

// Load читает значение по ключу. При промахе вызывает load, кладёт результат в кэш с ttl
// и возвращает его. Одновременные промахи по одному ключу объединяются (singleflight),
// поэтому после истечения TTL в БД идёт один запрос, а не сотня (защита от cache stampede).
// Если Redis недоступен, данные берутся из БД (fail-open): кэш ускоряет, но не ломает сервис.
func Load[T any](ctx context.Context, c *Client, name, key string, ttl time.Duration,
	load func(ctx context.Context) (T, error),
) (T, Source, error) {
	if c == nil {
		v, err := load(ctx)
		return v, "", err
	}
	if c.paused() {
		event(ctx, name, key, "bypass", nil)
		v, err := load(ctx)
		return v, SourceBypass, err
	}

	var value T
	raw, err := c.rdb.Get(ctx, key).Bytes()
	switch {
	case err == nil && json.Unmarshal(raw, &value) == nil:
		event(ctx, name, key, "hit", nil)
		return value, SourceHit, nil
	case err != nil && !errors.Is(err, redis.Nil):
		c.fail(ctx, name, key, err)
		value, err = load(ctx)
		return value, SourceBypass, err
	}

	event(ctx, name, key, "miss", nil)
	shared, err, _ := c.group.Do(key, func() (any, error) {
		v, err := load(ctx)
		if err != nil {
			return v, err // ошибки (например, 404) не кэшируются
		}
		c.set(ctx, name, key, v, ttl)
		return v, nil
	})
	if err != nil {
		var zero T
		return zero, SourceMiss, err
	}
	return shared.(T), SourceMiss, nil
}

func (c *Client) set(ctx context.Context, name, key string, v any, ttl time.Duration) {
	raw, err := json.Marshal(v)
	if err == nil {
		err = c.rdb.Set(ctx, key, raw, ttl).Err()
	}
	if err != nil {
		c.fail(ctx, name, key, err)
		return
	}
	event(ctx, name, key, "set", nil)
}

// Delete удаляет ключи (инвалидация при изменении данных). Пауза здесь не учитывается:
// пропущенная инвалидация оставила бы в кэше устаревшие данные до конца TTL.
func (c *Client) Delete(ctx context.Context, name string, keys ...string) {
	if c == nil {
		return
	}
	for _, key := range keys {
		if err := c.rdb.Del(ctx, key).Err(); err != nil {
			c.fail(ctx, name, key, err)
			continue
		}
		event(ctx, name, key, "invalidate", nil)
	}
}

// Version возвращает номер версии группы ключей (0, если версии ещё нет).
// Ключи группы содержат версию, поэтому Bump делает недоступными сразу все старые ключи,
// а они сами исчезнут по TTL — без поиска ключей по шаблону.
func (c *Client) Version(ctx context.Context, name, versionKey string) int64 {
	if c == nil || c.paused() {
		return 0
	}
	v, err := c.rdb.Get(ctx, versionKey).Int64()
	if err != nil && !errors.Is(err, redis.Nil) {
		c.fail(ctx, name, versionKey, err)
	}
	return v
}

// Bump увеличивает версию группы ключей — инвалидация всех списков разом.
func (c *Client) Bump(ctx context.Context, name, versionKey string) {
	if c == nil {
		return
	}
	if err := c.rdb.Incr(ctx, versionKey).Err(); err != nil {
		c.fail(ctx, name, versionKey, err)
		return
	}
	event(ctx, name, versionKey, "invalidate", nil)
}

// Jitter случайно меняет ttl в пределах ±spread (0.1 = ±10%), чтобы ключи, записанные
// одновременно, не истекали одновременно (формула TTL из лекции «Кэширование»).
func Jitter(ttl time.Duration, spread float64) time.Duration {
	k := 1 + spread*(2*rand.Float64()-1)
	return time.Duration(float64(ttl) * k)
}
