package adminapp

import (
	"context"
	"time"

	"github.com/google/uuid"
)

func (s *Service) DeleteUser(ctx context.Context, targetID, actorID uuid.UUID) (DeleteUserResult, error) {
	if targetID == actorID {
		return DeleteUserResult{}, ErrCannotDeleteSelf
	}
	if s.store == nil {
		return DeleteUserResult{}, nil
	}
	res, err := s.store.DeleteUser(ctx, targetID)
	if err != nil {
		return DeleteUserResult{}, err
	}
	if s.userDeleter != nil {
		_ = s.userDeleter.DeleteUserData(ctx, targetID.String())
	}
	return res, nil
}

func (s *Service) UpdateUser(ctx context.Context, in UpdateUserInput) (UpdateUserResult, error) {
	var out UpdateUserResult
	if s.store == nil {
		return out, nil
	}
	if in.GlobalRole != nil {
		role, ok := normalizeGlobalRole(*in.GlobalRole)
		if !ok {
			return out, ErrInvalidGlobalRole
		}
		if in.TargetID == in.ActorID && role != "super_admin" {
			return out, ErrCannotDemoteSelf
		}
		prev, email, err := s.store.UpdateUserRole(ctx, in.TargetID, role)
		if err != nil {
			return out, err
		}
		out.VictimEmail = email
		if prev != role {
			out.RoleChanged = true
			out.PrevRole = prev
			out.NewRole = role
			if s.tokenBumper != nil {
				_ = s.tokenBumper.BumpTokenVersion(ctx, in.TargetID)
			}
		}
	}
	if in.DisplayName != nil {
		dn, ok := normalizeDisplayName(*in.DisplayName)
		if !ok {
			return out, ErrInvalidDisplayName
		}
		if err := s.store.UpdateUserDisplayName(ctx, in.TargetID, dn); err != nil {
			return out, err
		}
	}
	return out, nil
}

func (s *Service) AddMember(ctx context.Context, in AddMemberInput) (AddMemberResult, error) {
	if s.store == nil || s.users == nil {
		return AddMemberResult{}, nil
	}
	email, ok := normalizeEmail(in.Email)
	if !ok {
		return AddMemberResult{}, ErrInvalidEmail
	}
	role, ok := normalizeMemberRole(in.Role)
	if !ok {
		return AddMemberResult{}, ErrInvalidRole
	}
	userID, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return AddMemberResult{}, ErrUserNotFound
	}
	if role == "owner" {
		n, err := s.store.CountOtherOwnedTeams(ctx, userID, in.TeamID)
		if err != nil {
			return AddMemberResult{}, err
		}
		if n > 0 {
			return AddMemberResult{}, ErrOwnerLimit
		}
	}
	if err := s.store.AddMember(ctx, in.TeamID, userID, role); err != nil {
		return AddMemberResult{}, err
	}
	return AddMemberResult{UserID: userID}, nil
}

func (s *Service) SetSubscription(ctx context.Context, in SetSubscriptionInput) (SetSubscriptionResult, error) {
	if s.store == nil {
		return SetSubscriptionResult{}, nil
	}
	prepared := SetSubscriptionInput{
		TeamID:     in.TeamID,
		RecordedBy: in.RecordedBy,
		Plan:       in.Plan,
		Until:      in.Until,
		Note:       in.Note,
	}
	if in.Plan != nil {
		plan, ok := normalizePlan(*in.Plan)
		if !ok {
			return SetSubscriptionResult{}, ErrInvalidPlan
		}
		prepared.Plan = &plan
	}
	if in.Until != nil && *in.Until != "" {
		if _, err := time.Parse(time.RFC3339, *in.Until); err != nil {
			return SetSubscriptionResult{}, ErrInvalidUntil
		}
	}
	if in.Payment != nil {
		if in.Payment.AmountCents <= 0 {
			return SetSubscriptionResult{}, ErrInvalidPayment
		}
		if _, err := time.Parse(time.RFC3339, in.Payment.CoversUntil); err != nil {
			return SetSubscriptionResult{}, ErrInvalidPayment
		}
		prepared.Payment = in.Payment
	}
	return s.store.SetSubscription(ctx, prepared)
}
