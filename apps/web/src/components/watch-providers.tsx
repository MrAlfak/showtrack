"use client";

import Image from "next/image";
import { useEffect, useState } from "react";
import { ExternalLink } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { useLocale } from "@/components/locale/locale-provider";
import { getWatchProviders, type WatchProviders } from "@/lib/api";

type Props = {
  mediaType: "tv" | "movie";
  tmdbId: string;
};

export function WatchProvidersSection({ mediaType, tmdbId }: Props) {
  const { t, locale } = useLocale();
  const [data, setData] = useState<WatchProviders | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    const country = locale === "fa" ? "IR" : "US";
    getWatchProviders(mediaType, tmdbId, country).then((providers) => {
      if (cancelled) return;
      setData(providers);
      setLoading(false);
    });
    return () => {
      cancelled = true;
    };
  }, [locale, mediaType, tmdbId]);

  if (loading) {
    return <Skeleton className="h-24 rounded-xl" />;
  }

  if (!data) {
    return null;
  }

  const hasProviders =
    (data.flatrate?.length ?? 0) + (data.rent?.length ?? 0) + (data.buy?.length ?? 0) > 0;

  if (!hasProviders) {
    return (
      <section className="space-y-2">
        <h2 className="text-base font-semibold">{t.detail.whereToWatch}</h2>
        <Card className="tv-card py-0 shadow-none">
          <CardContent className="p-4 text-sm text-muted-foreground">{t.detail.noProviders}</CardContent>
        </Card>
      </section>
    );
  }

  return (
    <section className="space-y-3">
      <div className="flex items-center justify-between gap-2">
        <h2 className="text-base font-semibold">{t.detail.whereToWatch}</h2>
        <Badge variant="outline" className="text-[10px] uppercase">
          {data.country}
        </Badge>
      </div>

      {data.flatrate?.length ? (
        <ProviderGroup title={t.detail.stream} providers={data.flatrate} />
      ) : null}
      {data.rent?.length ? <ProviderGroup title={t.detail.rent} providers={data.rent} /> : null}
      {data.buy?.length ? <ProviderGroup title={t.detail.buy} providers={data.buy} /> : null}

      {data.link ? (
        <a
          href={data.link}
          target="_blank"
          rel="noreferrer"
          className="inline-flex items-center gap-1 text-xs text-[var(--tv-yellow)] hover:underline"
        >
          {t.detail.viewOnTMDB}
          <ExternalLink className="size-3" />
        </a>
      ) : null}
    </section>
  );
}

function ProviderGroup({
  title,
  providers,
}: {
  title: string;
  providers: WatchProviders["flatrate"];
}) {
  return (
    <div className="space-y-2">
      <p className="text-xs text-muted-foreground">{title}</p>
      <div className="flex flex-wrap gap-2">
        {providers.map((provider) => (
          <div
            key={provider.id}
            className="flex items-center gap-2 rounded-lg border border-white/5 bg-card px-2 py-1.5"
          >
            <div className="relative size-6 overflow-hidden rounded">
              <Image src={provider.logo_url} alt={provider.name} fill className="object-cover" unoptimized />
            </div>
            <span className="text-xs font-medium">{provider.name}</span>
          </div>
        ))}
      </div>
    </div>
  );
}
