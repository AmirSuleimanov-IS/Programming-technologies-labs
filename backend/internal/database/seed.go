package database

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

// SeedDemo заполняет пустой каталог демонстрационными соревнованиями.
// Даты отсчитываются от текущего дня, чтобы соревнования всегда были в будущем.
func SeedDemo(ctx context.Context, pool *pgxpool.Pool) error {
	var count int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM competitions`).Scan(&count); err != nil {
		return fmt.Errorf("count competitions: %w", err)
	}
	if count > 0 {
		return nil
	}
	_, err := pool.Exec(ctx, `INSERT INTO competitions (title, sport, location, starts_on, capacity, description) VALUES
		('Первенство ЗГУ по волейболу', 'Волейбол', 'Норильск, спорткомплекс ЗГУ', CURRENT_DATE + 14, 12, 'Женские и мужские команды факультетов.'),
		('Лыжная гонка «Полярная звезда»', 'Лыжные гонки', 'Норильск, лыжная база «Оль-Гуль»', CURRENT_DATE + 30, 60, 'Классический стиль, дистанции 5 и 10 км.'),
		('Турнир по мини-футболу', 'Футбол', 'Норильск, ФОК «Арктика»', CURRENT_DATE + 45, 40, 'Команды по 5 человек.'),
		('Открытый турнир по шахматам', 'Шахматы', 'Норильск, библиотека ЗГУ', CURRENT_DATE + 7, 3, 'Быстрые шахматы, 9 туров по швейцарской системе.'),
		('Забег «Северное сияние»', 'Лёгкая атлетика', 'Норильск, Ленинский проспект', CURRENT_DATE + 60, 200, 'Дистанции 3 и 10 км.')`)
	if err != nil {
		return fmt.Errorf("seed competitions: %w", err)
	}
	slog.Info("demo competitions seeded")
	return nil
}
