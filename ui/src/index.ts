// Публичный API @eop/ui — FSD-подобная структура внутри пакета:
//   shared/lib   — утилиты (cn)
//   shared/ui    — примитивы / shadcn-компоненты
//   widgets      — композиции из примитивов
//   features     — сценарии с состоянием (императивные диалоги)
export { Button, buttonVariants } from "./shared/ui/button";
export { Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter } from "./shared/ui/card";
export { cn } from "./shared/lib/cn";
export { PlanBadge, type Plan } from "./widgets/plan-badge";
export { EmptyState } from "./widgets/empty-state";
export { Tabs, TabsList, TabsTrigger, TabsContent } from "./shared/ui/tabs";
export {
  Dialog,
  DialogTrigger,
  DialogPortal,
  DialogClose,
  DialogOverlay,
  DialogContent,
  DialogHeader,
  DialogFooter,
  DialogTitle,
  DialogDescription,
} from "./shared/ui/dialog";
export { Stepper, type StepperItem } from "./widgets/stepper";
export { Input } from "./shared/ui/input";
export { Label } from "./shared/ui/label";
export {
  Form,
  FormField,
  FormItem,
  FormLabel,
  FormControl,
  FormDescription,
  FormMessage,
  useFormField,
} from "./shared/ui/form";
export { Eyebrow } from "./shared/ui/eyebrow";
export { PageHeader } from "./widgets/page-header";
export { Skeleton, SkeletonText, SkeletonTable } from "./shared/ui/skeleton";
export { Avatar, AvatarImage, AvatarFallback, getInitials } from "./shared/ui/avatar";
export { Badge, badgeVariants, type BadgeVariant } from "./shared/ui/badge";
export {
  Select,
  SelectGroup,
  SelectValue,
  SelectTrigger,
  SelectContent,
  SelectLabel,
  SelectItem,
  SelectSeparator,
  SimpleSelect,
  type SimpleSelectOption,
} from "./shared/ui/select";
export { ConfirmDialog, ConfirmProvider, useConfirm } from "./features/confirm-dialog";
export {
  AlertDialog,
  AlertDialogTrigger,
  AlertDialogPortal,
  AlertDialogOverlay,
  AlertDialogContent,
  AlertDialogHeader,
  AlertDialogFooter,
  AlertDialogTitle,
  AlertDialogDescription,
  AlertDialogAction,
  AlertDialogCancel,
} from "./shared/ui/alert-dialog";
export { PromptDialog } from "./features/prompt-dialog";
export { DangerZone } from "./widgets/danger-zone";
export { StatTile, type StatTileAccent } from "./widgets/stat-tile";
export { IconButton } from "./shared/ui/icon-button";
export { SecretField } from "./shared/ui/secret-field";
