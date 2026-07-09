"use client";

import { useEffect, useState } from "react";
import { Download, X } from "lucide-react";
import { useLocale } from "@/components/locale/locale-provider";
import { Button } from "@/components/ui/button";

type BeforeInstallPromptEvent = Event & {
  prompt: () => Promise<void>;
  userChoice: Promise<{ outcome: "accepted" | "dismissed" }>;
};

export function InstallPrompt() {
  const { t } = useLocale();
  const [deferred, setDeferred] = useState<BeforeInstallPromptEvent | null>(null);
  const [dismissed, setDismissed] = useState(false);

  useEffect(() => {
    if (window.matchMedia("(display-mode: standalone)").matches) return;

    const dismissedKey = "showtrack.pwa.dismissed";
    if (window.localStorage.getItem(dismissedKey) === "1") {
      setDismissed(true);
      return;
    }

    function onBeforeInstall(e: Event) {
      e.preventDefault();
      setDeferred(e as BeforeInstallPromptEvent);
    }

    window.addEventListener("beforeinstallprompt", onBeforeInstall);
    return () => window.removeEventListener("beforeinstallprompt", onBeforeInstall);
  }, []);

  if (dismissed || !deferred) return null;

  return (
    <div className="mb-4 flex items-start gap-3 rounded-2xl border border-[var(--tv-yellow)]/30 bg-[var(--tv-yellow)]/10 p-3">
      <Download className="mt-0.5 size-5 shrink-0 text-[var(--tv-yellow)]" />
      <div className="min-w-0 flex-1">
        <p className="text-sm font-semibold">{t.pwa.installTitle}</p>
        <p className="text-xs text-muted-foreground">{t.pwa.installHint}</p>
        <Button
          size="sm"
          className="mt-2 h-8"
          onClick={async () => {
            await deferred.prompt();
            const choice = await deferred.userChoice;
            if (choice.outcome === "accepted") {
              setDeferred(null);
            }
          }}
        >
          {t.pwa.installAction}
        </Button>
      </div>
      <button
        type="button"
        className="text-muted-foreground hover:text-foreground"
        aria-label={t.pwa.dismiss}
        onClick={() => {
          window.localStorage.setItem("showtrack.pwa.dismissed", "1");
          setDismissed(true);
          setDeferred(null);
        }}
      >
        <X className="size-4" />
      </button>
    </div>
  );
}
