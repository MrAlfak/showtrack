"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import Image from "next/image";
import { useParams } from "next/navigation";
import { useAuth } from "@/components/auth/auth-provider";
import { useLocale } from "@/components/locale/locale-provider";
import { getList, type CustomListItem } from "@/lib/api";
import { ShowCard } from "@/components/show-card";

export default function ListDetailPage() {
  const params = useParams<{ id: string }>();
  const { token, ready } = useAuth();
  const { t } = useLocale();
  const [name, setName] = useState("");
  const [items, setItems] = useState<CustomListItem[]>([]);

  useEffect(() => {
    if (!token || !ready || !params.id) return;
    void getList(token, params.id).then((data) => {
      if (!data) return;
      setName(data.list.name);
      setItems(data.items);
    });
  }, [token, ready, params.id]);

  if (!ready || !token) return <p className="text-sm text-muted-foreground">{t.auth.signInRequired}</p>;

  return (
    <div className="space-y-4">
      <header>
        <h1 className="text-2xl font-bold">{name || t.extras.customLists}</h1>
      </header>
      {items.length > 0 ? (
        <div className="tv-scroll-row">
          {items.map((item) => (
            <ShowCard
              key={`${item.media_type}-${item.tmdb_id}`}
              show={{
                id: item.tmdb_id,
                tmdb_id: item.tmdb_id,
                title: item.title,
                poster_url: item.poster_url,
                media_type: item.media_type,
              }}
              compact
            />
          ))}
        </div>
      ) : (
        <p className="text-sm text-muted-foreground">{t.extras.emptyList}</p>
      )}
      <Link href="/profile" className="text-sm text-[var(--tv-yellow)] hover:underline">
        {t.extras.backToProfile}
      </Link>
    </div>
  );
}
