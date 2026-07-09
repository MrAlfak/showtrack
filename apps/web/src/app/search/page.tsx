"use client";

import { useEffect, useState } from "react";
import Image from "next/image";
import Link from "next/link";
import { Check, Plus, Search as SearchIcon } from "lucide-react";
import { useAuth } from "@/components/auth/auth-provider";
import { useLocale } from "@/components/locale/locale-provider";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { addShow, addMovie, search, type SearchItem } from "@/lib/api";

export default function SearchPage() {
  const { token, isAuthenticated } = useAuth();
  const { t } = useLocale();
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<SearchItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [savingId, setSavingId] = useState<number | null>(null);
  const [addedIds, setAddedIds] = useState<number[]>([]);

  useEffect(() => {
    if (query.length < 2) {
      return;
    }
    let cancelled = false;
    const timer = setTimeout(async () => {
      setLoading(true);
      const data = await search(query);
      if (cancelled) return;
      setResults(data?.results ?? []);
      setLoading(false);
    }, 300);
    return () => {
      cancelled = true;
      clearTimeout(timer);
    };
  }, [query]);

  const shows = results.filter((r) => r.media_type === "tv");
  const movies = results.filter((r) => r.media_type === "movie");
  const people = results.filter((r) => r.media_type === "person");

  async function handleAdd(item: SearchItem) {
    if (!token) return;
    setSavingId(item.id);
    const response =
      item.media_type === "movie"
        ? await addMovie(item.id, token)
        : await addShow(item.id, token);
    if (response?.ok) {
      setAddedIds((current) => [...new Set([...current, item.id])]);
    }
    setSavingId(null);
  }

  return (
    <div className="space-y-6">
      <header className="space-y-1">
        <h1 className="text-2xl font-bold tracking-tight">{t.search.title}</h1>
        <p className="text-sm text-muted-foreground">{t.search.subtitle}</p>
      </header>

      <div className="relative">
        <SearchIcon className="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
        <Input
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder={t.search.placeholder}
          className="h-11 pl-10"
        />
      </div>

      {loading ? (
        <div className="space-y-3">
          {Array.from({ length: 4 }).map((_, i) => (
            <Skeleton key={i} className="h-16 w-full rounded-xl" />
          ))}
        </div>
      ) : query.length < 2 ? (
        <p className="text-sm text-muted-foreground">{t.search.minChars}</p>
      ) : (
        <Tabs defaultValue="all">
          <TabsList className="w-full">
            <TabsTrigger value="all" className="flex-1">{t.search.all}</TabsTrigger>
            <TabsTrigger value="shows" className="flex-1">{t.search.tv}</TabsTrigger>
            <TabsTrigger value="movies" className="flex-1">{t.search.movies}</TabsTrigger>
            <TabsTrigger value="people" className="flex-1">{t.search.people}</TabsTrigger>
          </TabsList>

          <TabsContent value="all" className="mt-4 space-y-2">
            {results.map((item) => (
              <SearchResult
                key={`${item.media_type}-${item.id}`}
                item={item}
                canAdd={isAuthenticated && (item.media_type === "tv" || item.media_type === "movie")}
                added={addedIds.includes(item.id)}
                saving={savingId === item.id}
                onAdd={() => handleAdd(item)}
              />
            ))}
          </TabsContent>
          <TabsContent value="shows" className="mt-4 space-y-2">
            {shows.map((item) => (
              <SearchResult
                key={item.id}
                item={item}
                canAdd={isAuthenticated}
                added={addedIds.includes(item.id)}
                saving={savingId === item.id}
                onAdd={() => handleAdd(item)}
              />
            ))}
          </TabsContent>
          <TabsContent value="movies" className="mt-4 space-y-2">
            {movies.map((item) => (
              <SearchResult
                key={item.id}
                item={item}
                canAdd={isAuthenticated}
                added={addedIds.includes(item.id)}
                saving={savingId === item.id}
                onAdd={() => handleAdd(item)}
              />
            ))}
          </TabsContent>
          <TabsContent value="people" className="mt-4 space-y-2">
            {people.map((item) => (
              <SearchResult key={item.id} item={item} />
            ))}
          </TabsContent>
        </Tabs>
      )}
    </div>
  );
}

function SearchResult({
  item,
  canAdd,
  added,
  saving,
  onAdd,
}: {
  item: SearchItem;
  canAdd?: boolean;
  added?: boolean;
  saving?: boolean;
  onAdd?: () => void;
}) {
  const href =
    item.media_type === "person"
      ? `/persons/${item.id}`
      : item.media_type === "movie"
        ? `/movies/${item.id}`
        : `/shows/${item.id}`;

  return (
    <Card className="border-border/60 bg-card/80 py-0 shadow-none transition-colors hover:border-border">
      <CardContent className="flex items-center gap-3 p-3">
        <Link href={href} className="contents">
          <div className="relative size-12 shrink-0 overflow-hidden rounded-md bg-muted">
            <Image src={item.poster_url} alt={item.title} fill className="object-cover" sizes="48px" unoptimized />
          </div>
          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-2">
              <p className="truncate text-sm font-medium">{item.title}</p>
              <Badge variant="outline" className="text-[10px] capitalize">
                {item.media_type}
              </Badge>
            </div>
            {item.overview && (
              <p className="line-clamp-1 text-xs text-muted-foreground">{item.overview}</p>
            )}
          </div>
        </Link>
        {canAdd && (
          <Button
            size="sm"
            variant={added ? "secondary" : "outline"}
            disabled={added || saving}
            onClick={onAdd}
            className="shrink-0"
          >
            {added ? <Check className="size-4" /> : <Plus className="size-4" />}
          </Button>
        )}
      </CardContent>
    </Card>
  );
}
