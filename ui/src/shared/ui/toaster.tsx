import { Toaster as SonnerToaster, toast } from "sonner";
export type ToasterProps = React.ComponentProps<typeof SonnerToaster>;
export function Toaster(props: ToasterProps) {
  return (
    <SonnerToaster position="bottom-right" toastOptions={{ className: "font-sans" }} {...props} />
  );
}
export { toast };
