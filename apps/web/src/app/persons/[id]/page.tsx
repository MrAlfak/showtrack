"use client";

import Image from "next/image";
import Link from "next/link";
import { use, useEffect, useState } from "react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { useLocale } from "@/components/locale/locale-provider";
import { getPerson, type Credit, type PersonDetail } from "@/lib/api";

export default function PersonPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params);
  const { t } = useLocale();
  const [person, setPerson] = useState<PersonDetail | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    getPerson(id).then((data) => {
      if (cancelled) return;
      setPerson(data);
      setLoading(false);
    });
    return () => {
      cancelled = true;
    };
  }, [id]);

  if (loading) {
    return (
      <div className="space-y-4">
        <div className="flex flex-col items-center gap-4">
          <Skeleton className="size-28 rounded-full" />
          <Skeleton className="h-6 w-36 rounded-md" />
        </div>
        <Skeleton className="h-40 rounded-xl" />
      </div>
    );
  }

  if (!person) {
    return <div className="text-sm text-muted-foreground">{t.detail.notFoundPerson}</div>;
  }

  const tv = person.credits.filter((c) => c.media_type === "tv");
  const movies = person.credits.filter((c) => c.media_type === "movie");

  return (
    <div className="space-y-6">
      <div className="flex flex-col items-center gap-4 text-center">
        <div className="relative size-28 overflow-hidden rounded-full border-2 border-border/60">
          <Image src={person.profile_url} alt={person.name} fill className="object-cover" unoptimized />
        </div>
        <div className="space-y-1">
          <h1 className="text-2xl font-bold tracking-tight">{person.name}</h1>
          <Badge variant="secondary">{t.detail.actor}</Badge>
        </div>
      </div>

      {person.biography && (
        <section className="space-y-2">
          <h2 className="text-base font-semibold">{t.detail.biography}</h2>
          <p className="line-clamp-6 text-sm leading-relaxed text-muted-foreground">{person.biography}</p>
        </section>
      )}

      <Tabs defaultValue="all">
        <TabsList className="w-full">
          <TabsTrigger value="all" className="flex-1">
            {t.detail.allCredits} ({person.credits.length})
          </TabsTrigger>
          <TabsTrigger value="tv" className="flex-1">
            {t.home.tvShows} ({tv.length})
          </TabsTrigger>
          <TabsTrigger value="movies" className="flex-1">
            {t.home.movies} ({movies.length})
          </TabsTrigger>
        </TabsList>

        <TabsContent value="all" className="mt-4">
          <CreditGrid credits={person.credits} />
        </TabsContent>
        <TabsContent value="tv" className="mt-4">
          <CreditGrid credits={tv} />
        </TabsContent>
        <TabsContent value="movies" className="mt-4">
          <CreditGrid credits={movies} />
        </TabsContent>
      </Tabs>
    </div>
  );
}

function CreditGrid({ credits }: { credits: Credit[] }) {
  const { format, t } = useLocale();

  return (
    <div className="grid grid-cols-2 gap-3 sm:grid-cols-3">
      {credits.map((credit) => (
        <Link
          key={`${credit.media_type}-${credit.tmdb_id}`}
          href={credit.media_type === "movie" ? `/movies/${credit.tmdb_id}` : `/shows/${credit.tmdb_id}`}
        >
          <Card className="overflow-hidden border-border/60 bg-card/80 py-0 shadow-none">
            <CardContent className="p-0">
              <div className="relative aspect-[2/3] w-full">
                <Image src={credit.poster_url} alt={credit.title} fill className="object-cover" unoptimized />
              </div>
              <div className="space-y-1 p-2">
                <p className="line-clamp-1 text-xs font-medium">{credit.title}</p>
                <p className="line-clamp-1 text-[10px] text-muted-foreground">
                  {format(t.detail.asCharacter, { character: credit.character })}
                </p>
              </div>
            </CardContent>
          </Card>
        </Link>
      ))}
    </div>
  );
}
