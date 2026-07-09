"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { Clapperboard, Compass, Search, User } from "lucide-react";
import { useLocale } from "@/components/locale/locale-provider";
import { cn } from "@/lib/utils";

export function BottomNav() {
  const pathname = usePathname();
  const { t } = useLocale();

  const links = [
    { href: "/", label: t.nav.home, icon: Clapperboard },
    { href: "/discover", label: t.nav.discover, icon: Compass },
    { href: "/search", label: t.nav.search, icon: Search },
    { href: "/profile", label: t.nav.profile, icon: User },
  ];

  return (
    <nav className="fixed inset-x-0 bottom-0 z-50 border-t border-border/60 bg-background/90 backdrop-blur-xl">
      <div className="mx-auto flex max-w-lg items-center justify-around px-2 py-2">
        {links.map(({ href, label, icon: Icon }) => {
          const active = pathname === href || (href !== "/" && pathname.startsWith(href));
          return (
            <Link
              key={href}
              href={href}
              className={cn(
                "flex min-w-16 flex-col items-center gap-1 rounded-xl px-3 py-2 text-xs transition-colors",
                active ? "text-foreground" : "text-muted-foreground hover:text-foreground"
              )}
            >
              <Icon className={cn("size-5", active && "text-violet-400")} />
              <span>{label}</span>
            </Link>
          );
        })}
      </div>
    </nav>
  );
}
