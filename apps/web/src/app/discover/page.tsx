"use client";

import { useEffect, useState } from "react";
import { trending, trendingMovies, type ShowItem } from "@/lib/api";
import { useLocale } from "@/components/locale/locale-provider";
import { ShowCard } from "@/components/show-card";
import { SectionHeader } from "@/components/section-header";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

const genres = ["Drama", "Sci-Fi", "Crime", "Fantasy", "Comedy", "Thriller"];

export default function DiscoverPage() {
  const { t } = useLocale();
  const [shows, setShows] = useState<ShowItem[]>([]);
  const [movies, setMovies] = useState<ShowItem[]>([]);

  useEffect(() => {
    void Promise.all([trending(), trendingMovies()]).then(([trendingData, moviesData]) => {
      setShows(trendingData?.results ?? []);
      setMovies(moviesData?.results ?? []);
    });
  }, []);

  return (
    <div className="space-y-8">
      <header className="space-y-1">
        <h1 className="text-2xl font-bold tracking-tight">{t.discover.title}</h1>
        <p className="text-sm text-muted-foreground">{t.discover.subtitle}</p>
      </header>

      <section>
        <SectionHeader title={t.discover.genres} />
        <div className="flex flex-wrap gap-2">
          {genres.map((genre) => (
            <Badge key={genre} variant="secondary" className="cursor-pointer px-3 py-1.5 text-sm">
              {genre}
            </Badge>
          ))}
        </div>
      </section>

      <section>
        <SectionHeader title={t.discover.trendingWeek} />
        <Tabs defaultValue="tv">
          <TabsList className="w-full">
            <TabsTrigger value="tv" className="flex-1">{t.home.tvShows}</TabsTrigger>
            <TabsTrigger value="movies" className="flex-1">{t.home.movies}</TabsTrigger>
          </TabsList>

          <TabsContent value="tv" className="mt-4">
            {shows.length > 0 ? (
              <div className="grid grid-cols-2 gap-3 sm:grid-cols-3">
                {shows.map((show) => (
                  <ShowCard key={show.tmdb_id} show={{ ...show, media_type: "tv" }} />
                ))}
              </div>
            ) : (
              <Card className="border-border/60 bg-card/80 py-0 shadow-none">
                <CardContent className="p-4 text-sm text-muted-foreground">{t.discover.tvEmpty}</CardContent>
              </Card>
            )}
          </TabsContent>

          <TabsContent value="movies" className="mt-4">
            {movies.length > 0 ? (
              <div className="grid grid-cols-2 gap-3 sm:grid-cols-3">
                {movies.map((movie) => (
                  <ShowCard key={movie.tmdb_id} show={{ ...movie, media_type: "movie" }} />
                ))}
              </div>
            ) : (
              <Card className="border-border/60 bg-card/80 py-0 shadow-none">
                <CardContent className="p-4 text-sm text-muted-foreground">{t.discover.moviesEmpty}</CardContent>
              </Card>
            )}
          </TabsContent>
        </Tabs>
      </section>
    </div>
  );
}
