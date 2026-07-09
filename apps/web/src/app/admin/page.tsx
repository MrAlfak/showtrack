"use client";

import { useEffect, useState } from "react";
import { useAuth } from "@/components/auth/auth-provider";
import { useLocale } from "@/components/locale/locale-provider";
import { getAdminStats, getAdminUsers } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";

const ADMIN_KEY = "showtrack.adminToken";

export default function AdminPage() {
  const { t } = useLocale();
  const { isAuthenticated } = useAuth();
  const [token, setToken] = useState("");
  const [stats, setStats] = useState<Record<string, number> | null>(null);
  const [users, setUsers] = useState<Array<Record<string, string>>>([]);
  const [error, setError] = useState("");

  useEffect(() => {
    const saved = sessionStorage.getItem(ADMIN_KEY);
    if (saved) setToken(saved);
  }, []);

  async function load(adminToken: string) {
    setError("");
    const [statsData, usersData] = await Promise.all([
      getAdminStats(adminToken),
      getAdminUsers(adminToken),
    ]);
    if (!statsData) {
      setError(t.extras.adminDenied);
      setStats(null);
      setUsers([]);
      return;
    }
    setStats(statsData);
    setUsers(usersData?.users ?? []);
  }

  return (
    <div className="space-y-4">
      <header>
        <h1 className="text-2xl font-bold">{t.extras.adminTitle}</h1>
        <p className="text-sm text-muted-foreground">{t.extras.adminSubtitle}</p>
      </header>

      <Card className="tv-card shadow-none">
        <CardContent className="flex gap-2 p-4">
          <Input
            type="password"
            placeholder={t.extras.adminToken}
            value={token}
            onChange={(event) => setToken(event.target.value)}
          />
          <Button
            onClick={() => {
              sessionStorage.setItem(ADMIN_KEY, token);
              void load(token);
            }}
          >
            {t.extras.adminConnect}
          </Button>
        </CardContent>
      </Card>

      {error ? <p className="text-sm text-destructive">{error}</p> : null}

      {stats ? (
        <div className="grid grid-cols-2 gap-2">
          {Object.entries(stats).map(([key, value]) => (
            <Card key={key} className="tv-card shadow-none">
              <CardHeader className="pb-2">
                <CardTitle className="text-xs capitalize text-muted-foreground">{key}</CardTitle>
              </CardHeader>
              <CardContent className="pt-0 text-2xl font-semibold">{value}</CardContent>
            </Card>
          ))}
        </div>
      ) : null}

      {users.length > 0 ? (
        <Card className="tv-card shadow-none">
          <CardHeader>
            <CardTitle className="text-base">{t.extras.adminUsers}</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2 text-sm">
            {users.map((user) => (
              <div key={user.id} className="rounded-lg border border-border/60 px-3 py-2">
                <p className="font-medium">{user.display_name || user.username || user.email}</p>
                <p className="text-xs text-muted-foreground">{user.email}</p>
              </div>
            ))}
          </CardContent>
        </Card>
      ) : null}

      {!isAuthenticated ? (
        <p className="text-xs text-muted-foreground">{t.extras.adminNote}</p>
      ) : null}
    </div>
  );
}
