import type { ShowItem } from "@/lib/api";

/** TV Time progress bar colors: yellow = watching, green = caught up, purple = finished */
export function showProgressClass(show: ShowItem): string {
  if (show.media_type === "movie") {
    return (show.progress ?? 0) >= 100 ? "bg-emerald-400" : "bg-[var(--tv-yellow)]";
  }
  const finished = show.status === "Ended" || show.status === "Canceled";
  if (finished) return "bg-violet-500";
  if ((show.progress ?? 0) >= 100) return "bg-emerald-400";
  return "bg-[var(--tv-yellow)]";
}
