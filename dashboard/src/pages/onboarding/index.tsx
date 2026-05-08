import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { Onboarding } from "./Onboarding";
import { useAuth } from "../../shared/hooks/useAuth";
import { useTeams } from "../../entities/team";

export function OnboardingRoute() {
  const { isAuthed } = useAuth();
  const navigate = useNavigate();
  const teams = useTeams();

  useEffect(() => {
    if (!isAuthed) navigate("/login", { replace: true });
  }, [isAuthed, navigate]);

  // Если у юзера уже есть команда — онбординг не нужен.
  useEffect(() => {
    if (teams.isSuccess && teams.data.length > 0) {
      navigate("/dashboard", { replace: true });
    }
  }, [teams.isSuccess, teams.data, navigate]);

  return <Onboarding onFinish={() => navigate("/dashboard", { replace: true })} />;
}
