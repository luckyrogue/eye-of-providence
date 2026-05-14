import { Skeleton } from "@eop/ui";
export function RouteFallback() {
  return (
    <div className="mx-auto max-w-6xl px-4 sm:px-6 py-12 space-y-4">
      <Skeleton className="h-10 w-48" />
      <Skeleton className="h-32 w-full" />
      <Skeleton className="h-32 w-full" />
    </div>
  );
}
