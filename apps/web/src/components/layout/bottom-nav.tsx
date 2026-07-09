"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { Clapperboard, Compass, Search, User, Users } from "lucide-react";
import { useLocale } from "@/components/locale/locale-provider";
import { cn } from "@/lib/utils";
import { APP_MAX_WIDTH } from "@/lib/layout";

export function BottomNav() {
  const pathname = usePathname();
  const { t } = useLocale();

  const links = [
    { href: "/", label: t.nav.home, icon: Clapperboard },
    { href: "/feed", label: t.nav.feed, icon: Users },
    { href: "/discover", label: t.nav.discover, icon: Compass },
    { href: "/search", label: t.nav.search, icon: Search },
    { href: "/profile", label: t.nav.profile, icon: User },
  ];

  return (
    <nav
      className="fixed bottom-0 left-1/2 z-50 -translate-x-1/2 border-t border-border/80 bg-[#121212]/95 backdrop-blur-xl"
      style={{ width: "100%", maxWidth: APP_MAX_WIDTH }}
    >
      <div className="flex items-center justify-around px-1 py-1.5 pb-[max(0.375rem,env(safe-area-inset-bottom))]">
        {links.map(({ href, label, icon: Icon }) => {
          const active = pathname === href || (href !== "/" && pathname.startsWith(href));
          return (
            <Link
              key={href}
              href={href}
              className={cn(
                "flex min-w-[4.5rem] flex-col items-center gap-0.5 rounded-xl px-2 py-1.5 text-[10px] font-medium transition-colors",
                active ? "text-[var(--tv-yellow)]" : "text-muted-foreground hover:text-foreground"
              )}
            >
              <Icon className={cn("size-6", active && "fill-[var(--tv-yellow)]/15")} strokeWidth={active ? 2.25 : 1.75} />
              <span>{label}</span>
            </Link>
          );
        })}
      </div>
    </nav>
  );
}
