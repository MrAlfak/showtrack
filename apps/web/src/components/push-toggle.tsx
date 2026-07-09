"use client";

import { useState } from "react";
import { Bell, BellOff } from "lucide-react";
import { useAuth } from "@/components/auth/auth-provider";
import { useLocale } from "@/components/locale/locale-provider";
import { Button } from "@/components/ui/button";
import { enableWebPush, isPushConfigured } from "@/lib/push";

export function PushToggle() {
  const { token, isAuthenticated } = useAuth();
  const { t } = useLocale();
  const [enabled, setEnabled] = useState(false);
  const [pending, setPending] = useState(false);
  const [message, setMessage] = useState("");

  if (!isAuthenticated) return null;

  const configured = isPushConfigured();

  return (
    <div className="space-y-2">
      <p className="text-xs text-muted-foreground">{t.profile.notificationsHint}</p>
      <Button
        variant={enabled ? "secondary" : "outline"}
        className="w-full justify-start gap-2"
        disabled={!token || pending || !configured}
        onClick={async () => {
          if (!token) return;
          setPending(true);
          setMessage("");
          const result = await enableWebPush(token);
          if (result.ok) {
            setEnabled(true);
            setMessage(t.profile.notificationsEnabled);
          } else if (result.reason === "denied") {
            setMessage(t.profile.notificationsDenied);
          } else if (result.reason === "unsupported") {
            setMessage(t.profile.notificationsUnsupported);
          } else {
            setMessage(t.profile.notificationsDenied);
          }
          setPending(false);
        }}
      >
        {enabled ? <Bell className="size-4" /> : <BellOff className="size-4" />}
        {pending ? "..." : enabled ? t.profile.notificationsEnabled : t.profile.enableNotifications}
      </Button>
      {message ? <p className="text-xs text-foreground">{message}</p> : null}
      {!configured ? <p className="text-xs text-muted-foreground">{t.profile.notificationsUnsupported}</p> : null}
    </div>
  );
}
