"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { useAuth } from "@/components/auth/auth-provider";
import { useLocale } from "@/components/locale/locale-provider";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { getDashboard, trending, trendingMovies, type ShowItem, type UpcomingItem } from "@/lib/api";
import { ShowCard } from "@/components/show-card";
import { SectionHeader } from "@/components/section-header";

export default function HomePage() {
  const { token, ready, isAuthenticated } = useAuth();
  const { t, locale } = useLocale();
  const [trendingShows, setTrendingShows] = useState<ShowItem[]>([]);
  const [trendingMovieItems, setTrendingMovieItems] = useState<ShowItem[]>([]);
  const [library, setLibrary] = useState<ShowItem[]>([]);
  const [upcoming, setUpcoming] = useState<UpcomingItem[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    async function load() {
      const [trendingData, moviesData, dashboardData] = await Promise.all([
        trending(),
        trendingMovies(),
        token ? getDashboard(token) : Promise.resolve(null),
      ]);
      if (cancelled) return;
      setTrendingShows(trendingData?.results ?? []);
      setTrendingMovieItems(moviesData?.results ?? []);
      setLibrary(dashboardData?.library ?? []);
      setUpcoming(dashboardData?.upcoming ?? []);
      setLoading(false);
    }

    if (!ready) return;
    void load();
    return () => {
      cancelled = true;
    };
  }, [ready, token]);

  return (
    <div className="space-y-8">
      <header className="space-y-1">
        <p className="text-sm text-muted-foreground">{t.home.welcome}</p>
        <h1 className="text-2xl font-bold tracking-tight">{t.appName}</h1>
      </header>

      <section>
        <SectionHeader title={t.home.continueWatching} />
        {!ready || loading ? (
          <div className="grid grid-cols-2 gap-3 sm:grid-cols-3">
            {Array.from({ length: 2 }).map((_, index) => (
              <Skeleton key={index} className="aspect-[2/3] rounded-xl" />
            ))}
          </div>
        ) : library.length > 0 ? (
          <div className="grid grid-cols-2 gap-3 sm:grid-cols-3">
            {library.slice(0, 6).map((item) => (
              <ShowCard key={`${item.media_type ?? "tv"}-${item.id}`} show={item} />
            ))}
          </div>
        ) : (
          <Card className="border-border/60 bg-card/80 py-0 shadow-none">
            <CardContent className="p-4 text-sm text-muted-foreground">
              {isAuthenticated ? t.home.emptyLibrary : t.home.signInPrompt}
            </CardContent>
          </Card>
        )}
      </section>

      <section>
        <SectionHeader title={t.home.upcoming} />
        <div className="space-y-2">
          {(upcoming.length ? upcoming : []).map((item) => (
            <Card key={item.episode_id} className="border-border/60 bg-card/80 py-0 shadow-none">
              <CardContent className="flex items-center gap-3 p-3">
                <div className="size-12 shrink-0 overflow-hidden rounded-md bg-muted">
                  {/* eslint-disable-next-line @next/next/no-img-element */}
                  <img src={item.poster_url} alt={item.show_title} className="size-full object-cover" />
                </div>
                <div className="min-w-0 flex-1">
                  <p className="truncate text-sm font-medium">{item.show_title}</p>
                  <p className="text-xs text-muted-foreground">
                    S{item.season_number}E{item.episode_number} · {item.episode_name || t.home.nextEpisode}
                  </p>
                </div>
                <Badge variant="secondary">
                  {item.air_date
                    ? new Date(item.air_date).toLocaleDateString(locale === "fa" ? "fa-IR" : "en-US", {
                        month: "short",
                        day: "numeric",
                      })
                    : t.home.soon}
                </Badge>
              </CardContent>
            </Card>
          ))}
          {!loading && upcoming.length === 0 && (
            <Card className="border-border/60 bg-card/80 py-0 shadow-none">
              <CardContent className="p-4 text-sm text-muted-foreground">
                {t.home.upcomingEmpty}
              </CardContent>
            </Card>
          )}
        </div>
      </section>

      <section>
        <SectionHeader
          title={t.home.trending}
          action={
            <Link href="/discover" className="text-xs text-muted-foreground hover:text-foreground">
              {t.home.seeAll}
            </Link>
          }
        />
        <Tabs defaultValue="tv">
          <TabsList className="w-full">
            <TabsTrigger value="tv" className="flex-1">{t.home.tvShows}</TabsTrigger>
            <TabsTrigger value="movies" className="flex-1">{t.home.movies}</TabsTrigger>
          </TabsList>

          <TabsContent value="tv" className="mt-4">
            {trendingShows.length > 0 ? (
              <div className="flex gap-3 overflow-x-auto pb-1">
                {trendingShows.map((show) => (
                  <div key={show.tmdb_id} className="w-32">
                    <ShowCard show={{ ...show, media_type: "tv" }} compact />
                  </div>
                ))}
              </div>
            ) : (
              <Card className="border-border/60 bg-card/80 py-0 shadow-none">
                <CardContent className="p-4 text-sm text-muted-foreground">
                  {t.home.trendingTvEmpty}
                </CardContent>
              </Card>
            )}
          </TabsContent>

          <TabsContent value="movies" className="mt-4">
            {trendingMovieItems.length > 0 ? (
              <div className="flex gap-3 overflow-x-auto pb-1">
                {trendingMovieItems.map((movie) => (
                  <div key={movie.tmdb_id} className="w-32">
                    <ShowCard show={{ ...movie, media_type: "movie" }} compact />
                  </div>
                ))}
              </div>
            ) : (
              <Card className="border-border/60 bg-card/80 py-0 shadow-none">
                <CardContent className="p-4 text-sm text-muted-foreground">
                  {t.home.trendingMoviesEmpty}
                </CardContent>
              </Card>
            )}
          </TabsContent>
        </Tabs>
      </section>
    </div>
  );
}
