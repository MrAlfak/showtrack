"use client";

import type { ListStatus } from "@/lib/api";
import { useLocale } from "@/components/locale/locale-provider";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Button } from "@/components/ui/button";
import { ChevronDown } from "lucide-react";

const statuses: ListStatus[] = ["watching", "plan_to_watch", "watched", "dropped", "archived"];

type Props = {
  value: ListStatus;
  onChange: (status: ListStatus) => void;
  disabled?: boolean;
};

export function ListStatusSelect({ value, onChange, disabled }: Props) {
  const { t } = useLocale();

  const labels: Record<ListStatus, string> = {
    watching: t.library.watching,
    plan_to_watch: t.library.planToWatch,
    watched: t.library.watchedList,
    dropped: t.library.dropped,
    archived: t.library.archived,
  };

  return (
    <DropdownMenu>
      <DropdownMenuTrigger
        render={
          <Button variant="outline" size="sm" className="gap-1 text-xs" disabled={disabled}>
            {labels[value]}
            <ChevronDown className="size-3" />
          </Button>
        }
      />
      <DropdownMenuContent align="end">
        {statuses.map((status) => (
          <DropdownMenuItem key={status} onClick={() => onChange(status)}>
            {labels[status]}
          </DropdownMenuItem>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
