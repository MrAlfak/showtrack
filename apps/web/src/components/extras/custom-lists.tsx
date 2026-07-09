"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { ListPlus, Trash2 } from "lucide-react";
import { useAuth } from "@/components/auth/auth-provider";
import { useLocale } from "@/components/locale/locale-provider";
import { createList, deleteList, getMyLists, type CustomListSummary } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";

export function CustomListsSection() {
  const { token } = useAuth();
  const { t } = useLocale();
  const [lists, setLists] = useState<CustomListSummary[]>([]);
  const [name, setName] = useState("");
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    if (!token) return;
    void getMyLists(token).then((data) => setLists(data?.lists ?? []));
  }, [token]);

  if (!token) return null;

  return (
    <Card className="tv-card shadow-none">
      <CardHeader>
        <CardTitle className="text-base">{t.extras.customLists}</CardTitle>
      </CardHeader>
      <CardContent className="space-y-3">
        <p className="text-xs text-muted-foreground">{t.extras.customListsHint}</p>
        <div className="flex gap-2">
          <Input placeholder={t.extras.listName} value={name} onChange={(e) => setName(e.target.value)} />
          <Button
            size="sm"
            className="shrink-0 gap-1"
            disabled={busy || !name.trim()}
            onClick={async () => {
              setBusy(true);
              const result = await createList(token, name.trim());
              if (result?.ok) {
                const data = await getMyLists(token);
                setLists(data?.lists ?? []);
                setName("");
              }
              setBusy(false);
            }}
          >
            <ListPlus className="size-4" />
            {t.extras.createList}
          </Button>
        </div>
        <div className="space-y-2">
          {lists.map((list) => (
            <div key={list.id} className="flex items-center justify-between rounded-lg border border-border/60 px-3 py-2 text-sm">
              <Link href={`/lists/${list.id}`} className="min-w-0 flex-1 hover:underline">
                <p className="font-medium">{list.name}</p>
                <p className="text-xs text-muted-foreground">
                  {list.item_count} {t.extras.items}
                </p>
              </Link>
              <Button
                size="icon"
                variant="ghost"
                className="size-8"
                onClick={async () => {
                  await deleteList(token, list.id);
                  setLists((items) => items.filter((item) => item.id !== list.id));
                }}
              >
                <Trash2 className="size-4" />
              </Button>
            </div>
          ))}
          {lists.length === 0 ? <p className="text-xs text-muted-foreground">{t.extras.noLists}</p> : null}
        </div>
      </CardContent>
    </Card>
  );
}
