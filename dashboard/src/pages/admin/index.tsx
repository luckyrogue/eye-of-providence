import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { Admin } from "./admin";
import { useMe } from "../../entities/user";
import { getTz } from "../../shared/lib/tz";
export function AdminRoute() {
  const me = useMe();
  const navigate = useNavigate();
  useEffect(() => {
    if (me.isSuccess && me.data?.global_role !== "super_admin") {
      void navigate("/dashboard", { replace: true });
    }
  }, [me.isSuccess, me.data, navigate]);
  if (me.isLoading) return null;
  if (me.data?.global_role !== "super_admin") return null;
  return <Admin tz={getTz()} />;
}
