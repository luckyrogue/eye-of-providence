package adminapp

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type pgStore struct {
	pool *pgxpool.Pool
}

func NewPGStore(pool *pgxpool.Pool) Store {
	return pgStore{pool: pool}
}

func (p pgStore) ListTeams(ctx context.Context, limit, offset int) ([]TeamRow, error) {
	if p.pool == nil {
		return nil, nil
	}
	rows, err := p.pool.Query(ctx, `
		SELECT t.id, t.name, t.plan, t.created_at,
		       t.subscription_plan, t.subscription_until, t.subscription_note,
		       (SELECT count(*) FROM team_members WHERE team_id = t.id) AS member_count,
		       (SELECT u.email FROM team_members tm JOIN users u ON u.id = tm.user_id
		        WHERE tm.team_id = t.id AND tm.role = 'owner' ORDER BY u.created_at LIMIT 1) AS owner_email
		FROM teams t ORDER BY t.created_at DESC
		LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []TeamRow
	for rows.Next() {
		var t TeamRow
		if err := rows.Scan(&t.ID, &t.Name, &t.Plan, &t.CreatedAt,
			&t.SubscriptionPlan, &t.SubscriptionUntil, &t.SubscriptionNote,
			&t.MemberCount, &t.OwnerEmail); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (p pgStore) ListUsers(ctx context.Context, limit, offset int) ([]UserRow, error) {
	if p.pool == nil {
		return nil, nil
	}
	rows, err := p.pool.Query(ctx, `
		SELECT u.id, u.email, COALESCE(u.display_name, u.email), u.global_role, u.created_at,
		       (SELECT count(*) FROM team_members WHERE user_id = u.id) AS teams_count
		FROM users u ORDER BY u.created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []UserRow
	for rows.Next() {
		var u UserRow
		if err := rows.Scan(&u.ID, &u.Email, &u.DisplayName, &u.GlobalRole, &u.CreatedAt, &u.TeamsCount); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

func (p pgStore) Stats(ctx context.Context) (Stats, error) {
	if p.pool == nil {
		return Stats{}, nil
	}
	var st Stats
	_ = p.pool.QueryRow(ctx, "SELECT count(*) FROM users").Scan(&st.UsersTotal)
	_ = p.pool.QueryRow(ctx, "SELECT count(*) FROM teams").Scan(&st.TeamsTotal)
	_ = p.pool.QueryRow(ctx, "SELECT count(*) FROM team_members").Scan(&st.MembersTotal)
	return st, nil
}

func (p pgStore) Revenue(ctx context.Context) (RevenueReport, error) {
	if p.pool == nil {
		return RevenueReport{Currency: "USD", ByPlan: map[string]int{}}, nil
	}
	var rep RevenueReport
	rep.ByPlan = map[string]int{}
	_ = p.pool.QueryRow(ctx, "SELECT COALESCE(SUM(amount_cents), 0) FROM team_payments").Scan(&rep.TotalCents)
	_ = p.pool.QueryRow(ctx,
		"SELECT COALESCE(SUM(amount_cents), 0) FROM team_payments WHERE paid_at > now() - interval '30 days'",
	).Scan(&rep.Last30dCents)
	_ = p.pool.QueryRow(ctx, "SELECT count(DISTINCT team_id) FROM team_payments").Scan(&rep.PayingTeams)
	_ = p.pool.QueryRow(ctx, `
		SELECT currency FROM team_payments
		WHERE currency IS NOT NULL AND currency <> ''
		GROUP BY currency
		ORDER BY count(*) DESC LIMIT 1`).Scan(&rep.Currency)
	if rep.Currency == "" {
		rep.Currency = "USD"
	}
	rows, err := p.pool.Query(ctx,
		"SELECT COALESCE(subscription_plan, 'free') AS plan, count(*) FROM teams GROUP BY 1")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var plan string
			var cnt int
			if err := rows.Scan(&plan, &cnt); err == nil {
				rep.ByPlan[plan] = cnt
			}
		}
	}
	rrows, err := p.pool.Query(ctx, `
		SELECT p.id, p.team_id, t.name, p.amount_cents, p.currency, p.method, p.covers_until, p.paid_at, p.note
		FROM team_payments p
		JOIN teams t ON t.id = p.team_id
		ORDER BY p.paid_at DESC
		LIMIT 10`)
	if err == nil {
		defer rrows.Close()
		for rrows.Next() {
			var r RecentPayment
			if err := rrows.Scan(&r.ID, &r.TeamID, &r.TeamName, &r.AmountCents, &r.Currency, &r.Method, &r.CoversUntil, &r.PaidAt, &r.Note); err == nil {
				rep.Recent = append(rep.Recent, r)
			}
		}
	}
	if rep.Recent == nil {
		rep.Recent = []RecentPayment{}
	}
	return rep, nil
}

func (p pgStore) ListSSOConfigs(ctx context.Context) ([]SSOConfig, error) {
	if p.pool == nil {
		return nil, nil
	}
	rows, err := p.pool.Query(ctx, `
		SELECT sc.team_id, t.name, sc.provider, sc.enabled, COALESCE(sc.oidc_issuer, ''),
		       COALESCE(sc.oidc_client_id, ''), COALESCE(sc.allowed_domains, ARRAY[]::text[]),
		       sc.jit_provision, sc.jit_role, sc.created_at, sc.updated_at
		FROM sso_configs sc
		JOIN teams t ON t.id = sc.team_id
		ORDER BY sc.updated_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []SSOConfig
	for rows.Next() {
		var e SSOConfig
		if err := rows.Scan(&e.TeamID, &e.TeamName, &e.Provider, &e.Enabled, &e.OIDCIssuer,
			&e.OIDCClientID, &e.AllowedDomains, &e.JITProvision, &e.JITRole, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (p pgStore) DisableSSO(ctx context.Context, teamID uuid.UUID) error {
	if p.pool == nil {
		return ErrSSONotConfigured
	}
	tag, err := p.pool.Exec(ctx,
		"UPDATE sso_configs SET enabled = false, updated_at = now() WHERE team_id = $1", teamID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrSSONotConfigured
	}
	return nil
}

func (p pgStore) ListTeamPayments(ctx context.Context, teamID uuid.UUID) ([]PaymentRow, error) {
	if p.pool == nil {
		return nil, nil
	}
	rows, err := p.pool.Query(ctx, `
		SELECT id, amount_cents, currency, method, note, covers_until, paid_at, recorded_by
		FROM team_payments WHERE team_id=$1 ORDER BY paid_at DESC LIMIT 200`, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []PaymentRow
	for rows.Next() {
		var pay PaymentRow
		var note *string
		if err := rows.Scan(&pay.ID, &pay.AmountCents, &pay.Currency, &pay.Method, &note,
			&pay.CoversUntil, &pay.PaidAt, &pay.RecordedBy); err != nil {
			return nil, err
		}
		if note != nil {
			pay.Note = *note
		}
		out = append(out, pay)
	}
	return out, rows.Err()
}

func (p pgStore) DeleteTeam(ctx context.Context, teamID uuid.UUID) (DeleteTeamResult, error) {
	if p.pool == nil {
		return DeleteTeamResult{}, nil
	}
	var res DeleteTeamResult
	_ = p.pool.QueryRow(ctx, "SELECT name FROM teams WHERE id=$1", teamID).Scan(&res.TeamName)
	if _, err := p.pool.Exec(ctx, "DELETE FROM teams WHERE id=$1", teamID); err != nil {
		return DeleteTeamResult{}, err
	}
	return res, nil
}

var _ Store = pgStore{}
