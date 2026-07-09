"use client";

import Image from "next/image";
import { use, useEffect, useState } from "react";
import { CheckCircle2, Film, Plus, Star } from "lucide-react";
import { useAuth } from "@/components/auth/auth-provider";
import { useLocale } from "@/components/locale/locale-provider";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { Skeleton } from "@/components/ui/skeleton";
import { WatchProvidersSection } from "@/components/watch-providers";
import {
  addMovie,
  getMovie,
  markMovieWatched,
  removeMovie,
  unmarkMovieWatched,
  type MovieDetail,
} from "@/lib/api";
import { CastRow } from "@/components/cast-row";

export default function MoviePage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params);
  const { token, isAuthenticated } = useAuth();
  const { t, locale } = useLocale();
  const [data, setData] = useState<MovieDetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [pending, setPending] = useState(false);

  useEffect(() => {
    async function load() {
      setLoading(true);
      const movie = await getMovie(id, token ?? undefined);
      setData(movie);
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
    return <div className="text-sm text-muted-foreground">{t.detail.notFoundMovie}</div>;
  }

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
            <Badge>{t.common.movie}</Badge>
            {data.runtime ? <Badge variant="secondary">{data.runtime} min</Badge> : null}
            {data.vote_average ? (
              <Badge variant="secondary" className="gap-1 text-amber-400">
                <Star className="size-3 fill-current" />
                {data.vote_average.toFixed(1)}
              </Badge>
            ) : null}
          </div>
          <h1 className="text-2xl font-bold tracking-tight">{data.title}</h1>
          {data.release_date && (
            <p className="text-sm text-muted-foreground">
              {t.detail.released}{" "}
              {new Date(data.release_date).toLocaleDateString(locale === "fa" ? "fa-IR" : "en-US")}
            </p>
          )}
        </div>
      </div>

      <Card className="border-border/60 bg-card/80 shadow-none">
        <CardContent className="space-y-3 p-4">
          <div className="grid grid-cols-1 gap-2 sm:grid-cols-2">
            <Button
              className="w-full gap-2"
              variant={data.watched ? "secondary" : "default"}
              disabled={!isAuthenticated || !data.in_library || pending}
              onClick={async () => {
                if (!token) return;
                setPending(true);
                const response = data.watched
                  ? await unmarkMovieWatched(data.id, token)
                  : await markMovieWatched(data.id, token);
                if (response?.ok) {
                  const next = await getMovie(id, token);
                  setData(next);
                }
                setPending(false);
              }}
            >
              <CheckCircle2 className="size-4" />
              {data.watched ? t.detail.watched : t.detail.markWatched}
            </Button>
            <Button
              variant={data.in_library ? "secondary" : "outline"}
              className="w-full gap-2"
              disabled={!isAuthenticated || pending}
              onClick={async () => {
                if (!token) return;
                setPending(true);
                if (data.in_library) {
                  await removeMovie(data.id, token);
                } else {
                  await addMovie(Number(id), token);
                }
                const next = await getMovie(id, token);
                setData(next);
                setPending(false);
              }}
            >
              {data.in_library ? <Film className="size-4" /> : <Plus className="size-4" />}
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

      <WatchProvidersSection mediaType="movie" tmdbId={id} />

      <section className="space-y-3">
        <h2 className="text-base font-semibold">{t.detail.cast}</h2>
        <CastRow cast={data.cast ?? []} />
      </section>

      <Separator />
    </div>
  );
}
