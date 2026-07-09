"use client";

import { useEffect, useMemo, useState } from "react";
import { Clock, Download, Film, Flame, Globe, LogIn, LogOut, PlayCircle, Tv, Upload } from "lucide-react";
import { useAuth } from "@/components/auth/auth-provider";
import { useLocale } from "@/components/locale/locale-provider";
import { PushToggle } from "@/components/push-toggle";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Separator } from "@/components/ui/separator";
import { Progress } from "@/components/ui/progress";
import { exportWatchHistory, getDashboard, importWatchHistory, login, register, type DashboardStats, type ShowItem } from "@/lib/api";

export default function ProfilePage() {
  const { token, userId, isAuthenticated, ready, signIn, signOut } = useAuth();
  const { t, format, locale, setLocale } = useLocale();
  const [mode, setMode] = useState<"login" | "register">("login");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [displayName, setDisplayName] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [library, setLibrary] = useState<ShowItem[]>([]);
  const [importing, setImporting] = useState(false);
  const [message, setMessage] = useState<string>("");
  const [stats, setStats] = useState<DashboardStats>({
    shows: 0,
    movies: 0,
    episodes: 0,
    total: 0,
    hours: 0,
    streak: 0,
  });

  useEffect(() => {
    if (!token) return;
    let cancelled = false;
    getDashboard(token).then((data) => {
      if (!cancelled && data) {
        setLibrary(data.library ?? []);
        setStats(data.stats);
      }
    });
    return () => {
      cancelled = true;
    };
  }, [token]);

  const visibleLibrary = useMemo(() => (token ? library : []), [library, token]);
  const visibleStats = useMemo(
    () => (token ? stats : { shows: 0, movies: 0, episodes: 0, total: 0, hours: 0, streak: 0 }),
    [stats, token]
  );

  const statCards = [
    { label: t.profile.shows, value: visibleStats.shows, icon: Tv },
    { label: t.profile.movies, value: visibleStats.movies, icon: Film },
    { label: t.profile.watched, value: visibleStats.episodes, icon: PlayCircle },
    { label: t.profile.hours, value: visibleStats.hours, icon: Clock },
    { label: t.profile.streak, value: `${visibleStats.streak}d`, icon: Flame },
  ];

  async function refreshDashboard(activeToken: string) {
    const data = await getDashboard(activeToken);
    if (!data) return;
    setLibrary(data.library ?? []);
    setStats(data.stats);
  }

  return (
    <div className="space-y-8">
      <header className="space-y-1">
        <h1 className="text-2xl font-bold tracking-tight">{t.profile.title}</h1>
        <p className="text-sm text-muted-foreground">
          {isAuthenticated ? t.profile.statsSubtitle : t.profile.signInSubtitle}
        </p>
      </header>

      <Card className="border-border/60 bg-card/80 shadow-none">
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
        <Card className="border-border/60 bg-card/80 shadow-none">
          <CardHeader>
            <CardTitle className="text-base">{mode === "login" ? t.profile.signIn : t.profile.register}</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            {mode === "register" && (
              <Input
                placeholder={t.profile.displayName}
                value={displayName}
                onChange={(event) => setDisplayName(event.target.value)}
              />
            )}
            <Input placeholder={t.profile.email} value={email} onChange={(event) => setEmail(event.target.value)} />
            <Input
              type="password"
              placeholder={t.profile.password}
              value={password}
              onChange={(event) => setPassword(event.target.value)}
            />
            <Button
              className="w-full"
              disabled={submitting || !email || !password}
              onClick={async () => {
                setSubmitting(true);
                const response =
                  mode === "login"
                    ? await login({ email, password })
                    : await register({ email, password, display_name: displayName });
                if (response?.token && response?.user_id) {
                  signIn(response.token, response.user_id);
                  await refreshDashboard(response.token);
                  setEmail("");
                  setPassword("");
                  setDisplayName("");
                  setMessage(t.profile.signedIn);
                }
                setSubmitting(false);
              }}
            >
              {mode === "login" ? <LogIn className="mr-2 size-4" /> : null}
              {mode === "login" ? t.profile.signIn : t.profile.register}
            </Button>
            <Button variant="ghost" className="w-full" onClick={() => setMode(mode === "login" ? "register" : "login")}>
              {mode === "login" ? t.profile.needAccount : t.profile.haveAccount}
            </Button>
          </CardContent>
        </Card>
      )}

      {isAuthenticated && (
        <Card className="border-border/60 bg-card/80 shadow-none">
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
          <Card key={label} className="border-border/60 bg-card/80 shadow-none">
            <CardContent className="flex items-center gap-3 p-4">
              <div className="flex size-10 items-center justify-center rounded-lg bg-secondary">
                <Icon className="size-5 text-violet-400" />
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
        <h2 className="text-base font-semibold">{t.profile.watchProgress}</h2>
        {visibleLibrary.map((item) => (
          <Card key={`${item.media_type ?? "tv"}-${item.id}`} className="border-border/60 bg-card/80 shadow-none">
            <CardContent className="space-y-2 p-4">
              <div className="flex items-center justify-between text-sm">
                <span className="font-medium">{item.title}</span>
                <span className="text-muted-foreground">
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
          <Card className="border-border/60 bg-card/80 shadow-none">
            <CardContent className="p-4 text-sm text-muted-foreground">{t.profile.emptyLibrary}</CardContent>
          </Card>
        )}
      </section>

      {isAuthenticated && (
        <Card className="border-border/60 bg-card/80 shadow-none">
          <CardHeader>
            <CardTitle className="text-base">{t.profile.notifications}</CardTitle>
          </CardHeader>
          <CardContent>
            <PushToggle />
          </CardContent>
        </Card>
      )}

      <Separator />

      <Card className="border-border/60 bg-card/80 shadow-none">
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
