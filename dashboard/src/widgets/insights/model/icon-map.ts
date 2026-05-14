import {
  Sparkles,
  TrendingDown,
  TrendingUp,
  Minus,
  Zap,
  Languages,
  Calendar,
  Clock,
} from "lucide-react";
import type { LucideIcon } from "lucide-react";
export const INSIGHT_ICON_MAP: Record<
  string,
  {
    icon: LucideIcon;
    color: string;
  }
> = {
  ai_trend_up: { icon: TrendingUp, color: "text-purple-500" },
  ai_trend_down: { icon: TrendingDown, color: "text-blue-500" },
  ai_trend_flat: { icon: Minus, color: "text-muted-foreground" },
  ai_trend_started: { icon: Sparkles, color: "text-purple-500" },
  ai_ratio: { icon: Zap, color: "text-warning" },
  top_lang: { icon: Languages, color: "text-success" },
  productive_day: { icon: Calendar, color: "text-blue-500" },
  total_activity: { icon: Clock, color: "text-muted-foreground" },
};
export const FALLBACK_INSIGHT_ICON = { icon: Sparkles, color: "text-muted-foreground" };
