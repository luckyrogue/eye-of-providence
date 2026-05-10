// Centralized zod schemas для всех форм. Один источник правды для:
//   - validation rules (max length, regex, refinements)
//   - inferred TS-types (вместо ручных interface)
//   - error messages (i18n keys через .min/.email/.refine errorMessage)
//
// Convention: `xxxSchema` — zod schema, `XxxValues` — inferred TS-type.
// i18n-keys как error messages — translateZodError() резолвит на frontend
// перед передачей в react-hook-form.

import { z } from "zod";

// --- Common primitives ---

const email = z
  .string()
  .min(1, "auth:validation_email_required")
  .regex(/^[^\s@]+@[^\s@]+\.[^\s@]+$/, "auth:validation_email_invalid")
  .max(254, "auth:validation_email_invalid");

const passwordLogin = z.string().min(1, "auth:validation_password_required");
const passwordRegister = z
  .string()
  .min(8, "auth:validation_password_min")
  .max(256, "auth:validation_password_min");

const displayName = z
  .string()
  .min(1, "auth:validation_name_required")
  .max(64, "auth:validation_name_max");

// --- Auth ---

export const loginSchema = z.object({ email, password: passwordLogin });
export type LoginValues = z.infer<typeof loginSchema>;

export const registerSchema = z.object({
  email,
  password: passwordRegister,
  displayName,
});
export type RegisterValues = z.infer<typeof registerSchema>;

// --- Password reset ---

export const forgotPasswordSchema = z.object({ email });
export type ForgotPasswordValues = z.infer<typeof forgotPasswordSchema>;

export const resetPasswordSchema = z
  .object({
    password: passwordRegister,
    confirmPassword: z.string().min(1, "auth:validation_password_required"),
  })
  .refine((data) => data.password === data.confirmPassword, {
    message: "auth:reset_passwords_mismatch",
    path: ["confirmPassword"],
  });
export type ResetPasswordValues = z.infer<typeof resetPasswordSchema>;

// --- Team ---

const teamName = z
  .string()
  .min(1, "team:validation_name_required")
  .max(100, "team:validation_name_max");

export const teamSchema = z.object({ name: teamName });
export type TeamValues = z.infer<typeof teamSchema>;

// --- Invite ---

export const inviteSchema = z.object({
  email: z
    .string()
    .min(1, "auth:validation_email_required")
    .regex(/^[^\s@]+@[^\s@]+\.[^\s@]+$/, "auth:validation_email_invalid"),
});
export type InviteValues = z.infer<typeof inviteSchema>;

// --- API Token ---

export const apiTokenSchema = z.object({
  name: z
    .string()
    .min(1, "developer:validation_name_required")
    .max(64, "developer:validation_name_max"),
  scope: z.enum(["read", "write:ingest", "admin"]),
  ttlDays: z.number().int().min(0).max(365),
});
export type APITokenValues = z.infer<typeof apiTokenSchema>;

// --- Webhook ---

export const webhookSchema = z.object({
  url: z.string().url("developer:validation_url_invalid").max(2048),
  events: z
    .array(z.enum(["commit.ingested", "report.generated", "anomaly.detected"]))
    .min(1, "developer:validation_events_required"),
  format: z.enum(["raw", "slack"]),
});
export type WebhookValues = z.infer<typeof webhookSchema>;
