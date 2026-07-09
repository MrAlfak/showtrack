"use client";

import { useEffect, useState } from "react";
import { Link2, RefreshCw, Unlink } from "lucide-react";
import { useAuth } from "@/components/auth/auth-provider";
import { useLocale } from "@/components/locale/locale-provider";
import { disconnectTrakt, getTraktStatus, startTraktConnect, syncTrakt } from "@/lib/api";
import { Button } from "@/components/ui/button";

type TraktConnectProps = {
  onSynced?: () => void;
};

export function TraktConnect({ onSynced }: TraktConnectProps) {
  const { token } = useAuth();
  const { t, format } = useLocale();
  const [connected, setConnected] = useState(false);
  const [username, setUsername] = useState("");
  const [loading, setLoading] = useState(true);
  const [syncing, setSyncing] = useState(false);
  const [message, setMessage] = useState("");

  useEffect(() => {
    if (!token) return;
    let cancelled = false;
    async function load() {
      const status = await getTraktStatus(token!);
      if (cancelled) return;
      setConnected(Boolean(status?.connected));
      setUsername(status?.username ?? "");
      setLoading(false);
    }
    void load();
    return () => {
      cancelled = true;
    };
  }, [token]);

  if (!token || loading) return null;

  return (
    <div className="space-y-3">
      <p className="text-xs text-muted-foreground">{t.profile.traktHint}</p>
      {connected ? (
        <div className="space-y-2">
          <p className="text-sm">
            {t.profile.traktConnected}: <span className="font-medium">{username}</span>
          </p>
          <div className="flex flex-wrap gap-2">
            <Button
              size="sm"
              variant="outline"
              className="gap-2"
              disabled={syncing}
              onClick={async () => {
                setSyncing(true);
                setMessage("");
                const result = await syncTrakt(token);
                if (result?.ok) {
                  setMessage(format(t.profile.traktSyncOk, { imported: result.imported, skipped: result.skipped }));
                  onSynced?.();
                } else {
                  setMessage(t.profile.traktSyncFailed);
                }
                setSyncing(false);
              }}
            >
              <RefreshCw className={`size-4 ${syncing ? "animate-spin" : ""}`} />
              {syncing ? t.profile.traktSyncing : t.profile.traktSync}
            </Button>
            <Button
              size="sm"
              variant="ghost"
              className="gap-2"
              onClick={async () => {
                await disconnectTrakt(token);
                setConnected(false);
                setUsername("");
                setMessage(t.profile.traktDisconnected);
              }}
            >
              <Unlink className="size-4" />
              {t.profile.traktDisconnect}
            </Button>
          </div>
        </div>
      ) : (
        <Button
          size="sm"
          className="gap-2"
          onClick={async () => {
            const result = await startTraktConnect(token);
            if (result?.url) {
              window.location.href = result.url;
            } else {
              setMessage(t.profile.traktUnavailable);
            }
          }}
        >
          <Link2 className="size-4" />
          {t.profile.traktConnect}
        </Button>
      )}
      {message ? <p className="text-xs text-foreground">{message}</p> : null}
    </div>
  );
}
