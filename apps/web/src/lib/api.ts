export const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1";

export type SearchItem = {
  id: number;
  title: string;
  overview?: string;
  media_type: "tv" | "movie" | "person";
  poster_url: string;
  vote_average?: number;
};

export type ListStatus = "watching" | "plan_to_watch" | "watched" | "dropped" | "archived";

export type ShowItem = {
  id: number;
  tmdb_id: number;
  title: string;
  overview?: string;
  status?: string;
  list_status?: ListStatus;
  media_type?: "tv" | "movie";
  vote_average?: number;
  poster_url: string;
  progress?: number;
  watched?: number;
  total?: number;
  rec_reason?: string;
  score?: number;
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
  user_rating?: { score: number; review: string };
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
  user_rating?: { score: number; review: string };
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
  binge_today: number;
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

export type ApiError = {
  ok: false;
  error: string;
};

export type ApiSuccess<T> = {
  ok: true;
  data: T;
};

export type ApiResult<T> = ApiSuccess<T> | ApiError;

async function parseError(res: Response): Promise<string> {
  try {
    const body = (await res.json()) as { message?: string; error?: string };
    return body.message ?? body.error ?? `Request failed (${res.status})`;
  } catch {
    return `Request failed (${res.status})`;
  }
}

async function fetchAPI<T>(path: string, init?: RequestInit): Promise<T | null> {
  try {
    const res = await fetch(`${API_URL}${path}`, {
      cache: "no-store",
      ...init,
      headers: {
        "Content-Type": "application/json",
        ...(init?.headers ?? {}),
      },
    });
    if (!res.ok) return null;
    if (res.status === 204) return null;
    return res.json();
  } catch {
    return null;
  }
}

async function fetchAuthAction<T>(path: string, init?: RequestInit): Promise<ApiResult<T>> {
  try {
    const res = await fetch(`${API_URL}${path}`, {
      cache: "no-store",
      ...init,
      headers: {
        "Content-Type": "application/json",
        ...(init?.headers ?? {}),
      },
    });
    if (!res.ok) {
      return { ok: false, error: await parseError(res) };
    }
    return { ok: true, data: (await res.json()) as T };
  } catch {
    return { ok: false, error: "Network error" };
  }
}

function authHeaders(token?: string) {
  return token ? ({ Authorization: `Bearer ${token}` } satisfies Record<string, string>) : undefined;
}

export type Genre = { id: number; name: string };

export type RecommendationsResponse = {
  results: ShowItem[];
  seed_title?: string;
  engine?: string;
  explanation?: string;
  signal_counts?: Record<string, number>;
};

export function getGenres(mediaType: "tv" | "movie" = "tv") {
  return fetchAPI<{ genres: Genre[] }>(`/genres?type=${mediaType}`);
}

export function discover(mediaType: "tv" | "movie" = "tv", genreId?: number, sort = "popularity.desc") {
  const params = new URLSearchParams({ type: mediaType, sort });
  if (genreId) params.set("genre", String(genreId));
  return fetchAPI<{ results: ShowItem[] }>(`/discover?${params}`);
}

export function getRecommendations(token: string) {
  return fetchAPI<RecommendationsResponse>("/me/recommendations", { headers: authHeaders(token) });
}

export function submitOnboarding(
  token: string,
  items: { tmdb_id: number; media_type: "tv" | "movie" }[]
) {
  return fetchAPI<{ ok: boolean; added: number }>("/me/onboarding", {
    method: "POST",
    headers: authHeaders(token),
    body: JSON.stringify({ items }),
  });
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

export function getLibrary(token: string, listStatus?: ListStatus) {
  const query = listStatus ? `?list_status=${listStatus}` : "";
  return fetchAPI<LibraryResponse>(`/me/library${query}`, { headers: authHeaders(token) });
}

export function getDashboard(token: string) {
  return fetchAPI<DashboardResponse>("/me/dashboard", { headers: authHeaders(token) });
}

export function register(payload: { email: string; password: string; display_name?: string }) {
  return fetchAuthAction<AuthResponse>("/auth/register", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export function login(payload: { email: string; password: string }) {
  return fetchAuthAction<AuthResponse>("/auth/login", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export function loginWithGoogle(idToken: string) {
  return fetchAuthAction<AuthResponse & { is_new?: boolean }>("/auth/google", {
    method: "POST",
    body: JSON.stringify({ id_token: idToken }),
  });
}

export type TraktStatus = {
  connected: boolean;
  username?: string;
};

export function getTraktStatus(token: string) {
  return fetchAPI<TraktStatus>("/me/trakt", { headers: authHeaders(token) });
}

export function startTraktConnect(token: string) {
  return fetchAPI<{ url: string }>("/me/trakt/connect", {
    method: "POST",
    headers: authHeaders(token),
  });
}

export function syncTrakt(token: string) {
  return fetchAPI<{ ok: boolean; imported: number; skipped: number }>("/me/trakt/sync", {
    method: "POST",
    headers: authHeaders(token),
  });
}

export function disconnectTrakt(token: string) {
  return fetchAPI<{ ok: boolean }>("/me/trakt", {
    method: "DELETE",
    headers: authHeaders(token),
  });
}

export function addShow(tmdbId: number, token: string, listStatus: ListStatus = "watching") {
  return fetchAPI<{ ok: boolean; show_id: number }>("/shows", {
    method: "POST",
    headers: authHeaders(token),
    body: JSON.stringify({ tmdb_id: tmdbId, list_status: listStatus }),
  });
}

export function removeShow(showId: number, token: string) {
  return fetchAPI<{ ok: boolean }>(`/shows/${showId}`, {
    method: "DELETE",
    headers: authHeaders(token),
  });
}

export function addMovie(tmdbId: number, token: string, listStatus: ListStatus = "watching") {
  return fetchAPI<{ ok: boolean; movie_id: number }>("/movies", {
    method: "POST",
    headers: authHeaders(token),
    body: JSON.stringify({ tmdb_id: tmdbId, list_status: listStatus }),
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

export function updateShowStatus(showId: number, listStatus: ListStatus, token: string) {
  return fetchAPI<{ ok: boolean }>(`/shows/${showId}/status`, {
    method: "PATCH",
    headers: authHeaders(token),
    body: JSON.stringify({ list_status: listStatus }),
  });
}

export function updateMovieStatus(movieId: number, listStatus: ListStatus, token: string) {
  return fetchAPI<{ ok: boolean }>(`/movies/${movieId}/status`, {
    method: "PATCH",
    headers: authHeaders(token),
    body: JSON.stringify({ list_status: listStatus }),
  });
}

export function registerDevice(token: string, payload: { token: string; platform: "android" | "ios" | "web"; app_version?: string }) {
  return fetchAPI<{ ok: boolean }>("/devices", {
    method: "POST",
    cache: "no-store",
    headers: authHeaders(token),
    body: JSON.stringify(payload),
  });
}

export type UserSummary = {
  id: string;
  username: string;
  display_name: string;
  avatar_url: string;
  is_following?: boolean;
};

export type UserProfile = UserSummary & {
  bio: string;
  is_public: boolean;
  stats: DashboardStats;
  followers: number;
  following: number;
  library_preview?: Array<{
    tmdb_id: number;
    title: string;
    poster_url: string;
    media_type: "tv" | "movie";
  }>;
};

export type ActivityItem = {
  id: number;
  user_id: string;
  activity_type: string;
  payload: {
    title?: string;
    tmdb_id?: number;
    media_type?: "tv" | "movie";
    poster_url?: string;
    season_number?: number;
    episode_number?: number;
    episode_name?: string;
    list_status?: ListStatus;
  };
  created_at: string;
  username: string;
  display_name: string;
  avatar_url: string;
};

export function getMyProfile(token: string) {
  return fetchAPI<UserProfile>("/me/profile", { headers: authHeaders(token) });
}

export function updateMyProfile(
  token: string,
  body: Partial<{ username: string; display_name: string; bio: string; is_public: boolean }>
) {
  return fetchAPI<UserProfile>("/me/profile", {
    method: "PATCH",
    headers: authHeaders(token),
    body: JSON.stringify(body),
  });
}

export function searchUsers(token: string, query: string) {
  return fetchAPI<{ results: UserSummary[] }>(`/users/search?q=${encodeURIComponent(query)}`, {
    headers: authHeaders(token),
  });
}

export function getUserProfile(token: string, userId: string) {
  return fetchAPI<UserProfile>(`/users/${userId}`, { headers: authHeaders(token) });
}

export function followUser(token: string, userId: string) {
  return fetchAPI<{ ok: boolean }>(`/users/${userId}/follow`, {
    method: "POST",
    headers: authHeaders(token),
  });
}

export function unfollowUser(token: string, userId: string) {
  return fetchAPI<{ ok: boolean }>(`/users/${userId}/follow`, {
    method: "DELETE",
    headers: authHeaders(token),
  });
}

export function getFeed(token: string, beforeId?: number) {
  const query = beforeId ? `?before_id=${beforeId}` : "";
  return fetchAPI<{ items: ActivityItem[] }>(`/me/feed${query}`, { headers: authHeaders(token) });
}

export function getFollowing(token: string) {
  return fetchAPI<{ results: UserSummary[] }>("/me/following", { headers: authHeaders(token) });
}

export type CustomListSummary = {
  id: string;
  name: string;
  description: string;
  is_public: boolean;
  item_count: number;
};

export type CustomListItem = {
  media_type: "tv" | "movie";
  tmdb_id: number;
  title: string;
  poster_url: string;
};

export function setMyRating(token: string, mediaType: "tv" | "movie", tmdbId: number, score: number, review: string) {
  return fetchAPI<{ ok: boolean }>(`/me/ratings/${mediaType}/${tmdbId}`, {
    method: "PUT",
    headers: authHeaders(token),
    body: JSON.stringify({ score, review }),
  });
}

export function deleteMyRating(token: string, mediaType: "tv" | "movie", tmdbId: number) {
  return fetchAPI<{ ok: boolean }>(`/me/ratings/${mediaType}/${tmdbId}`, {
    method: "DELETE",
    headers: authHeaders(token),
  });
}

export function getMyLists(token: string) {
  return fetchAPI<{ lists: CustomListSummary[] }>("/me/lists", { headers: authHeaders(token) });
}

export function createList(token: string, name: string, description = "") {
  return fetchAPI<{ ok: boolean; id: string }>("/me/lists", {
    method: "POST",
    headers: authHeaders(token),
    body: JSON.stringify({ name, description }),
  });
}

export function getList(token: string, listId: string) {
  return fetchAPI<{ list: CustomListSummary; items: CustomListItem[] }>(`/me/lists/${listId}`, {
    headers: authHeaders(token),
  });
}

export function deleteList(token: string, listId: string) {
  return fetchAPI<{ ok: boolean }>(`/me/lists/${listId}`, {
    method: "DELETE",
    headers: authHeaders(token),
  });
}

export function addListItem(
  token: string,
  listId: string,
  item: { media_type: "tv" | "movie"; tmdb_id: number; title: string; poster_url: string }
) {
  return fetchAPI<{ ok: boolean }>(`/me/lists/${listId}/items`, {
    method: "POST",
    headers: authHeaders(token),
    body: JSON.stringify(item),
  });
}

async function fetchAdmin<T>(path: string, adminToken: string): Promise<T | null> {
  try {
    const res = await fetch(`${API_URL}${path}`, {
      cache: "no-store",
      headers: {
        "Content-Type": "application/json",
        "X-Admin-Token": adminToken,
      },
    });
    if (!res.ok) return null;
    return res.json() as Promise<T>;
  } catch {
    return null;
  }
}

export function getAdminStats(adminToken: string) {
  return fetchAdmin<Record<string, number>>("/admin/stats", adminToken);
}

export function getAdminUsers(adminToken: string) {
  return fetchAdmin<{ users: Array<Record<string, string>> }>("/admin/users", adminToken);
}
