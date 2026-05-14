package reports

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const pgTimeout = 5 * time.Second

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
}

func (s *PostgresStore) Save(r Report) {
	ctx, cancel := context.WithTimeout(context.Background(), pgTimeout)
	defer cancel()

	userID, err := uuid.Parse(r.UserID)
	if err != nil {
		log.Printf("postgres reports.Save: bad user_id %q: %v", r.UserID, err)
		return
	}
	id, err := uuid.Parse(r.ID)
	if err != nil {
		id = uuid.New()
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO reports (id, user_id, period, model, body_md, prompt_version, generated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, id, userID, r.Period, r.Model, r.BodyMD, r.PromptVersion, r.GeneratedAt)
	if err != nil {
		log.Printf("postgres reports.Save failed: %v", err)
	}
}

func (s *PostgresStore) ListForUser(userID string, limit int) []Report {
	ctx, cancel := context.WithTimeout(context.Background(), pgTimeout)
	defer cancel()

	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id, user_id, period, model, body_md, prompt_version, generated_at
		FROM reports
		WHERE user_id = $1
		ORDER BY generated_at DESC
		LIMIT $2
	`, uid, limit)
	if err != nil {
		return nil
	}
	defer rows.Close()

	out := make([]Report, 0, limit)
	for rows.Next() {
		var r Report
		var idU, userU uuid.UUID
		if err := rows.Scan(&idU, &userU, &r.Period, &r.Model, &r.BodyMD, &r.PromptVersion, &r.GeneratedAt); err != nil {
			return out
		}
		r.ID = idU.String()
		r.UserID = userU.String()
		out = append(out, r)
	}
	return out
}

func (s *PostgresStore) Get(id, userID string) (Report, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), pgTimeout)
	defer cancel()

	idU, err := uuid.Parse(id)
	if err != nil {
		return Report{}, false
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return Report{}, false
	}
	var r Report
	var idResult, userResult uuid.UUID
	err = s.pool.QueryRow(ctx, `
		SELECT id, user_id, period, model, body_md, prompt_version, generated_at
		FROM reports
		WHERE id = $1 AND user_id = $2
	`, idU, uid).Scan(&idResult, &userResult, &r.Period, &r.Model, &r.BodyMD, &r.PromptVersion, &r.GeneratedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Report{}, false
	}
	if err != nil {
		return Report{}, false
	}
	r.ID = idResult.String()
	r.UserID = userResult.String()
	return r, true
}

func (s *PostgresStore) Close() {
	s.pool.Close()
}
