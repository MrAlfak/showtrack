"use client";

import { useEffect, useState } from "react";
import { Users } from "lucide-react";
import { useAuth } from "@/components/auth/auth-provider";
import { useLocale } from "@/components/locale/locale-provider";
import { ActivityCard } from "@/components/social/activity-card";
import { UserSearch } from "@/components/social/user-search";
import { SignInPrompt } from "@/components/auth/auth-form";
import { EmptyState } from "@/components/ui/empty-state";
import { getFeed, type ActivityItem } from "@/lib/api";

export default function FeedPage() {
  const { token, ready, isAuthenticated } = useAuth();
  const { t } = useLocale();
  const [items, setItems] = useState<ActivityItem[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!ready) return;
    if (!token) {
      setLoading(false);
      return;
    }
    let cancelled = false;
    async function load() {
      const data = await getFeed(token!);
      if (!cancelled) {
        setItems(data?.items ?? []);
        setLoading(false);
      }
    }
    void load();
    return () => {
      cancelled = true;
    };
  }, [ready, token]);

  return (
    <div className="space-y-6">
      <header className="space-y-1">
        <h1 className="text-2xl font-bold tracking-tight">{t.social.feedTitle}</h1>
        <p className="text-sm text-muted-foreground">{t.social.feedSubtitle}</p>
      </header>

      {!isAuthenticated && ready ? <SignInPrompt /> : null}

      {isAuthenticated ? (
        <>
          <section className="space-y-2">
            <h2 className="text-base font-semibold">{t.social.findFriends}</h2>
            <UserSearch />
          </section>

          <section className="space-y-2">
            <h2 className="text-base font-semibold">{t.social.activity}</h2>
            {loading ? (
              <p className="text-sm text-muted-foreground">{t.social.loadingFeed}</p>
            ) : items.length > 0 ? (
              <div className="space-y-2">
                {items.map((item) => (
                  <ActivityCard key={item.id} item={item} />
                ))}
              </div>
            ) : (
              <EmptyState
                icon={Users}
                title={t.social.feedEmpty}
                description={t.social.feedEmptyHint}
              />
            )}
          </section>
        </>
      ) : null}
    </div>
  );
}
