"use client";

import { useEffect, useState } from "react";
import Image from "next/image";
import Link from "next/link";
import { useParams } from "next/navigation";
import { UserMinus, UserPlus } from "lucide-react";
import { useAuth } from "@/components/auth/auth-provider";
import { useLocale } from "@/components/locale/locale-provider";
import { followUser, getUserProfile, unfollowUser, type UserProfile } from "@/lib/api";
import { ShowCard } from "@/components/show-card";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";

export default function UserProfilePage() {
  const params = useParams<{ id: string }>();
  const { token, userId, ready } = useAuth();
  const { t } = useLocale();
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    if (!token || !ready || !params.id) return;
    let cancelled = false;
    async function load() {
      const data = await getUserProfile(token!, params.id);
      if (cancelled) return;
      if (!data) {
        setError(t.social.profileUnavailable);
        return;
      }
      setProfile(data);
    }
    void load();
    return () => {
      cancelled = true;
    };
  }, [token, ready, params.id, t.social.profileUnavailable]);

  if (!ready || !token) {
    return <p className="text-sm text-muted-foreground">{t.auth.signInRequired}</p>;
  }

  if (error) {
    return <p className="text-sm text-destructive">{error}</p>;
  }

  if (!profile) {
    return <p className="text-sm text-muted-foreground">{t.social.loadingProfile}</p>;
  }

  const isSelf = profile.id === userId;
  const label = profile.display_name || profile.username || profile.id.slice(0, 8);

  return (
    <div className="space-y-6">
      <Card className="tv-card shadow-none">
        <CardContent className="space-y-4 p-4">
          <div className="flex items-start gap-4">
            {profile.avatar_url ? (
              <Image src={profile.avatar_url} alt={label} width={72} height={72} className="size-[4.5rem] rounded-full object-cover" unoptimized />
            ) : (
              <div className="flex size-[4.5rem] items-center justify-center rounded-full bg-secondary text-2xl font-semibold">
                {label.slice(0, 1).toUpperCase()}
              </div>
            )}
            <div className="min-w-0 flex-1 space-y-1">
              <h1 className="text-xl font-bold">{label}</h1>
              {profile.username ? <p className="text-sm text-muted-foreground">@{profile.username}</p> : null}
              {profile.bio ? <p className="text-sm">{profile.bio}</p> : null}
              <p className="text-xs text-muted-foreground">
                {profile.followers} {t.social.followers} · {profile.following} {t.social.followingCount}
              </p>
            </div>
          </div>

          {!isSelf ? (
            <Button
              className="gap-2"
              variant={profile.is_following ? "outline" : "default"}
              disabled={busy}
              onClick={async () => {
                setBusy(true);
                if (profile.is_following) {
                  await unfollowUser(token, profile.id);
                  setProfile({ ...profile, is_following: false, followers: Math.max(0, profile.followers - 1) });
                } else {
                  await followUser(token, profile.id);
                  setProfile({ ...profile, is_following: true, followers: profile.followers + 1 });
                }
                setBusy(false);
              }}
            >
              {profile.is_following ? <UserMinus className="size-4" /> : <UserPlus className="size-4" />}
              {profile.is_following ? t.social.unfollow : t.social.follow}
            </Button>
          ) : (
            <Link href="/profile">
              <Button variant="outline">{t.social.editProfile}</Button>
            </Link>
          )}
        </CardContent>
      </Card>

      <div className="grid grid-cols-3 gap-2 text-center">
        <Card className="tv-card shadow-none">
          <CardContent className="p-3">
            <p className="text-lg font-semibold">{profile.stats.shows}</p>
            <p className="text-xs text-muted-foreground">{t.profile.shows}</p>
          </CardContent>
        </Card>
        <Card className="tv-card shadow-none">
          <CardContent className="p-3">
            <p className="text-lg font-semibold">{profile.stats.movies}</p>
            <p className="text-xs text-muted-foreground">{t.profile.movies}</p>
          </CardContent>
        </Card>
        <Card className="tv-card shadow-none">
          <CardContent className="p-3">
            <p className="text-lg font-semibold">{profile.stats.episodes}</p>
            <p className="text-xs text-muted-foreground">{t.profile.watched}</p>
          </CardContent>
        </Card>
      </div>

      {profile.library_preview && profile.library_preview.length > 0 ? (
        <section className="space-y-3">
          <h2 className="text-base font-semibold">{t.social.libraryPreview}</h2>
          <div className="tv-scroll-row">
            {profile.library_preview.map((item) => (
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
        </section>
      ) : null}
    </div>
  );
}
