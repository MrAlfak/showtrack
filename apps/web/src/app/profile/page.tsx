"use client";

import { useEffect, useMemo, useState } from "react";
import { Clock, Download, Film, Flame, Globe, LogOut, PlayCircle, Popcorn, Tv, Upload, Users } from "lucide-react";
import Link from "next/link";
import { useAuth } from "@/components/auth/auth-provider";
import { AuthForm } from "@/components/auth/auth-form";
import { useLocale } from "@/components/locale/locale-provider";
import { PushToggle } from "@/components/push-toggle";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Separator } from "@/components/ui/separator";
import { Progress } from "@/components/ui/progress";
import { exportWatchHistory, getDashboard, getLibrary, getMyProfile, importWatchHistory, updateMovieStatus, updateMyProfile, updateShowStatus, type DashboardStats, type ListStatus, type ShowItem, type UserProfile } from "@/lib/api";
import { ListStatusSelect } from "@/components/library/list-status-select";
import { TraktConnect } from "@/components/trakt-connect";
import { CustomListsSection } from "@/components/extras/custom-lists";

export default function ProfilePage() {
  const { token, userId, isAuthenticated, ready, signOut } = useAuth();
  const { t, format, locale, setLocale } = useLocale();
  const [library, setLibrary] = useState<ShowItem[]>([]);
  const [libraryTab, setLibraryTab] = useState<ListStatus | "all">("watching");
  const [importing, setImporting] = useState(false);
  const [message, setMessage] = useState<string>("");
  const [socialProfile, setSocialProfile] = useState<UserProfile | null>(null);
  const [username, setUsername] = useState("");
  const [bio, setBio] = useState("");
  const [isPublic, setIsPublic] = useState(true);
  const [savingProfile, setSavingProfile] = useState(false);
  const [stats, setStats] = useState<DashboardStats>({
    shows: 0,
    movies: 0,
    episodes: 0,
    total: 0,
    hours: 0,
    streak: 0,
    binge_today: 0,
  });

  useEffect(() => {
    if (!token || !ready) return;
    let cancelled = false;
    async function loadProfile() {
      const profile = await getMyProfile(token!);
      if (cancelled || !profile) return;
      setSocialProfile(profile);
      setUsername(profile.username ?? "");
      setBio(profile.bio ?? "");
      setIsPublic(profile.is_public);
    }
    void loadProfile();
    return () => {
      cancelled = true;
    };
  }, [token, ready]);

  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    if (params.get("trakt") === "connected") {
      setMessage(t.profile.traktLinked);
    }
  }, [t.profile.traktLinked]);

  useEffect(() => {
    if (!token || !ready) return;
    let cancelled = false;
    async function load() {
      const [dashboard, lib] = await Promise.all([
        getDashboard(token!),
        getLibrary(token!, libraryTab === "all" ? undefined : libraryTab),
      ]);
      if (cancelled) return;
      if (dashboard) setStats(dashboard.stats);
      if (lib) {
        setLibrary([...(lib.shows ?? []), ...(lib.movies ?? [])]);
      }
    }
    void load();
    return () => {
      cancelled = true;
    };
  }, [token, ready, libraryTab]);

  const visibleLibrary = useMemo(() => (token ? library : []), [library, token]);
  const visibleStats = useMemo(
    () => (token ? stats : { shows: 0, movies: 0, episodes: 0, total: 0, hours: 0, streak: 0, binge_today: 0 }),
    [stats, token]
  );

  const statCards = [
    { label: t.profile.shows, value: visibleStats.shows, icon: Tv },
    { label: t.profile.movies, value: visibleStats.movies, icon: Film },
    { label: t.profile.watched, value: visibleStats.episodes, icon: PlayCircle },
    { label: t.profile.hours, value: visibleStats.hours, icon: Clock },
    { label: t.profile.streak, value: `${visibleStats.streak}d`, icon: Flame },
    { label: t.extras.bingeToday, value: visibleStats.binge_today ?? 0, icon: Popcorn },
  ];

  async function refreshDashboard(activeToken: string) {
    const [dashboard, lib] = await Promise.all([
      getDashboard(activeToken),
      getLibrary(activeToken, libraryTab === "all" ? undefined : libraryTab),
    ]);
    if (dashboard) setStats(dashboard.stats);
    if (lib) setLibrary([...(lib.shows ?? []), ...(lib.movies ?? [])]);
  }

  async function changeItemStatus(item: ShowItem, status: ListStatus) {
    if (!token) return;
    if (item.media_type === "movie") {
      await updateMovieStatus(item.id, status, token);
    } else {
      await updateShowStatus(item.id, status, token);
    }
    await refreshDashboard(token);
  }

  return (
    <div className="space-y-8">
      <header className="space-y-1">
        <h1 className="text-2xl font-bold tracking-tight">{t.profile.title}</h1>
        <p className="text-sm text-muted-foreground">
          {isAuthenticated ? t.profile.statsSubtitle : t.profile.signInSubtitle}
        </p>
      </header>

      <Card className="tv-card shadow-none">
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-base">
            <Globe className="size-4" />
            {t.profile.language}
          </CardTitle>
        </CardHeader>
        <CardContent className="flex gap-2">
          <Button variant={locale === "en" ? "default" : "outline"} size="sm" onClick={() => setLocale("en")}>
            {t.profile.english}
          </Button>
          <Button variant={locale === "fa" ? "default" : "outline"} size="sm" onClick={() => setLocale("fa")}>
            {t.profile.persian}
          </Button>
        </CardContent>
      </Card>

      {!isAuthenticated && ready && (
        <AuthForm onSuccess={(activeToken) => void refreshDashboard(activeToken)} />
      )}

      {isAuthenticated && (
        <Card className="tv-card shadow-none">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-base">
              <Users className="size-4" />
              {t.social.publicProfile}
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <p className="text-xs text-muted-foreground">{t.social.publicProfileHint}</p>
            <Input
              placeholder={t.social.username}
              value={username}
              onChange={(event) => setUsername(event.target.value)}
              autoComplete="username"
            />
            <p className="text-[10px] text-muted-foreground">{t.social.usernameHint}</p>
            <Input
              placeholder={t.social.bio}
              value={bio}
              onChange={(event) => setBio(event.target.value)}
            />
            <label className="flex items-center gap-2 text-sm">
              <input
                type="checkbox"
                checked={isPublic}
                onChange={(event) => setIsPublic(event.target.checked)}
              />
              {t.social.publicProfile}
            </label>
            <Button
              size="sm"
              disabled={savingProfile || !token}
              onClick={async () => {
                if (!token) return;
                setSavingProfile(true);
                const updated = await updateMyProfile(token, {
                  username: username.trim() || undefined,
                  bio,
                  is_public: isPublic,
                });
                if (updated) {
                  setSocialProfile(updated);
                  setMessage(t.social.profileSaved);
                } else {
                  setMessage(t.social.profileSaveFailed);
                }
                setSavingProfile(false);
              }}
            >
              {savingProfile ? t.profile.importing : t.social.saveProfile}
            </Button>
            {socialProfile?.username ? (
              <Link href={`/users/${socialProfile.id}`} className="text-xs text-[var(--tv-yellow)] hover:underline">
                @{socialProfile.username}
              </Link>
            ) : null}
          </CardContent>
        </Card>
      )}

      {isAuthenticated && (
        <CustomListsSection />
      )}

      {isAuthenticated && (
        <Card className="tv-card shadow-none">
          <CardContent className="flex items-center justify-between p-4">
            <div>
              <p className="text-sm font-medium">{t.profile.connected}</p>
              <p className="text-xs text-muted-foreground">{userId}</p>
            </div>
            <Button variant="outline" size="sm" className="gap-2" onClick={signOut}>
              <LogOut className="size-4" />
              {t.profile.signOut}
            </Button>
          </CardContent>
        </Card>
      )}

      <div className="grid grid-cols-2 gap-3">
        {statCards.map(({ label, value, icon: Icon }) => (
          <Card key={label} className="tv-card shadow-none">
            <CardContent className="flex items-center gap-3 p-4">
              <div className="flex size-10 items-center justify-center rounded-lg bg-secondary">
                <Icon className="size-5 text-[var(--tv-yellow)]" />
              </div>
              <div>
                <p className="text-xs text-muted-foreground">{label}</p>
                <p className="text-lg font-semibold">{value}</p>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      <section className="space-y-3">
        <h2 className="text-base font-semibold">{t.library.title}</h2>
        <div className="flex flex-wrap gap-1">
          {(["watching", "plan_to_watch", "watched", "dropped", "archived"] as const).map((tab) => (
            <Button
              key={tab}
              size="sm"
              variant={libraryTab === tab ? "default" : "outline"}
              className="text-xs"
              onClick={() => setLibraryTab(tab)}
            >
              {tab === "watching"
                ? t.library.watching
                : tab === "plan_to_watch"
                  ? t.library.planToWatch
                  : tab === "watched"
                    ? t.library.watchedList
                    : tab === "dropped"
                      ? t.library.dropped
                      : t.library.archived}
            </Button>
          ))}
        </div>
        <div className="space-y-2">
        {visibleLibrary.map((item) => (
          <Card key={`${item.media_type ?? "tv"}-${item.id}`} className="tv-card shadow-none">
            <CardContent className="space-y-2 p-4">
              <div className="flex items-center justify-between gap-2 text-sm">
                <Link
                  href={item.media_type === "movie" ? `/movies/${item.tmdb_id}` : `/shows/${item.tmdb_id}`}
                  className="min-w-0 flex-1 truncate font-medium hover:underline"
                >
                  {item.title}
                </Link>
                <ListStatusSelect
                  value={item.list_status ?? "watching"}
                  onChange={(status) => void changeItemStatus(item, status)}
                />
              </div>
              <div className="flex items-center justify-between text-xs text-muted-foreground">
                <span>
                  {item.media_type === "movie"
                    ? item.progress && item.progress >= 100
                      ? t.profile.watchedLabel
                      : t.profile.notWatched
                    : `${item.watched}/${item.total}`}
                </span>
              </div>
              <Progress value={item.progress ?? 0} className="h-1.5" />
            </CardContent>
          </Card>
        ))}
        {isAuthenticated && visibleLibrary.length === 0 && (
          <Card className="tv-card shadow-none">
            <CardContent className="p-4 text-sm text-muted-foreground">{t.library.empty}</CardContent>
          </Card>
        )}
        </div>
      </section>

      {isAuthenticated && (
        <Card className="tv-card shadow-none">
          <CardHeader>
            <CardTitle className="text-base">{t.profile.notifications}</CardTitle>
          </CardHeader>
          <CardContent>
            <PushToggle />
          </CardContent>
        </Card>
      )}

      <Separator />

      {isAuthenticated && (
        <Card className="tv-card shadow-none">
          <CardHeader>
            <CardTitle className="text-base">{t.profile.trakt}</CardTitle>
          </CardHeader>
          <CardContent>
            <TraktConnect onSynced={() => token && void refreshDashboard(token)} />
          </CardContent>
        </Card>
      )}

      <Card className="tv-card shadow-none">
        <CardHeader>
          <CardTitle className="text-base">{t.profile.data}</CardTitle>
        </CardHeader>
        <CardContent className="space-y-2 text-sm text-muted-foreground">
          <label className="block">
            <Input
              type="file"
              accept="application/json"
              disabled={!token || importing}
              onChange={async (event) => {
                const file = event.target.files?.[0];
                if (!file || !token) return;
                setImporting(true);
                setMessage("");
                try {
                  const text = await file.text();
                  const payload = JSON.parse(text);
                  const result = await importWatchHistory(token, payload);
                  if (result?.ok) {
                    await refreshDashboard(token);
                    setMessage(format(t.profile.importOk, { imported: result.imported, skipped: result.skipped }));
                  } else {
                    setMessage(t.profile.importFailed);
                  }
                } catch {
                  setMessage(t.profile.invalidJson);
                } finally {
                  event.currentTarget.value = "";
                  setImporting(false);
                }
              }}
            />
          </label>
          <p className="text-xs">{t.profile.importHint}</p>
          <Button
            variant="outline"
            className="w-full justify-start gap-2"
            disabled={!token}
            onClick={async () => {
              if (!token) return;
              const payload = await exportWatchHistory(token);
              if (!payload) return;
              const blob = new Blob([JSON.stringify(payload, null, 2)], { type: "application/json" });
              const url = URL.createObjectURL(blob);
              const anchor = document.createElement("a");
              anchor.href = url;
              anchor.download = "showtrack-watch-history.json";
              anchor.click();
              URL.revokeObjectURL(url);
            }}
          >
            <Download className="size-4" />
            {t.profile.exportHistory}
          </Button>
          <div className="flex items-center gap-2 text-xs">
            <Upload className="size-3.5" />
            <span>{importing ? t.profile.importing : t.profile.importLabel}</span>
          </div>
          {message ? <p className="text-xs text-foreground">{message}</p> : null}
        </CardContent>
      </Card>

      <p className="text-center text-[10px] text-muted-foreground">{t.profile.footer}</p>
    </div>
  );
}
