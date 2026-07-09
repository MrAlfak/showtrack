"use client";

import Image from "next/image";
import { use, useEffect, useState } from "react";
import { CheckCircle2, Plus, Star, Tv2 } from "lucide-react";
import { useAuth } from "@/components/auth/auth-provider";
import { useLocale } from "@/components/locale/locale-provider";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { Separator } from "@/components/ui/separator";
import { Skeleton } from "@/components/ui/skeleton";
import { WatchProvidersSection } from "@/components/watch-providers";
import { addShow, getShow, markWatched, removeShow, type ShowDetail, unmarkWatched } from "@/lib/api";
import { CastRow } from "@/components/cast-row";

export default function ShowPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params);
  const { token, isAuthenticated } = useAuth();
  const { t, locale } = useLocale();
  const [data, setData] = useState<ShowDetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [pendingEpisode, setPendingEpisode] = useState<number | null>(null);
  const [pendingLibrary, setPendingLibrary] = useState(false);

  useEffect(() => {
    async function load() {
      setLoading(true);
      const show = await getShow(id, token ?? undefined);
      setData(show);
      setLoading(false);
    }
    void load();
  }, [id, token]);

  if (loading) {
    return (
      <div className="space-y-4">
        <Skeleton className="aspect-[16/10] rounded-xl" />
        <Skeleton className="h-28 rounded-xl" />
        <Skeleton className="h-40 rounded-xl" />
      </div>
    );
  }

  if (!data) {
    return <div className="text-sm text-muted-foreground">{t.detail.notFoundShow}</div>;
  }

  const progress = data.progress ?? 0;
  const nextEpisode = data.seasons.flatMap((season) => season.episodes).find((episode) => !episode.watched);

  return (
    <div className="space-y-6">
      <div className="relative -mx-4 -mt-4 aspect-[16/10] overflow-hidden sm:mx-0 sm:mt-0 sm:rounded-xl">
        <Image
          src={data.poster_url}
          alt={data.title}
          fill
          className="object-cover"
          priority
          unoptimized
        />
        <div className="absolute inset-0 bg-gradient-to-t from-background via-background/40 to-transparent" />
        <div className="absolute bottom-0 left-0 right-0 space-y-2 p-4">
          <div className="flex flex-wrap gap-2">
            {data.status && (
              <Badge>{data.status === "Returning Series" ? t.common.ongoing : data.status}</Badge>
            )}
            {data.vote_average ? (
              <Badge variant="secondary" className="gap-1 text-amber-400">
                <Star className="size-3 fill-current" />
                {data.vote_average.toFixed(1)}
              </Badge>
            ) : null}
          </div>
          <h1 className="text-2xl font-bold tracking-tight">{data.title}</h1>
        </div>
      </div>

      <Card className="border-border/60 bg-card/80 shadow-none">
        <CardContent className="space-y-3 p-4">
          <div className="flex items-center justify-between text-sm">
            <span className="text-muted-foreground">{t.detail.yourProgress}</span>
            <span className="font-medium">{Math.round(progress)}%</span>
          </div>
          <Progress value={progress} className="h-2" />
          <div className="grid grid-cols-1 gap-2 sm:grid-cols-2">
            <Button
              className="w-full gap-2"
              disabled={!isAuthenticated || !nextEpisode || pendingEpisode === nextEpisode.id}
              onClick={async () => {
                if (!token || !nextEpisode) return;
                setPendingEpisode(nextEpisode.id);
                const response = await markWatched(nextEpisode.id, token);
                if (response?.ok) {
                  const next = await getShow(id, token);
                  setData(next);
                }
                setPendingEpisode(null);
              }}
            >
              <CheckCircle2 className="size-4" />
              {nextEpisode ? t.detail.markNextWatched : t.detail.allCaughtUp}
            </Button>
            <Button
              variant={data.in_library ? "secondary" : "outline"}
              className="w-full gap-2"
              disabled={!isAuthenticated || pendingLibrary}
              onClick={async () => {
                if (!token) return;
                setPendingLibrary(true);
                if (data.in_library) {
                  await removeShow(data.id, token);
                } else {
                  await addShow(Number(id), token);
                }
                const next = await getShow(id, token);
                setData(next);
                setPendingLibrary(false);
              }}
            >
              {data.in_library ? <Tv2 className="size-4" /> : <Plus className="size-4" />}
              {data.in_library ? t.detail.inLibrary : t.detail.addToLibrary}
            </Button>
          </div>
        </CardContent>
      </Card>

      {data.overview && (
        <section className="space-y-2">
          <h2 className="text-base font-semibold">{t.detail.overview}</h2>
          <p className="text-sm leading-relaxed text-muted-foreground">{data.overview}</p>
        </section>
      )}

      <WatchProvidersSection mediaType="tv" tmdbId={id} />

      <section className="space-y-3">
        <h2 className="text-base font-semibold">{t.detail.cast}</h2>
        <CastRow cast={data.cast ?? []} />
      </section>

      <Separator />

      <section className="space-y-3">
        <h2 className="text-base font-semibold">{t.detail.seasons}</h2>
        <div className="space-y-2">
          {(data.seasons ?? []).map((season) => (
            <Card key={season.id} className="border-border/60 bg-card/80 py-0 shadow-none">
              <CardContent className="space-y-3 p-3">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium">{season.name}</p>
                    <p className="text-xs text-muted-foreground">
                      {season.episode_count} {t.detail.episodes}
                    </p>
                  </div>
                  <Badge variant="outline">S{season.season_number}</Badge>
                </div>
                <div className="space-y-2">
                  {season.episodes.map((episode) => (
                    <div key={episode.id} className="rounded-lg border border-border/60 px-3 py-2">
                      <div className="flex items-start justify-between gap-3">
                        <div className="min-w-0">
                          <p className="text-sm font-medium">
                            E{episode.episode_number} · {episode.name}
                          </p>
                          <p className="line-clamp-2 text-xs text-muted-foreground">
                            {episode.overview || t.detail.noSynopsis}
                          </p>
                          {episode.air_date && (
                            <p className="mt-1 text-[11px] text-muted-foreground">
                              {t.detail.airs}{" "}
                              {new Date(episode.air_date).toLocaleDateString(
                                locale === "fa" ? "fa-IR" : "en-US"
                              )}
                            </p>
                          )}
                        </div>
                        <Button
                          size="sm"
                          variant={episode.watched ? "secondary" : "outline"}
                          disabled={!isAuthenticated || pendingEpisode === episode.id}
                          onClick={async () => {
                            if (!token) return;
                            setPendingEpisode(episode.id);
                            const response = episode.watched
                              ? await unmarkWatched(episode.id, token)
                              : await markWatched(episode.id, token);
                            if (response?.ok) {
                              const next = await getShow(id, token);
                              setData(next);
                            }
                            setPendingEpisode(null);
                          }}
                        >
                          {episode.watched ? t.detail.watched : t.detail.watch}
                        </Button>
                      </div>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      </section>
    </div>
  );
}
