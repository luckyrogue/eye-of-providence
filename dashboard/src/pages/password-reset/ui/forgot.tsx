import { useState } from "react";
import { Link } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@eop/ui";
import { Mail } from "lucide-react";
import { ForgotPasswordForm } from "./forgot-password-form";

export function ForgotPasswordRoute() {
  const { t } = useTranslation("auth");
  const [sent, setSent] = useState(false);

  return (
    <div className="min-h-screen">
      <div className="mx-auto max-w-md px-4 pt-12 sm:pt-16 pb-12">
        <Card className="card-hover reveal">
          <CardHeader>
            <div className="mx-auto mb-2 h-12 w-12 rounded-full bg-gradient-to-br from-primary to-primary/60 flex items-center justify-center">
              <Mail className="h-6 w-6 text-primary-foreground" />
            </div>
            <CardTitle className="text-center font-display tracking-tight">
              {t("forgot_title")}
            </CardTitle>
            <CardDescription className="text-center">
              {sent ? t("forgot_sent") : t("forgot_lead")}
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            {!sent && <ForgotPasswordForm onSent={() => setSent(true)} />}
            <Link
              to="/login"
              className="block w-full text-center text-xs text-muted-foreground hover:text-foreground transition-colors"
            >
              {t("forgot_back")}
            </Link>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
