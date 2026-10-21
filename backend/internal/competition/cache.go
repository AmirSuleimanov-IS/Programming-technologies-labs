package competition

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"sportreg/internal/cache"
)

// Что кэшируется и почему — см. backend/README.md, раздел «Кэширование».
const (
	cacheItem      = "competition"      // одно соревнование: competition:{id}
	cacheList      = "competition_list" // результат поиска: competitions:list:v{N}:q=…:sport=…
	listVersionKey = "competitions:list:version"
	ttlSpread      = 0.1 // джиттер TTL ±10%
)

// CacheTTL — время жизни записей. Инвалидация при записи срабатывает сразу,
// TTL ограничивает устаревание, если инвалидация не дошла (например, Redis был недоступен).
type CacheTTL struct {
	Item time.Duration
	List time.Duration
}

func itemKey(id int64) string {
	return fmt.Sprintf("competition:%d", id)
}

// listKey включает версию списков и нормализованный фильтр: поиск в БД не зависит
// от регистра, поэтому «Футбол» и «футбол» попадают в один ключ.
func listKey(version int64, f Filter) string {
	norm := func(s string) string { return url.QueryEscape(strings.ToLower(strings.TrimSpace(s))) }
	return fmt.Sprintf("competitions:list:v%d:q=%s:sport=%s", version, norm(f.Query), norm(f.Sport))
}

// Invalidate сбрасывает кэш соревнования и всех списков. Вызывается после изменения
// соревнования и после изменения числа одобренных заявок (оно влияет на free_places).
func (s *Service) Invalidate(ctx context.Context, id int64) {
	if id > 0 {
		s.cache.Delete(ctx, cacheItem, itemKey(id))
	}
	s.cache.Bump(ctx, cacheList, listVersionKey)
}

func (s *Service) loadList(ctx context.Context, f Filter) ([]Competition, cache.Source, error) {
	key := listKey(s.cache.Version(ctx, cacheList, listVersionKey), f)
	return cache.Load(ctx, s.cache, cacheList, key, cache.Jitter(s.ttl.List, ttlSpread),
		func(ctx context.Context) ([]Competition, error) { return s.store.List(ctx, f) })
}

func (s *Service) loadItem(ctx context.Context, id int64) (Competition, cache.Source, error) {
	return cache.Load(ctx, s.cache, cacheItem, itemKey(id), cache.Jitter(s.ttl.Item, ttlSpread),
		func(ctx context.Context) (Competition, error) {
			c, err := s.store.Get(ctx, id)
			return c, mapErr(err)
		})
}
