"use client";

import { useState } from "react";
import { Star } from "lucide-react";
import { useLocale } from "@/components/locale/locale-provider";
import { deleteMyRating, setMyRating } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { cn } from "@/lib/utils";

type UserRatingProps = {
  token: string;
  mediaType: "tv" | "movie";
  tmdbId: number;
  initialScore?: number;
  initialReview?: string;
  onChange?: (score: number, review: string) => void;
};

export function UserRatingSection({
  token,
  mediaType,
  tmdbId,
  initialScore = 0,
  initialReview = "",
  onChange,
}: UserRatingProps) {
  const { t } = useLocale();
  const [score, setScore] = useState(initialScore);
  const [review, setReview] = useState(initialReview);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState("");

  return (
    <Card className="tv-card shadow-none">
      <CardHeader>
        <CardTitle className="text-base">{t.extras.myRating}</CardTitle>
      </CardHeader>
      <CardContent className="space-y-3">
        <p className="text-xs text-muted-foreground">{t.extras.ratingPrivate}</p>
        <div className="flex flex-wrap gap-1">
          {Array.from({ length: 10 }).map((_, index) => {
            const value = index + 1;
            const active = value <= score;
            return (
              <button
                key={value}
                type="button"
                className={cn(
                  "flex size-8 items-center justify-center rounded-md border text-xs font-semibold transition-colors",
                  active
                    ? "border-[var(--tv-yellow)] bg-[var(--tv-yellow)]/15 text-[var(--tv-yellow)]"
                    : "border-border text-muted-foreground hover:border-foreground/30"
                )}
                onClick={() => setScore(value)}
              >
                {value}
              </button>
            );
          })}
        </div>
        <textarea
          className="flex min-h-20 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          placeholder={t.extras.reviewPlaceholder}
          value={review}
          onChange={(event) => setReview(event.target.value)}
          rows={3}
        />
        <div className="flex gap-2">
          <Button
            size="sm"
            disabled={saving || score < 1}
            onClick={async () => {
              setSaving(true);
              setMessage("");
              const result = await setMyRating(token, mediaType, tmdbId, score, review);
              if (result?.ok) {
                setMessage(t.extras.ratingSaved);
                onChange?.(score, review);
              } else {
                setMessage(t.extras.ratingFailed);
              }
              setSaving(false);
            }}
          >
            <Star className="mr-1 size-4" />
            {saving ? t.extras.saving : t.extras.saveRating}
          </Button>
          {score > 0 ? (
            <Button
              size="sm"
              variant="ghost"
              disabled={saving}
              onClick={async () => {
                setSaving(true);
                await deleteMyRating(token, mediaType, tmdbId);
                setScore(0);
                setReview("");
                setMessage(t.extras.ratingRemoved);
                setSaving(false);
              }}
            >
              {t.extras.removeRating}
            </Button>
          ) : null}
        </div>
        {message ? <p className="text-xs text-foreground">{message}</p> : null}
      </CardContent>
    </Card>
  );
}
