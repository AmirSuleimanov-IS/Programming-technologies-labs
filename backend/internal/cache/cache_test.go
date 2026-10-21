package cache

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

// newTestClient — кэш поверх miniredis (Redis в памяти процесса).
func newTestClient(t *testing.T) (*Client, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr(), MaxRetries: -1, DialerRetries: 1,
		DialTimeout: 100 * time.Millisecond, ReadTimeout: 100 * time.Millisecond})
	t.Cleanup(func() { _ = rdb.Close() })
	return New(rdb), mr
}

type item struct {
	Title string `json:"title"`
}

// counter — функция загрузки «из БД», считающая вызовы.
func counter(value item, err error) (func(context.Context) (item, error), *atomic.Int32) {
	var calls atomic.Int32
	return func(context.Context) (item, error) {
		calls.Add(1)
		return value, err
	}, &calls
}

func TestLoad_MissThenHit(t *testing.T) {
	c, mr := newTestClient(t)
	ctx := context.Background()
	load, calls := counter(item{Title: "Кубок"}, nil)

	v, src, err := Load(ctx, c, "test", "item:1", time.Minute, load)
	if err != nil || src != SourceMiss || v.Title != "Кубок" {
		t.Fatalf("первое чтение: %v %q %v", v, src, err)
	}
	v, src, err = Load(ctx, c, "test", "item:1", time.Minute, load)
	if err != nil || src != SourceHit || v.Title != "Кубок" {
		t.Fatalf("второе чтение: %v %q %v", v, src, err)
	}
	if calls.Load() != 1 {
		t.Errorf("загрузок из БД: %d, ожидалась 1", calls.Load())
	}
	if ttl := mr.TTL("item:1"); ttl != time.Minute {
		t.Errorf("TTL ключа %v", ttl)
	}

	mr.FastForward(time.Minute + time.Second) // TTL истёк
	if _, src, _ = Load(ctx, c, "test", "item:1", time.Minute, load); src != SourceMiss || calls.Load() != 2 {
		t.Errorf("после TTL: источник %q, загрузок %d", src, calls.Load())
	}
}

func TestLoad_ErrorsAreNotCached(t *testing.T) {
	c, mr := newTestClient(t)
	notFound := errors.New("not found")
	load, calls := counter(item{}, notFound)
	for range 2 {
		if _, _, err := Load(context.Background(), c, "test", "item:404", time.Minute, load); !errors.Is(err, notFound) {
			t.Fatalf("ошибка загрузки потеряна: %v", err)
		}
	}
	if calls.Load() != 2 || mr.Exists("item:404") {
		t.Errorf("ошибка попала в кэш: загрузок %d", calls.Load())
	}
}

func TestLoad_NilClientDisablesCache(t *testing.T) {
	load, calls := counter(item{Title: "x"}, nil)
	_, src, err := Load(context.Background(), (*Client)(nil), "test", "k", time.Minute, load)
	if err != nil || src != "" || calls.Load() != 1 {
		t.Errorf("без кэша: источник %q, загрузок %d, %v", src, calls.Load(), err)
	}
}

// Cache stampede: 50 одновременных промахов по одному ключу дают один запрос к БД
func TestLoad_SingleflightOnConcurrentMisses(t *testing.T) {
	c, _ := newTestClient(t)
	var calls atomic.Int32
	slowLoad := func(context.Context) (item, error) {
		calls.Add(1)
		time.Sleep(50 * time.Millisecond)
		return item{Title: "Кубок"}, nil
	}
	var wg sync.WaitGroup
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if v, _, err := Load(context.Background(), c, "test", "hot", time.Minute, slowLoad); err != nil || v.Title != "Кубок" {
				t.Errorf("Load() = %v, %v", v, err)
			}
		}()
	}
	wg.Wait()
	if calls.Load() != 1 {
		t.Errorf("загрузок из БД: %d, ожидалась 1", calls.Load())
	}
}

// Отказ Redis: данные берутся из БД (fail-open), следующие 5 с Redis не опрашивается
func TestLoad_RedisDownFailsOpen(t *testing.T) {
	c, mr := newTestClient(t)
	ctx := context.Background()
	load, calls := counter(item{Title: "Кубок"}, nil)
	mr.Close()

	v, src, err := Load(ctx, c, "test", "item:1", time.Minute, load)
	if err != nil || src != SourceBypass || v.Title != "Кубок" {
		t.Fatalf("первый запрос при отказе: %v %q %v", v, src, err)
	}
	if !c.paused() {
		t.Fatal("после ошибки Redis не включилась пауза")
	}
	start := time.Now()
	if _, src, _ = Load(ctx, c, "test", "item:1", time.Minute, load); src != SourceBypass {
		t.Errorf("во время паузы источник %q", src)
	}
	if elapsed := time.Since(start); elapsed > 20*time.Millisecond {
		t.Errorf("во время паузы запрос ждал Redis %v", elapsed)
	}
	if calls.Load() != 2 {
		t.Errorf("загрузок из БД: %d", calls.Load())
	}
}

func TestVersionAndBump(t *testing.T) {
	c, _ := newTestClient(t)
	ctx := context.Background()
	if v := c.Version(ctx, "test", "list:version"); v != 0 {
		t.Errorf("начальная версия %d", v)
	}
	c.Bump(ctx, "test", "list:version")
	c.Bump(ctx, "test", "list:version")
	if v := c.Version(ctx, "test", "list:version"); v != 2 {
		t.Errorf("версия после двух Bump: %d", v)
	}
}

func TestDelete(t *testing.T) {
	c, mr := newTestClient(t)
	_ = mr.Set("item:1", "{}")
	c.Delete(context.Background(), "test", "item:1")
	if mr.Exists("item:1") {
		t.Error("ключ не удалён")
	}
}

// Джиттер: все значения в пределах ±10% и не одинаковые
func TestJitter(t *testing.T) {
	const ttl = 5 * time.Minute
	minV, maxV := ttl, ttl
	for range 10_000 {
		v := Jitter(ttl, 0.1)
		if v < ttl*9/10 || v > ttl*11/10 {
			t.Fatalf("Jitter вышел за ±10%%: %v", v)
		}
		minV, maxV = min(minV, v), max(maxV, v)
	}
	if maxV-minV < ttl/10 {
		t.Errorf("разброс слишком мал: %v…%v", minV, maxV)
	}
}
