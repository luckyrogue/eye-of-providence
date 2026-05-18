package adminapp

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (p pgStore) DeleteUser(ctx context.Context, targetID uuid.UUID) (DeleteUserResult, error) {
	if p.pool == nil {
		return DeleteUserResult{}, nil
	}
	var role string
	if err := p.pool.QueryRow(ctx,
		"SELECT global_role FROM users WHERE id=$1", targetID).Scan(&role); err != nil {
		return DeleteUserResult{}, err
	}
	if role == "super_admin" {
		var count int
		_ = p.pool.QueryRow(ctx,
			"SELECT count(*) FROM users WHERE global_role='super_admin'").Scan(&count)
		if count <= 1 {
			return DeleteUserResult{}, ErrLastSuperAdmin
		}
	}
	var email string
	_ = p.pool.QueryRow(ctx, "SELECT email FROM users WHERE id=$1", targetID).Scan(&email)
	if _, err := p.pool.Exec(ctx, "DELETE FROM users WHERE id=$1", targetID); err != nil {
		return DeleteUserResult{}, err
	}
	return DeleteUserResult{Email: email, Role: role}, nil
}

func (p pgStore) UpdateUserRole(ctx context.Context, targetID uuid.UUID, role string) (prevRole, email string, err error) {
	if p.pool == nil {
		return "", "", nil
	}
	_ = p.pool.QueryRow(ctx,
		"SELECT global_role, email FROM users WHERE id=$1", targetID).Scan(&prevRole, &email)
	if _, err = p.pool.Exec(ctx,
		"UPDATE users SET global_role=$1 WHERE id=$2", role, targetID); err != nil {
		return "", "", err
	}
	return prevRole, email, nil
}

func (p pgStore) UpdateUserDisplayName(ctx context.Context, targetID uuid.UUID, displayName string) error {
	if p.pool == nil {
		return nil
	}
	_, err := p.pool.Exec(ctx,
		"UPDATE users SET display_name=$1 WHERE id=$2", displayName, targetID)
	return err
}

func (p pgStore) CountOtherOwnedTeams(ctx context.Context, userID, teamID uuid.UUID) (int, error) {
	if p.pool == nil {
		return 0, nil
	}
	var n int
	err := p.pool.QueryRow(ctx,
		"SELECT count(*) FROM team_members WHERE user_id=$1 AND role='owner' AND team_id<>$2",
		userID, teamID).Scan(&n)
	return n, err
}

func (p pgStore) AddMember(ctx context.Context, teamID, userID uuid.UUID, role string) error {
	if p.pool == nil {
		return nil
	}
	_, err := p.pool.Exec(ctx, `
		INSERT INTO team_members (team_id, user_id, role) VALUES ($1, $2, $3)
		ON CONFLICT (team_id, user_id) DO UPDATE SET role = EXCLUDED.role`,
		teamID, userID, role)
	return err
}

func (p pgStore) SetSubscription(ctx context.Context, in SetSubscriptionInput) (SetSubscriptionResult, error) {
	if p.pool == nil {
		return SetSubscriptionResult{}, nil
	}
	var out SetSubscriptionResult
	var untilTS *time.Time
	clearUntil := false
	if in.Until != nil {
		out.HasUntil = true
		if *in.Until == "" {
			clearUntil = true
			out.ClearUntil = true
		} else {
			ts, err := time.Parse(time.RFC3339, *in.Until)
			if err != nil {
				return SetSubscriptionResult{}, ErrInvalidUntil
			}
			untilTS = &ts
			out.UntilTS = untilTS
		}
	}
	var coversUntil *time.Time
	if in.Payment != nil {
		ts, err := time.Parse(time.RFC3339, in.Payment.CoversUntil)
		if err != nil {
			return SetSubscriptionResult{}, ErrInvalidPayment
		}
		coversUntil = &ts
		out.PaymentMeta = in.Payment
	}

	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return SetSubscriptionResult{}, err
	}
	defer tx.Rollback(ctx)

	if in.Plan != nil {
		out.HasPlan = true
		out.PlanNorm = *in.Plan
		if _, err := tx.Exec(ctx,
			"UPDATE teams SET subscription_plan=$1 WHERE id=$2", out.PlanNorm, in.TeamID); err != nil {
			return SetSubscriptionResult{}, err
		}
	}
	if clearUntil {
		if _, err := tx.Exec(ctx,
			"UPDATE teams SET subscription_until=NULL WHERE id=$1", in.TeamID); err != nil {
			return SetSubscriptionResult{}, err
		}
	} else if untilTS != nil {
		if _, err := tx.Exec(ctx,
			"UPDATE teams SET subscription_until=$1 WHERE id=$2", *untilTS, in.TeamID); err != nil {
			return SetSubscriptionResult{}, err
		}
	}
	if in.Note != nil {
		note := strings.TrimSpace(*in.Note)
		var noteVal any
		if note == "" {
			noteVal = nil
		} else {
			noteVal = note
		}
		if _, err := tx.Exec(ctx,
			"UPDATE teams SET subscription_note=$1 WHERE id=$2", noteVal, in.TeamID); err != nil {
			return SetSubscriptionResult{}, err
		}
	}
	if coversUntil != nil && in.Payment != nil {
		method := strings.TrimSpace(in.Payment.Method)
		if method == "" {
			method = "manual_transfer"
		}
		currency := strings.ToUpper(strings.TrimSpace(in.Payment.Currency))
		if currency == "" {
			currency = "USD"
		}
		pid := uuid.New()
		if _, err := tx.Exec(ctx, `
			INSERT INTO team_payments
			  (id, team_id, amount_cents, currency, method, note, covers_until, recorded_by)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			pid, in.TeamID, in.Payment.AmountCents, currency, method,
			strings.TrimSpace(in.Payment.Note), *coversUntil, in.RecordedBy); err != nil {
			return SetSubscriptionResult{}, err
		}
		out.PaymentID = &pid
	}
	if err := tx.Commit(ctx); err != nil {
		return SetSubscriptionResult{}, err
	}
	return out, nil
}
