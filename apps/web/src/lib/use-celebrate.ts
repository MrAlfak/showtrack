"use client";

import { useCallback, useState } from "react";
import { cn } from "@/lib/utils";

export function useCelebrate() {
  const [active, setActive] = useState(false);

  const celebrate = useCallback(() => {
    setActive(true);
    window.setTimeout(() => setActive(false), 500);
  }, []);

  return { active, celebrate, className: cn(active && "tv-celebrate") };
}
