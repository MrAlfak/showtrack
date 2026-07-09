"use client";

import { useEffect, useState } from "react";
import { discover, getGenres, type Genre, type ShowItem } from "@/lib/api";
import { useLocale } from "@/components/locale/locale-provider";
import { ShowCard } from "@/components/show-card";
import { SectionHeader } from "@/components/section-header";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

export default function DiscoverPage() {
  const { t } = useLocale();
  const [mediaTab, setMediaTab] = useState<"tv" | "movie">("tv");
  const [genres, setGenres] = useState<Genre[]>([]);
  const [selectedGenre, setSelectedGenre] = useState<number | null>(null);
  const [items, setItems] = useState<ShowItem[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    void getGenres(mediaTab).then((data) => {
      setGenres(data?.genres ?? []);
    });
  }, [mediaTab]);

  useEffect(() => {
    let cancelled = false;
    async function load() {
      setLoading(true);
      const data = await discover(mediaTab, selectedGenre ?? undefined);
      if (!cancelled) {
        setItems(data?.results ?? []);
        setLoading(false);
      }
    }
    void load();
    return () => {
      cancelled = true;
    };
  }, [mediaTab, selectedGenre]);

  return (
    <div className="space-y-8">
      <header className="space-y-1">
        <h1 className="text-2xl font-bold tracking-tight">{t.discover.title}</h1>
        <p className="text-sm text-muted-foreground">{t.discover.subtitle}</p>
      </header>

      <Tabs
        value={mediaTab}
        onValueChange={(v) => {
          setMediaTab(v as "tv" | "movie");
          setSelectedGenre(null);
        }}
      >
        <TabsList className="h-9 w-full rounded-xl bg-secondary/80 p-1">
          <TabsTrigger value="tv" className="flex-1 rounded-lg data-active:bg-[var(--tv-yellow)] data-active:text-black">
            {t.home.tvShows}
          </TabsTrigger>
          <TabsTrigger value="movies" className="flex-1 rounded-lg data-active:bg-[var(--tv-yellow)] data-active:text-black">
            {t.home.movies}
          </TabsTrigger>
        </TabsList>

        <TabsContent value={mediaTab} className="mt-6 space-y-6">
          <section>
            <SectionHeader title={t.discover.genres} />
            <div className="flex flex-wrap gap-2">
              <Badge
                variant={selectedGenre === null ? "default" : "secondary"}
                className="cursor-pointer px-3 py-1.5 text-sm"
                onClick={() => setSelectedGenre(null)}
              >
                {t.discover.allGenres}
              </Badge>
              {genres.map((genre) => (
                <Badge
                  key={genre.id}
                  variant={selectedGenre === genre.id ? "default" : "secondary"}
                  className="cursor-pointer px-3 py-1.5 text-sm"
                  onClick={() => setSelectedGenre(genre.id)}
                >
                  {genre.name}
                </Badge>
              ))}
            </div>
          </section>

          <section>
            <SectionHeader title={t.discover.trendingWeek} />
            {loading ? (
              <div className="tv-scroll-row">
                {Array.from({ length: 4 }).map((_, i) => (
                  <Skeleton key={i} className="aspect-[2/3] w-[7.25rem] shrink-0 rounded-xl" />
                ))}
              </div>
            ) : items.length > 0 ? (
              <div className="tv-scroll-row">
                {items.map((item) => (
                  <ShowCard key={item.tmdb_id} show={{ ...item, media_type: mediaTab }} compact />
                ))}
              </div>
            ) : (
              <Card className="tv-card py-0 shadow-none">
                <CardContent className="p-4 text-sm text-muted-foreground">
                  {mediaTab === "tv" ? t.discover.tvEmpty : t.discover.moviesEmpty}
                </CardContent>
              </Card>
            )}
          </section>
        </TabsContent>
      </Tabs>
    </div>
  );
}
