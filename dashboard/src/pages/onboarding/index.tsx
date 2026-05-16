import { useEffect } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { Skeleton } from "@eop/ui";
import { Onboarding } from "./onboarding";
import { useAuth } from "../../entities/session";
import { useTeams } from "../../entities/team";
import { useDismissOnboarding, useOnboardingStatus } from "../../entities/user";
import { buildLoginURL } from "../../shared/lib/redirect";

export function OnboardingRoute() {
  const { isAuthed } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const status = useOnboardingStatus({ enabled: isAuthed });
  const teams = useTeams();
  const dismiss = useDismissOnboarding();

  useEffect(() => {
    if (!isAuthed) {
      navigate(buildLoginURL(location.pathname + location.search), { replace: true });
    }
  }, [isAuthed, navigate, location.pathname, location.search]);

  // Не возвращаем сюда юзера, который:
  //   - явно dismiss'нул wizard, или
  //   - уже завершил все шаги (есть команда + хоть одно событие) — это
  //     случай существующих пользователей до того как мы добавили wizard.
  useEffect(() => {
    if (!status.data) return;
    const completed = status.data.teams_count > 0 && status.data.has_event;
    if (status.data.dismissed || completed) {
      navigate("/dashboard", { replace: true });
    }
  }, [status.data, navigate]);

  if (!status.data || teams.isPending) {
    return (
      <div className="mx-auto max-w-2xl pt-12 space-y-3">
        <Skeleton className="h-10 w-48" />
        <Skeleton className="h-32" />
      </div>
    );
  }

  // Initial step из server-status:
  //   - team ещё нет → начинаем с company
  //   - team есть, событий нет → пропускаем company/install и идём в event
  //     (если юзер вернулся на /onboarding после refresh)
  // Промежуточные шаги (install/invite) — UI-only, server их не отслеживает.
  const initialStep = status.data.teams_count === 0 ? "company" : "event";
  const initialTeamID = teams.data?.[0]?.id ?? null;

  function finish() {
    dismiss.mutate(undefined, {
      onSettled: () => navigate("/dashboard", { replace: true }),
    });
  }

  return <Onboarding initialStep={initialStep} initialTeamID={initialTeamID} onFinish={finish} />;
}
