package adminlists

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type pgListQuerier struct {
	pool *pgxpool.Pool
}

func NewPGListQuerier(pool *pgxpool.Pool) ListQuerier {
	return pgListQuerier{pool: pool}
}

func (p pgListQuerier) ListWebhooks(ctx context.Context, limit, offset int) ([]WebhookRow, error) {
	if p.pool == nil {
		return nil, nil
	}
	rows, err := p.pool.Query(ctx, `
		SELECT w.id, w.user_id, COALESCE(u.email, '') AS user_email, w.url, w.events,
		       w.format, w.active, w.last_delivery_at, w.last_status, w.created_at
		FROM webhooks w
		LEFT JOIN users u ON u.id = w.user_id
		ORDER BY w.created_at DESC
		LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []WebhookRow
	for rows.Next() {
		var r WebhookRow
		if err := rows.Scan(&r.ID, &r.UserID, &r.UserEmail, &r.URL, &r.Events,
			&r.Format, &r.Active, &r.LastDeliveryAt, &r.LastStatus, &r.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (p pgListQuerier) ListAPITokens(ctx context.Context, limit, offset int, includeRevoked bool) ([]APITokenRow, error) {
	if p.pool == nil {
		return nil, nil
	}
	q := `
		SELECT t.id, t.user_id, COALESCE(u.email, '') AS user_email, t.name, t.scope,
		       t.prefix, t.created_at, t.expires_at, t.last_used_at, t.revoked_at
		FROM api_tokens t
		LEFT JOIN users u ON u.id = t.user_id`
	if !includeRevoked {
		q += " WHERE t.revoked_at IS NULL"
	}
	q += " ORDER BY t.created_at DESC LIMIT $1 OFFSET $2"
	rows, err := p.pool.Query(ctx, q, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []APITokenRow
	for rows.Next() {
		var r APITokenRow
		if err := rows.Scan(&r.ID, &r.UserID, &r.UserEmail, &r.Name, &r.Scope,
			&r.Prefix, &r.CreatedAt, &r.ExpiresAt, &r.LastUsedAt, &r.RevokedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
