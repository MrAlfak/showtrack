export const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1";

export type SearchItem = {
  id: number;
  title: string;
  overview?: string;
  media_type: "tv" | "movie" | "person";
  poster_url: string;
  vote_average?: number;
};

export type ShowItem = {
  id: number;
  tmdb_id: number;
  title: string;
  overview?: string;
  status?: string;
  media_type?: "tv" | "movie";
  vote_average?: number;
  poster_url: string;
  progress?: number;
  watched?: number;
  total?: number;
};

export type CastMember = {
  tmdb_id: number;
  name: string;
  character: string;
  profile_url: string;
};

export type Season = {
  id: number;
  season_number: number;
  name: string;
  episode_count: number;
  episodes: Episode[];
};

export type Episode = {
  id: number;
  episode_number: number;
  name: string;
  overview?: string;
  air_date?: string;
  runtime?: number;
  still_url: string;
  watched: boolean;
};

export type ShowDetail = ShowItem & {
  cast: CastMember[];
  seasons: Season[];
  in_library?: boolean;
};

export type MovieDetail = {
  id: number;
  tmdb_id: number;
  title: string;
  overview?: string;
  runtime?: number;
  release_date?: string;
  vote_average?: number;
  poster_url: string;
  cast: CastMember[];
  in_library?: boolean;
  watched?: boolean;
};

export type Credit = {
  media_type: string;
  tmdb_id: number;
  title: string;
  character: string;
  poster_url: string;
  vote_average?: number;
};

export type PersonDetail = {
  id: number;
  tmdb_id: number;
  name: string;
  biography?: string;
  profile_url: string;
  credits: Credit[];
};

export type LibraryResponse = {
  shows: ShowItem[];
  movies: ShowItem[];
};

export type DashboardStats = {
  shows: number;
  movies: number;
  episodes: number;
  total: number;
  hours: number;
  streak: number;
};

export type UpcomingItem = {
  show_title: string;
  show_tmdb_id: number;
  episode_id: number;
  season_number: number;
  episode_number: number;
  episode_name: string;
  air_date?: string;
  poster_url: string;
};

export type DashboardResponse = {
  stats: DashboardStats;
  library: ShowItem[];
  upcoming: UpcomingItem[];
};

export type AuthResponse = {
  token: string;
  user_id: string;
};

async function fetchAPI<T>(path: string, init?: RequestInit): Promise<T | null> {
  try {
    const res = await fetch(`${API_URL}${path}`, {
      next: { revalidate: 60 },
      ...init,
      headers: {
        "Content-Type": "application/json",
        ...(init?.headers ?? {}),
      },
    });
    if (!res.ok) return null;
    return res.json();
  } catch {
    return null;
  }
}

function authHeaders(token?: string) {
  return token ? ({ Authorization: `Bearer ${token}` } satisfies Record<string, string>) : undefined;
}

export function search(query: string) {
  return fetchAPI<{ results: SearchItem[] }>(`/search?q=${encodeURIComponent(query)}`);
}

export function trending() {
  return fetchAPI<{ results: ShowItem[] }>("/trending");
}

export function trendingMovies() {
  return fetchAPI<{ results: ShowItem[] }>("/trending/movies");
}

export function getShow(id: string, token?: string) {
  return fetchAPI<ShowDetail>(`/shows/${id}`, { headers: authHeaders(token) });
}

export type WatchProvider = {
  id: number;
  name: string;
  logo_url: string;
};

export type WatchProviders = {
  country: string;
  link: string;
  flatrate: WatchProvider[];
  rent: WatchProvider[];
  buy: WatchProvider[];
};

export function getWatchProviders(mediaType: "tv" | "movie", id: string, country = "US") {
  const path =
    mediaType === "movie"
      ? `/movies/${id}/watch?country=${encodeURIComponent(country)}`
      : `/shows/${id}/watch?country=${encodeURIComponent(country)}`;
  return fetchAPI<WatchProviders>(path);
}

export function getMovie(id: string, token?: string) {
  return fetchAPI<MovieDetail>(`/movies/${id}`, { headers: authHeaders(token) });
}

export function getPerson(id: string) {
  return fetchAPI<PersonDetail>(`/persons/${id}`);
}

export function getLibrary(token: string) {
  return fetchAPI<LibraryResponse>("/me/library", { headers: authHeaders(token) });
}

export function getDashboard(token: string) {
  return fetchAPI<DashboardResponse>("/me/dashboard", { headers: authHeaders(token) });
}

export function register(payload: { email: string; password: string; display_name?: string }) {
  return fetchAPI<AuthResponse>("/auth/register", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export function login(payload: { email: string; password: string }) {
  return fetchAPI<AuthResponse>("/auth/login", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export function addShow(tmdbId: number, token: string) {
  return fetchAPI<{ ok: boolean; show_id: number }>("/shows", {
    method: "POST",
    headers: authHeaders(token),
    body: JSON.stringify({ tmdb_id: tmdbId }),
  });
}

export function removeShow(showId: number, token: string) {
  return fetchAPI<{ ok: boolean }>(`/shows/${showId}`, {
    method: "DELETE",
    headers: authHeaders(token),
  });
}

export function addMovie(tmdbId: number, token: string) {
  return fetchAPI<{ ok: boolean; movie_id: number }>("/movies", {
    method: "POST",
    headers: authHeaders(token),
    body: JSON.stringify({ tmdb_id: tmdbId }),
  });
}

export function removeMovie(movieId: number, token: string) {
  return fetchAPI<{ ok: boolean }>(`/movies/${movieId}`, {
    method: "DELETE",
    headers: authHeaders(token),
  });
}

export function markMovieWatched(movieId: number, token: string) {
  return fetchAPI<{ ok: boolean }>(`/movies/${movieId}/watched`, {
    method: "POST",
    headers: authHeaders(token),
  });
}

export function unmarkMovieWatched(movieId: number, token: string) {
  return fetchAPI<{ ok: boolean }>(`/movies/${movieId}/watched`, {
    method: "DELETE",
    headers: authHeaders(token),
  });
}

export function markWatched(episodeId: number, token: string) {
  return fetchAPI<{ ok: boolean }>(`/episodes/${episodeId}/watched`, {
    method: "POST",
    headers: authHeaders(token),
  });
}

export function unmarkWatched(episodeId: number, token: string) {
  return fetchAPI<{ ok: boolean }>(`/episodes/${episodeId}/watched`, {
    method: "DELETE",
    headers: authHeaders(token),
  });
}

export async function exportWatchHistory(token: string) {
  try {
    const res = await fetch(`${API_URL}/me/export`, {
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
    });
    if (!res.ok) return null;
    return res.json();
  } catch {
    return null;
  }
}

export async function importWatchHistory(token: string, payload: unknown) {
  try {
    const res = await fetch(`${API_URL}/me/import`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify(payload),
    });
    if (!res.ok) return null;
    return res.json() as Promise<{ ok: boolean; imported: number; skipped: number; errors: string[] }>;
  } catch {
    return null;
  }
}

export function registerDevice(token: string, payload: { token: string; platform: "android" | "ios" | "web"; app_version?: string }) {
  return fetchAPI<{ ok: boolean }>("/devices", {
    method: "POST",
    cache: "no-store",
    headers: authHeaders(token),
    body: JSON.stringify(payload),
  });
}
