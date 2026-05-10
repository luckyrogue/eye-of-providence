import { Toaster as SonnerToaster, toast } from "sonner";

// Тонкая обёртка над Sonner — единая точка для toast UI во всех приложениях.
// Дефолты подобраны под наш design system: bottom-right, font-sans для тостов.
export type ToasterProps = React.ComponentProps<typeof SonnerToaster>;

export function Toaster(props: ToasterProps) {
  return (
    <SonnerToaster position="bottom-right" toastOptions={{ className: "font-sans" }} {...props} />
  );
}

export { toast };
