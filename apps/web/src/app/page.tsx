"use client";

import { Clapperboard } from "lucide-react";
import Link from "next/link";
import { useEffect, useState } from "react";
import { useAuth } from "@/components/auth/auth-provider";
import { useLocale } from "@/components/locale/locale-provider";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { getDashboard, getRecommendations, trending, trendingMovies, type ShowItem, type UpcomingItem } from "@/lib/api";
import { ShowCard } from "@/components/show-card";
import { SectionHeader } from "@/components/section-header";
import { EmptyState } from "@/components/ui/empty-state";

export default function HomePage() {
  const { token, ready, isAuthenticated } = useAuth();
  const { t, locale } = useLocale();
  const [trendingShows, setTrendingShows] = useState<ShowItem[]>([]);
  const [trendingMovieItems, setTrendingMovieItems] = useState<ShowItem[]>([]);
  const [library, setLibrary] = useState<ShowItem[]>([]);
  const [upcoming, setUpcoming] = useState<UpcomingItem[]>([]);
  const [recommendations, setRecommendations] = useState<ShowItem[]>([]);
  const [recSeed, setRecSeed] = useState("");
  const [recExplanation, setRecExplanation] = useState("");
  const [recEngine, setRecEngine] = useState("");
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    async function load() {
      const [trendingData, moviesData, dashboardData, recData] = await Promise.all([
        trending(),
        trendingMovies(),
        token ? getDashboard(token) : Promise.resolve(null),
        token ? getRecommendations(token) : Promise.resolve(null),
      ]);
      if (cancelled) return;
      setTrendingShows(trendingData?.results ?? []);
      setTrendingMovieItems(moviesData?.results ?? []);
      setLibrary(dashboardData?.library ?? []);
      setUpcoming(dashboardData?.upcoming ?? []);
      setRecommendations(recData?.results ?? []);
      setRecSeed(recData?.seed_title ?? "");
      setRecExplanation(recData?.explanation ?? "");
      setRecEngine(recData?.engine ?? "");
      setLoading(false);
    }

    if (!ready) return;
    void load();
    return () => {
      cancelled = true;
    };
  }, [ready, token]);

  return (
    <div className="space-y-7">
      <section>
        <SectionHeader title={t.home.continueWatching} />
        {!ready || loading ? (
          <div className="tv-scroll-row">
            {Array.from({ length: 3 }).map((_, index) => (
              <Skeleton key={index} className="aspect-[2/3] w-[7.25rem] shrink-0 rounded-xl" />
            ))}
          </div>
        ) : library.length > 0 ? (
          <div className="tv-scroll-row">
            {library.map((item) => (
              <ShowCard key={`${item.media_type ?? "tv"}-${item.id}`} show={item} compact />
            ))}
          </div>
        ) : (
          <EmptyState
            icon={Clapperboard}
            title={isAuthenticated ? t.home.emptyLibrary : t.auth.signInRequired}
            description={isAuthenticated ? undefined : t.auth.signInHint}
            action={
              !isAuthenticated ? (
                <Link
                  href="/login"
                  className="inline-flex h-8 items-center justify-center rounded-lg bg-primary px-3 text-sm font-medium text-primary-foreground"
                >
                  {t.auth.signIn}
                </Link>
              ) : (
                <Link
                  href="/search"
                  className="inline-flex h-8 items-center justify-center rounded-lg border border-border px-3 text-sm font-medium"
                >
                  {t.nav.search}
                </Link>
              )
            }
          />
        )}
      </section>

      {isAuthenticated && recommendations.length > 0 && (
        <section>
          <SectionHeader
            title={recSeed ? `${t.home.becauseYouWatched} ${recSeed}` : t.home.recommended}
            subtitle={recExplanation || (recEngine === "hybrid" ? t.home.mlExplanation : undefined)}
          />
          <div className="tv-scroll-row">
            {recommendations.map((item) => (
              <div key={`rec-${item.media_type ?? "tv"}-${item.tmdb_id}`} className="relative shrink-0">
                <ShowCard show={item} compact />
                {item.rec_reason ? (
                  <span className="absolute bottom-1 left-1 rounded bg-black/75 px-1.5 py-0.5 text-[9px] text-white">
                    {item.rec_reason === "collaborative"
                      ? t.home.recCollaborative
                      : item.rec_reason === "following"
                        ? t.home.recFollowing
                        : item.rec_reason === "genre"
                          ? t.home.recGenre
                          : item.rec_reason === "similar"
                            ? t.home.recSimilar
                            : t.home.recTrending}
                  </span>
                ) : null}
              </div>
            ))}
          </div>
        </section>
      )}

      <section>
        <SectionHeader title={t.home.upcoming} />
        <div className="space-y-2">
          {(upcoming.length ? upcoming : []).map((item) => (
            <Link key={item.episode_id} href={`/shows/${item.show_tmdb_id}`}>
              <Card className="tv-card py-0 shadow-none transition-colors hover:border-white/10">
                <CardContent className="flex items-center gap-3 p-3">
                  <div className="size-14 shrink-0 overflow-hidden rounded-lg bg-muted">
                    {/* eslint-disable-next-line @next/next/no-img-element */}
                    <img src={item.poster_url} alt={item.show_title} className="size-full object-cover" />
                  </div>
                  <div className="min-w-0 flex-1">
                    <p className="truncate text-sm font-semibold">{item.show_title}</p>
                    <p className="text-xs text-muted-foreground">
                      S{item.season_number}E{item.episode_number}
                      {item.episode_name ? ` · ${item.episode_name}` : ""}
                    </p>
                  </div>
                  <Badge className="shrink-0 border-0 bg-[var(--tv-yellow)] font-semibold text-black hover:bg-[var(--tv-yellow)]">
                    {item.air_date
                      ? new Date(item.air_date).toLocaleDateString(locale === "fa" ? "fa-IR" : "en-US", {
                          month: "short",
                          day: "numeric",
                        })
                      : t.home.soon}
                  </Badge>
                </CardContent>
              </Card>
            </Link>
          ))}
          {!loading && upcoming.length === 0 && (
            <Card className="tv-card py-0 shadow-none">
              <CardContent className="p-4 text-sm text-muted-foreground">{t.home.upcomingEmpty}</CardContent>
            </Card>
          )}
        </div>
      </section>

      <section>
        <SectionHeader
          title={t.home.trending}
          action={
            <Link href="/discover" className="text-xs font-medium text-[var(--tv-yellow)] hover:underline">
              {t.home.seeAll}
            </Link>
          }
        />
        <Tabs defaultValue="tv">
          <TabsList className="h-9 w-full rounded-xl bg-secondary/80 p-1">
            <TabsTrigger value="tv" className="flex-1 rounded-lg data-active:bg-[var(--tv-yellow)] data-active:text-black">
              {t.home.tvShows}
            </TabsTrigger>
            <TabsTrigger value="movies" className="flex-1 rounded-lg data-active:bg-[var(--tv-yellow)] data-active:text-black">
              {t.home.movies}
            </TabsTrigger>
          </TabsList>

          <TabsContent value="tv" className="mt-4">
            {trendingShows.length > 0 ? (
              <div className="tv-scroll-row">
                {trendingShows.map((show) => (
                  <ShowCard key={show.tmdb_id} show={{ ...show, media_type: "tv" }} compact />
                ))}
              </div>
            ) : (
              <Card className="tv-card py-0 shadow-none">
                <CardContent className="p-4 text-sm text-muted-foreground">{t.home.trendingTvEmpty}</CardContent>
              </Card>
            )}
          </TabsContent>

          <TabsContent value="movies" className="mt-4">
            {trendingMovieItems.length > 0 ? (
              <div className="tv-scroll-row">
                {trendingMovieItems.map((movie) => (
                  <ShowCard key={movie.tmdb_id} show={{ ...movie, media_type: "movie" }} compact />
                ))}
              </div>
            ) : (
              <Card className="tv-card py-0 shadow-none">
                <CardContent className="p-4 text-sm text-muted-foreground">{t.home.trendingMoviesEmpty}</CardContent>
              </Card>
            )}
          </TabsContent>
        </Tabs>
      </section>
    </div>
  );
}
