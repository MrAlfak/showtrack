"use client";

import Link from "next/link";
import { WifiOff } from "lucide-react";
import { useLocale } from "@/components/locale/locale-provider";
import { EmptyState } from "@/components/ui/empty-state";

export default function OfflinePage() {
  const { t } = useLocale();

  return (
    <div className="tv-fade-in flex min-h-[60vh] items-center">
      <EmptyState
        icon={WifiOff}
        title={t.offline.title}
        description={t.offline.description}
        action={
          <Link
            href="/"
            className="inline-flex h-8 items-center justify-center rounded-lg bg-primary px-3 text-sm font-medium text-primary-foreground hover:bg-primary/80"
          >
            {t.offline.retry}
          </Link>
        }
        className="w-full"
      />
    </div>
  );
}
