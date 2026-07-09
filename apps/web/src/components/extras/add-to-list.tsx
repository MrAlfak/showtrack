"use client";

import { useEffect, useState } from "react";
import { useLocale } from "@/components/locale/locale-provider";
import { addListItem, getMyLists, type CustomListSummary } from "@/lib/api";
import { Button } from "@/components/ui/button";

type AddToListButtonProps = {
  token: string;
  mediaType: "tv" | "movie";
  tmdbId: number;
  title: string;
  posterUrl: string;
};

export function AddToListButton({ token, mediaType, tmdbId, title, posterUrl }: AddToListButtonProps) {
  const { t } = useLocale();
  const [lists, setLists] = useState<CustomListSummary[]>([]);
  const [message, setMessage] = useState("");

  useEffect(() => {
    void getMyLists(token).then((data) => setLists(data?.lists ?? []));
  }, [token]);

  if (lists.length === 0) return null;

  return (
    <div className="flex flex-wrap items-center gap-2">
      <select
        className="h-9 rounded-md border border-input bg-background px-2 text-sm"
        defaultValue=""
        onChange={async (event) => {
          const listId = event.target.value;
          if (!listId) return;
          const ok = await addListItem(token, listId, {
            media_type: mediaType,
            tmdb_id: tmdbId,
            title,
            poster_url: posterUrl,
          });
          setMessage(ok?.ok ? t.extras.addedToList : t.extras.listAddFailed);
          event.target.value = "";
        }}
      >
        <option value="">{t.extras.addToList}</option>
        {lists.map((list) => (
          <option key={list.id} value={list.id}>
            {list.name}
          </option>
        ))}
      </select>
      {message ? <span className="text-xs text-muted-foreground">{message}</span> : null}
    </div>
  );
}
