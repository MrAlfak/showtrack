import Image from "next/image";
import Link from "next/link";
import { Card, CardContent } from "@/components/ui/card";
import type { CastMember } from "@/lib/api";

export function CastRow({ cast }: { cast: CastMember[] }) {
  return (
    <div className="flex gap-3 overflow-x-auto pb-2">
      {cast.map((person) => (
        <Link key={person.tmdb_id} href={`/persons/${person.tmdb_id}`} className="group shrink-0">
          <Card className="w-24 border-border/60 bg-card/80 py-0 shadow-none">
            <CardContent className="flex flex-col items-center gap-2 p-2">
              <div className="relative size-16 overflow-hidden rounded-full border border-border/60">
                <Image
                  src={person.profile_url}
                  alt={person.name}
                  fill
                  className="object-cover"
                  sizes="64px"
                  unoptimized
                />
              </div>
              <div className="text-center">
                <p className="line-clamp-1 text-xs font-medium">{person.name}</p>
                <p className="line-clamp-1 text-[10px] text-muted-foreground">{person.character}</p>
              </div>
            </CardContent>
          </Card>
        </Link>
      ))}
    </div>
  );
}
