// UserExporter — GDPR Article 20 "Right to data portability" + privacy
// principle из README §8 пункт 6: данные пользователя должны быть
// извлекаемы в machine-readable виде по запросу.
//
// Возвращаем полный список событий (без агрегации) — пользователь сам
// решит, что с ними делать. Backend cap'ит размер response через limit;
// если пользователь имеет миллионы событий, ему понадобится несколько
// пагинированных запросов. На alpha-этапе hard cap = 200k событий (≈
// 100MB JSON), что покрывает ~6 месяцев активного использования.

package store

import (
	"context"

	"github.com/google/uuid"
)

const exportHardCap = 200_000

type UserExporter interface {
	// ExportUserEvents возвращает события пользователя по возрастанию ts.
	// Cap'ит на `exportHardCap` чтобы один запрос не съел всю RAM ingest'а.
	ExportUserEvents(ctx context.Context, userID string) ([]Event, error)
}

func (s *MemoryStore) ExportUserEvents(_ context.Context, userID string) ([]Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Event, 0, len(s.events))
	for _, e := range s.events {
		if e.UserID == userID {
			out = append(out, e)
			if len(out) >= exportHardCap {
				break
			}
		}
	}
	return out, nil
}

func (s *ClickHouseStore) ExportUserEvents(ctx context.Context, userID string) ([]Event, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	// SELECT по PRIMARY KEY (user_id, ts) — ClickHouse прокатит за O(N log).
	// LIMIT и ORDER BY ts ASC — стабильный порядок, юзер увидит хронологию.
	const q = `
		SELECT ts, user_id, device_id, session_id, app_bundle, category, source,
		       ai_provider, ai_channel, project_id, file_lang,
		       duration_ms, chars_in, lines_added, lines_removed
		FROM events
		WHERE user_id = ?
		ORDER BY ts ASC
		LIMIT ?
	`
	rows, err := s.conn.Query(ctx, q, uid, uint64(exportHardCap))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Event
	for rows.Next() {
		var (
			e            Event
			userUUID     uuid.UUID
			deviceUUID   uuid.UUID
			sessionUUID  uuid.UUID
			projectUUID  uuid.UUID
			aiProvider   string
			aiChannel    string
			fileLang     string
		)
		if err := rows.Scan(
			&e.TS, &userUUID, &deviceUUID, &sessionUUID,
			&e.AppBundle, &e.Category, &e.Source,
			&aiProvider, &aiChannel, &projectUUID, &fileLang,
			&e.DurationMS, &e.CharsIn, &e.LinesAdded, &e.LinesRemoved,
		); err != nil {
			return nil, err
		}
		e.UserID = userUUID.String()
		e.DeviceID = deviceUUID.String()
		e.SessionID = sessionUUID.String()
		e.ProjectID = projectUUID.String()
		e.AIProvider = aiProvider
		e.AIChannel = aiChannel
		e.FileLang = fileLang
		out = append(out, e)
	}
	return out, rows.Err()
}

// CachedEventStore просто проксирует к Inner — кэш для export'а
// смысла не имеет (запросы редкие, объём данных большой).
func (s *CachedEventStore) ExportUserEvents(ctx context.Context, userID string) ([]Event, error) {
	if ex, ok := s.Inner.(UserExporter); ok {
		return ex.ExportUserEvents(ctx, userID)
	}
	return nil, nil
}
