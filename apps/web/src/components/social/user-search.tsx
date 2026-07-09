"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import Image from "next/image";
import { Search, UserPlus, UserMinus } from "lucide-react";
import { useAuth } from "@/components/auth/auth-provider";
import { useLocale } from "@/components/locale/locale-provider";
import { followUser, searchUsers, unfollowUser, type UserSummary } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent } from "@/components/ui/card";

export function UserSearch() {
  const { token } = useAuth();
  const { t } = useLocale();
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<UserSummary[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!token || query.trim().length < 2) {
      setResults([]);
      return;
    }
    let cancelled = false;
    const timer = setTimeout(async () => {
      setLoading(true);
      const data = await searchUsers(token, query.trim());
      if (!cancelled) {
        setResults(data?.results ?? []);
        setLoading(false);
      }
    }, 300);
    return () => {
      cancelled = true;
      clearTimeout(timer);
    };
  }, [query, token]);

  if (!token) return null;

  return (
    <div className="space-y-3">
      <div className="relative">
        <Search className="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
        <Input
          className="pl-9"
          placeholder={t.social.searchUsers}
          value={query}
          onChange={(event) => setQuery(event.target.value)}
        />
      </div>
      {loading ? <p className="text-xs text-muted-foreground">{t.social.searching}</p> : null}
      <div className="space-y-2">
        {results.map((user) => (
          <UserRow key={user.id} user={user} onChange={setResults} />
        ))}
      </div>
    </div>
  );
}

function UserRow({
  user,
  onChange,
}: {
  user: UserSummary;
  onChange: React.Dispatch<React.SetStateAction<UserSummary[]>>;
}) {
  const { token } = useAuth();
  const { t } = useLocale();
  const [following, setFollowing] = useState(Boolean(user.is_following));
  const [busy, setBusy] = useState(false);
  const label = user.display_name || user.username || user.id.slice(0, 8);

  return (
    <Card className="tv-card shadow-none">
      <CardContent className="flex items-center gap-3 p-3">
        <Link href={`/users/${user.id}`} className="flex min-w-0 flex-1 items-center gap-3">
          {user.avatar_url ? (
            <Image src={user.avatar_url} alt={label} width={40} height={40} className="size-10 rounded-full object-cover" unoptimized />
          ) : (
            <div className="flex size-10 items-center justify-center rounded-full bg-secondary text-sm font-semibold">
              {label.slice(0, 1).toUpperCase()}
            </div>
          )}
          <div className="min-w-0">
            <p className="truncate text-sm font-medium">{label}</p>
            {user.username ? <p className="truncate text-xs text-muted-foreground">@{user.username}</p> : null}
          </div>
        </Link>
        <Button
          size="sm"
          variant={following ? "outline" : "default"}
          className="gap-1"
          disabled={busy || !token}
          onClick={async () => {
            if (!token) return;
            setBusy(true);
            if (following) {
              await unfollowUser(token, user.id);
              setFollowing(false);
              onChange((items) => items.map((item) => (item.id === user.id ? { ...item, is_following: false } : item)));
            } else {
              await followUser(token, user.id);
              setFollowing(true);
              onChange((items) => items.map((item) => (item.id === user.id ? { ...item, is_following: true } : item)));
            }
            setBusy(false);
          }}
        >
          {following ? <UserMinus className="size-3.5" /> : <UserPlus className="size-3.5" />}
          {following ? t.social.unfollow : t.social.follow}
        </Button>
      </CardContent>
    </Card>
  );
}
