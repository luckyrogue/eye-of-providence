// Re-export admin email-template hooks из entities/admin для удобного
// импорта на уровне feature (FSD: features импортируют entities).
export {
  useEmailTemplates,
  useEmailTemplate,
  useSaveEmailTemplate,
  useRevertEmailTemplate,
  type EmailTemplateKey,
  type EmailTemplateLocale,
  type EmailTemplateRow,
  type EmailTemplateDetail,
  type EmailTemplateSaveReq,
} from "../../entities/admin";
