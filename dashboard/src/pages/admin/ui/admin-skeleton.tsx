import { Skeleton } from "@eop/ui";

export function AdminSkeleton() {
  return (
    <div className="space-y-4">
      <div className="kpi-grid">
        {[0, 1, 2, 3].map((i) => (
          <div key={i} className="kpi">
            <Skeleton className="h-3 w-24" />
            <Skeleton className="h-8 w-16" />
            <Skeleton className="h-3 w-32" />
          </div>
        ))}
      </div>
      <div className="eop-card" style={{ minHeight: 320 }}>
        <Skeleton className="h-5 w-40 mb-3" />
        <Skeleton className="h-3 w-64" />
      </div>
    </div>
  );
}
