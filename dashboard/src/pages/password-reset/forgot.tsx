import { useState } from "react";
import { useForm } from "react-hook-form";
import { Link } from "react-router-dom";
import { Button, Card, CardContent, CardDescription, CardHeader, CardTitle, Input } from "@eop/ui";
import { Mail } from "lucide-react";
import { forgotPassword } from "../../entities/user";
import { useMutationToast } from "../../shared/hooks/use-mutation-toast";

export function ForgotPasswordRoute() {
  const [sent, setSent] = useState(false);
  const runToast = useMutationToast();
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<{ email: string }>();

  async function onSubmit(values: { email: string }) {
    const ok = await runToast(forgotPassword(values.email), {
      error: "Не удалось отправить письмо",
    });
    // Backend всегда возвращает 200 (privacy by design), так что показываем
    // "проверьте почту" даже если email несуществующий — это by design.
    if (ok !== undefined) setSent(true);
  }

  return (
    <div className="relative">
      <div className="dot-grid pointer-events-none absolute inset-x-0 top-0 h-[420px] -z-10 [mask-image:linear-gradient(to_bottom,black,transparent)]" />
      <div className="mx-auto max-w-md pt-16">
        <Card className="card-hover reveal">
          <CardHeader>
            <div className="mx-auto mb-2 h-12 w-12 rounded-full bg-gradient-to-br from-primary to-primary/60 flex items-center justify-center">
              <Mail className="h-6 w-6 text-primary-foreground" />
            </div>
            <CardTitle className="text-center font-display tracking-tight">Сброс пароля</CardTitle>
            <CardDescription className="text-center">
              {sent
                ? "Если такой email зарегистрирован, мы отправили письмо со ссылкой для сброса. Срок действия — 1 час."
                : "Введите email — мы пришлём ссылку для сброса пароля."}
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            {!sent && (
              <form onSubmit={handleSubmit(onSubmit)} className="space-y-3">
                <Input
                  label="Email"
                  type="email"
                  autoComplete="email"
                  placeholder="you@example.com"
                  error={errors.email?.message}
                  {...register("email", {
                    required: "email обязателен",
                    pattern: { value: /^[^\s@]+@[^\s@]+\.[^\s@]+$/, message: "невалидный email" },
                  })}
                />
                <Button type="submit" disabled={isSubmitting} className="w-full">
                  {isSubmitting ? "..." : "Отправить ссылку"}
                </Button>
              </form>
            )}
            <Link
              to="/login"
              className="block w-full text-center text-xs text-muted-foreground hover:text-foreground transition-colors"
            >
              Вернуться ко входу
            </Link>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
