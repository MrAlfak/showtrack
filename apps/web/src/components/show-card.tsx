"use client";

import Image from "next/image";
import Link from "next/link";
import { Star } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { useLocale } from "@/components/locale/locale-provider";
import type { ShowItem } from "@/lib/api";

type Props = {
  show: ShowItem;
  href?: string;
  compact?: boolean;
};

export function ShowCard({ show, href, compact }: Props) {
  const { t } = useLocale();
  const link = href ?? (show.media_type === "movie" ? `/movies/${show.tmdb_id}` : `/shows/${show.tmdb_id}`);

  return (
    <Link href={link} className="group block shrink-0">
      <Card className="overflow-hidden border-border/60 bg-card/80 py-0 shadow-none transition-colors hover:border-border">
        <CardContent className="p-0">
          <div className={compact ? "relative aspect-[2/3] w-28" : "relative aspect-[2/3] w-full"}>
            <Image
              src={show.poster_url}
              alt={show.title}
              fill
              className="object-cover transition-transform duration-300 group-hover:scale-[1.03]"
              sizes={compact ? "112px" : "200px"}
              unoptimized
            />
            {show.media_type === "movie" ? (
              <Badge className="absolute left-2 top-2 bg-black/70 text-[10px] text-white hover:bg-black/70">
                {t.common.movie}
              </Badge>
            ) : show.status ? (
              <Badge className="absolute left-2 top-2 bg-black/70 text-[10px] text-white hover:bg-black/70">
                {show.status === "Returning Series" ? t.common.ongoing : show.status}
              </Badge>
            ) : null}
          </div>
          {!compact && (
            <div className="space-y-2 p-3">
              <div className="flex items-start justify-between gap-2">
                <h3 className="line-clamp-2 text-sm font-medium leading-tight">{show.title}</h3>
                {show.vote_average ? (
                  <span className="flex shrink-0 items-center gap-0.5 text-xs text-amber-400">
                    <Star className="size-3 fill-current" />
                    {show.vote_average.toFixed(1)}
                  </span>
                ) : null}
              </div>
              {typeof show.progress === "number" && (
                <div className="space-y-1">
                  <Progress value={show.progress} className="h-1.5" />
                  <p className="text-[11px] text-muted-foreground">
                    {show.media_type === "movie"
                      ? show.progress >= 100
                        ? t.profile.watchedLabel
                        : t.profile.notWatched
                      : `${show.watched}/${show.total} episodes`}
                  </p>
                </div>
              )}
            </div>
          )}
        </CardContent>
      </Card>
    </Link>
  );
}
