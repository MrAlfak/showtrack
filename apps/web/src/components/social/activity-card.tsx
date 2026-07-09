"use client";

import Image from "next/image";
import Link from "next/link";
import { Film, PlayCircle, Plus, RefreshCw, Tv } from "lucide-react";
import { useLocale } from "@/components/locale/locale-provider";
import type { ActivityItem } from "@/lib/api";
import { Card, CardContent } from "@/components/ui/card";

type ActivityCardProps = {
  item: ActivityItem;
};

function activityText(item: ActivityItem, t: ReturnType<typeof import("@/lib/i18n").getTranslations>) {
  const name = item.display_name || item.username || t.social.someone;
  const title = item.payload.title ?? t.social.unknownTitle;
  switch (item.activity_type) {
    case "episode_watched":
      return t.social.watchedEpisode
        .replace("{name}", name)
        .replace("{title}", title)
        .replace("{season}", String(item.payload.season_number ?? "?"))
        .replace("{episode}", String(item.payload.episode_number ?? "?"));
    case "movie_watched":
      return t.social.watchedMovie.replace("{name}", name).replace("{title}", title);
    case "show_added":
      return t.social.addedShow.replace("{name}", name).replace("{title}", title);
    case "movie_added":
      return t.social.addedMovie.replace("{name}", name).replace("{title}", title);
    case "status_changed":
      return t.social.changedStatus
        .replace("{name}", name)
        .replace("{title}", title)
        .replace("{status}", item.payload.list_status ?? "");
    default:
      return `${name} ${title}`;
  }
}

function activityIcon(type: string) {
  switch (type) {
    case "episode_watched":
      return PlayCircle;
    case "movie_watched":
      return Film;
    case "movie_added":
      return Plus;
    case "status_changed":
      return RefreshCw;
    default:
      return Tv;
  }
}

function mediaHref(item: ActivityItem) {
  const tmdbId = item.payload.tmdb_id;
  if (!tmdbId) return "#";
  return item.payload.media_type === "movie" ? `/movies/${tmdbId}` : `/shows/${tmdbId}`;
}

export function ActivityCard({ item }: ActivityCardProps) {
  const { t, locale } = useLocale();
  const Icon = activityIcon(item.activity_type);
  const when = item.created_at
    ? new Date(item.created_at).toLocaleString(locale === "fa" ? "fa-IR" : "en-US", {
        month: "short",
        day: "numeric",
        hour: "2-digit",
        minute: "2-digit",
      })
    : "";

  return (
    <Card className="tv-card shadow-none">
      <CardContent className="flex gap-3 p-4">
        <Link href={`/users/${item.user_id}`} className="shrink-0">
          {item.avatar_url ? (
            <Image
              src={item.avatar_url}
              alt={item.display_name}
              width={40}
              height={40}
              className="size-10 rounded-full object-cover"
              unoptimized
            />
          ) : (
            <div className="flex size-10 items-center justify-center rounded-full bg-secondary text-sm font-semibold">
              {(item.display_name || item.username || "?").slice(0, 1).toUpperCase()}
            </div>
          )}
        </Link>
        <div className="min-w-0 flex-1 space-y-2">
          <div className="flex items-start gap-2">
            <Icon className="mt-0.5 size-4 shrink-0 text-[var(--tv-yellow)]" />
            <div className="min-w-0">
              <p className="text-sm leading-snug">{activityText(item, t)}</p>
              <p className="text-[10px] text-muted-foreground">{when}</p>
            </div>
          </div>
          {item.payload.poster_url ? (
            <Link href={mediaHref(item)} className="inline-block">
              <Image
                src={item.payload.poster_url}
                alt={item.payload.title ?? ""}
                width={56}
                height={84}
                className="rounded-md object-cover"
                unoptimized
              />
            </Link>
          ) : null}
        </div>
      </CardContent>
    </Card>
  );
}
