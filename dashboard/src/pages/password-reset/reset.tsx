import { useState } from "react";
import { useForm } from "react-hook-form";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import { Button, Card, CardContent, CardDescription, CardHeader, CardTitle, Input } from "@eop/ui";
import { CheckCircle2, Lock } from "lucide-react";
import { resetPassword } from "../../entities/user";
import { useMutationToast } from "../../shared/hooks/use-mutation-toast";

export function ResetPasswordRoute() {
  const [params] = useSearchParams();
  const navigate = useNavigate();
  const token = params.get("token") ?? "";
  const [done, setDone] = useState(false);
  const runToast = useMutationToast();
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
    watch,
  } = useForm<{ password: string; password2: string }>();

  async function onSubmit(values: { password: string; password2: string }) {
    if (values.password !== values.password2) return;
    const ok = await runToast(resetPassword(token, values.password), {
      error: "Не удалось сбросить пароль",
    });
    if (ok !== undefined) {
      setDone(true);
      // Все старые сессии инвалидированы (token_version bumped) — отправляем
      // юзера логиниться заново.
      setTimeout(() => navigate("/login", { replace: true }), 1500);
    }
  }

  if (!token) {
    return (
      <div className="mx-auto max-w-md pt-16">
        <Card>
          <CardHeader>
            <CardTitle className="text-center">Невалидная ссылка</CardTitle>
            <CardDescription className="text-center">
              В URL не хватает токена. Запроси сброс пароля заново.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Link to="/forgot-password" className="block text-center text-sm underline">
              Запросить новую ссылку
            </Link>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="relative">
      <div className="dot-grid pointer-events-none absolute inset-x-0 top-0 h-[420px] -z-10 [mask-image:linear-gradient(to_bottom,black,transparent)]" />
      <div className="mx-auto max-w-md pt-16">
        <Card className="card-hover reveal">
          <CardHeader>
            <div className="mx-auto mb-2 h-12 w-12 rounded-full bg-gradient-to-br from-primary to-primary/60 flex items-center justify-center">
              {done ? (
                <CheckCircle2 className="h-6 w-6 text-primary-foreground" />
              ) : (
                <Lock className="h-6 w-6 text-primary-foreground" />
              )}
            </div>
            <CardTitle className="text-center font-display tracking-tight">
              {done ? "Пароль изменён" : "Новый пароль"}
            </CardTitle>
            <CardDescription className="text-center">
              {done
                ? "Открываю страницу входа…"
                : "Минимум 8 символов. Все активные сессии будут разлогинены."}
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            {!done && (
              <form onSubmit={handleSubmit(onSubmit)} className="space-y-3">
                <Input
                  label="Новый пароль"
                  type="password"
                  autoComplete="new-password"
                  error={errors.password?.message}
                  {...register("password", {
                    required: "пароль обязателен",
                    minLength: { value: 8, message: "минимум 8 символов" },
                  })}
                />
                <Input
                  label="Повторите пароль"
                  type="password"
                  autoComplete="new-password"
                  error={
                    errors.password2?.message ||
                    (watch("password2") && watch("password") !== watch("password2")
                      ? "пароли не совпадают"
                      : undefined)
                  }
                  {...register("password2", { required: "повторите пароль" })}
                />
                <Button type="submit" disabled={isSubmitting} className="w-full">
                  {isSubmitting ? "..." : "Сменить пароль"}
                </Button>
              </form>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
