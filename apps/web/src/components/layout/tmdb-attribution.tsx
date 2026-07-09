"use client";

import Image from "next/image";
import Link from "next/link";
import { useLocale } from "@/components/locale/locale-provider";

export function TmdbAttribution() {
  const { t } = useLocale();
  return (
    <div className="flex items-center justify-center gap-2 py-2 text-[10px] text-muted-foreground">
      <Image src="https://www.themoviedb.org/assets/2/v4/logos/v2/blue_short-8e7b30f73a402069b365aca80847ae0a8940c3b38f10306add558478cfb3177d.svg" alt="TMDB" width={48} height={12} unoptimized />
      <span>{t.extras.tmdbAttribution}</span>
      <Link href="https://www.themoviedb.org/" className="underline hover:text-foreground" target="_blank" rel="noreferrer">
        TMDB
      </Link>
    </div>
  );
}
