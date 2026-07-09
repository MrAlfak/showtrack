"use client";

import Image from "next/image";
import Link from "next/link";
import { Star } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { useLocale } from "@/components/locale/locale-provider";
import type { ShowItem } from "@/lib/api";
import { showProgressClass } from "@/lib/show-progress";
import { cn } from "@/lib/utils";

type Props = {
  show: ShowItem;
  href?: string;
  compact?: boolean;
};

export function ShowCard({ show, href, compact }: Props) {
  const { t } = useLocale();
  const link = href ?? (show.media_type === "movie" ? `/movies/${show.tmdb_id}` : `/shows/${show.tmdb_id}`);
  const progressClass = showProgressClass(show);

  return (
    <Link href={link} className={cn("group block shrink-0 snap-start", compact ? "w-[7.25rem]" : "w-full")}>
      <div className="overflow-hidden rounded-xl bg-card">
        <div className={cn("relative aspect-[2/3] w-full overflow-hidden rounded-xl")}>
          <Image
            src={show.poster_url}
            alt={show.title}
            fill
            className="object-cover transition-transform duration-300 group-hover:scale-[1.04]"
            sizes={compact ? "116px" : "200px"}
            unoptimized
          />
          {show.media_type === "movie" ? (
            <Badge className="absolute left-1.5 top-1.5 border-0 bg-black/75 text-[9px] font-semibold text-white hover:bg-black/75">
              {t.common.movie}
            </Badge>
          ) : show.status ? (
            <Badge className="absolute left-1.5 top-1.5 border-0 bg-black/75 text-[9px] font-semibold text-white hover:bg-black/75">
              {show.status === "Returning Series" ? t.common.ongoing : show.status}
            </Badge>
          ) : null}
          {typeof show.progress === "number" && (
            <div className="absolute inset-x-0 bottom-0 h-1 bg-black/50">
              <div className={cn("h-full transition-all", progressClass)} style={{ width: `${show.progress}%` }} />
            </div>
          )}
        </div>
        {!compact && (
          <div className="space-y-2 px-0.5 py-2.5">
            <div className="flex items-start justify-between gap-2">
              <h3 className="line-clamp-2 text-sm font-semibold leading-tight">{show.title}</h3>
              {show.vote_average ? (
                <span className="flex shrink-0 items-center gap-0.5 text-xs text-[var(--tv-yellow)]">
                  <Star className="size-3 fill-current" />
                  {show.vote_average.toFixed(1)}
                </span>
              ) : null}
            </div>
            {typeof show.progress === "number" && (
              <div className="space-y-1">
                <div className="h-1 overflow-hidden rounded-full bg-white/10">
                  <div className={cn("h-full transition-all", progressClass)} style={{ width: `${show.progress}%` }} />
                </div>
                <p className="text-[11px] text-muted-foreground">
                  {show.media_type === "movie"
                    ? show.progress >= 100
                      ? t.profile.watchedLabel
                      : t.profile.notWatched
                    : `${show.watched}/${show.total} ${t.common.episodes}`}
                </p>
              </div>
            )}
          </div>
        )}
        {compact && (
          <p className="mt-1.5 line-clamp-2 text-[11px] font-medium leading-tight text-foreground/90">{show.title}</p>
        )}
      </div>
    </Link>
  );
}
