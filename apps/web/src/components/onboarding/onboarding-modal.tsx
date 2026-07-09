"use client";

import { useEffect, useState } from "react";
import { discover, submitOnboarding, trending, type ShowItem } from "@/lib/api";
import { useLocale } from "@/components/locale/locale-provider";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";

type OnboardingModalProps = {
  open: boolean;
  token: string;
  onDone: () => void;
};

export function OnboardingModal({ open, token, onDone }: OnboardingModalProps) {
  const { t } = useLocale();
  const [picks, setPicks] = useState<ShowItem[]>([]);
  const [selected, setSelected] = useState<Set<number>>(new Set());
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (!open) return;
    void Promise.all([trending(), discover("tv")]).then(([trendingData, discoverData]) => {
      const merged = [...(trendingData?.results ?? []), ...(discoverData?.results ?? [])];
      const seen = new Set<number>();
      const unique: ShowItem[] = [];
      for (const item of merged) {
        if (seen.has(item.tmdb_id)) continue;
        seen.add(item.tmdb_id);
        unique.push({ ...item, media_type: item.media_type ?? "tv" });
        if (unique.length >= 12) break;
      }
      setPicks(unique);
    });
  }, [open]);

  function toggle(id: number) {
    setSelected((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else if (next.size < 10) next.add(id);
      return next;
    });
  }

  return (
    <Dialog open={open} onOpenChange={(v) => !v && onDone()}>
      <DialogContent className="max-h-[85vh] overflow-y-auto sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{t.onboarding.title}</DialogTitle>
          <DialogDescription>{t.onboarding.subtitle}</DialogDescription>
        </DialogHeader>

        <div className="grid grid-cols-3 gap-2">
          {picks.map((show) => {
            const active = selected.has(show.tmdb_id);
            return (
              <button
                key={show.tmdb_id}
                type="button"
                onClick={() => toggle(show.tmdb_id)}
                className={`overflow-hidden rounded-lg border text-left transition ${
                  active ? "border-[var(--tv-yellow)] ring-2 ring-[var(--tv-yellow)]" : "border-border"
                }`}
              >
                {/* eslint-disable-next-line @next/next/no-img-element */}
                <img src={show.poster_url} alt={show.title} className="aspect-[2/3] w-full object-cover" />
                <p className="truncate p-1 text-[10px] font-medium">{show.title}</p>
              </button>
            );
          })}
        </div>

        <DialogFooter className="gap-2 sm:gap-0">
          <Button variant="ghost" onClick={onDone}>
            {t.onboarding.skip}
          </Button>
          <Button
            disabled={submitting || selected.size === 0}
            onClick={async () => {
              setSubmitting(true);
              const items = picks
                .filter((s) => selected.has(s.tmdb_id))
                .map((s) => ({ tmdb_id: s.tmdb_id, media_type: (s.media_type ?? "tv") as "tv" | "movie" }));
              await submitOnboarding(token, items);
              setSubmitting(false);
              onDone();
            }}
          >
            {t.onboarding.continue}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
