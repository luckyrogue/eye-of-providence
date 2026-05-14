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
      void navigate(buildLoginURL(location.pathname + location.search), { replace: true });
    }
  }, [isAuthed, navigate, location.pathname, location.search]);
  useEffect(() => {
    if (!status.data) return;
    const completed = status.data.teams_count > 0 && status.data.has_event;
    if (status.data.dismissed || completed) {
      void navigate("/dashboard", { replace: true });
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
  const initialStep = status.data.teams_count === 0 ? "company" : "event";
  const initialTeamID = teams.data?.[0]?.id ?? null;
  function finish() {
    dismiss.mutate(undefined, {
      onSettled: () => {
        void navigate("/dashboard", { replace: true });
      },
    });
  }
  return (
    <Onboarding
      initialStep={initialStep}
      initialTeamID={initialTeamID}
      onFinish={() => {
        finish();
      }}
    />
  );
}
