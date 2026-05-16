import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useTranslation } from "react-i18next";
import { z } from "zod";
import {
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  Form,
  InputField,
} from "@eop/ui";
import { ArrowRight, Building2, Loader2 } from "lucide-react";

const companyFormSchema = z.object({
  name: z.string().trim().min(2, "company.errors.too_short").max(100, "company.errors.too_long"),
});

type CompanyForm = z.infer<typeof companyFormSchema>;

export function CompanyStep({
  busy,
  onSubmit,
}: {
  busy: boolean;
  onSubmit: (name: string) => void;
}) {
  const { t } = useTranslation("onboarding");
  const form = useForm<CompanyForm>({
    defaultValues: { name: "" },
    resolver: zodResolver(companyFormSchema),
  });

  return (
    <Card className="card-hover reveal">
      <CardHeader>
        <CardTitle className="font-display tracking-tight flex items-center gap-2">
          <Building2 className="h-5 w-5" />
          {t("company.title")}
        </CardTitle>
        <CardDescription>{t("company.lead")}</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <Form {...form}>
          <form onSubmit={form.handleSubmit((v) => onSubmit(v.name.trim()))} className="space-y-4">
            <InputField
              control={form.control}
              name="name"
              label={t("company.label")}
              placeholder={t("company.placeholder")}
              disabled={busy}
              translateError={(msg) => (msg ? t(msg, { defaultValue: msg }) : undefined)}
            />
            <div className="flex justify-end">
              <Button type="submit" disabled={busy || !form.formState.isValid}>
                {busy ? <Loader2 className="h-4 w-4 mr-2 animate-spin" /> : null}
                {t("company.submit")} <ArrowRight className="h-4 w-4 ml-2" />
              </Button>
            </div>
          </form>
        </Form>
      </CardContent>
    </Card>
  );
}
