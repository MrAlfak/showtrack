"use client";

import { useLocale } from "@/components/locale/locale-provider";

export function AppHeader() {
  const { t } = useLocale();

  return (
    <header className="sticky top-0 z-40 -mx-4 mb-4 border-b border-border/60 bg-background/90 px-4 py-3 backdrop-blur-md">
      <div className="flex items-center gap-2.5">
        <div className="flex size-9 items-center justify-center rounded-xl bg-[var(--tv-yellow)] shadow-[0_0_20px_rgba(255,214,10,0.25)]">
          <span className="text-lg font-black text-black">S</span>
        </div>
        <div>
          <h1 className="text-lg font-bold leading-none tracking-tight">{t.appName}</h1>
          <p className="text-[11px] text-muted-foreground">{t.home.welcome}</p>
        </div>
      </div>
    </header>
  );
}
