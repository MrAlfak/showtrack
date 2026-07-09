import type { CastMember, PersonDetail, SearchItem, ShowDetail, ShowItem } from "./api";

export const demoLibrary: ShowItem[] = [
  {
    id: 1,
    tmdb_id: 1396,
    title: "Breaking Bad",
    status: "Ended",
    vote_average: 8.9,
    poster_url: "https://image.tmdb.org/t/p/w500/ggFHVNu6YYI5L9pCfOacjizRGt.jpg",
    progress: 72,
    watched: 45,
    total: 62,
  },
  {
    id: 2,
    tmdb_id: 66732,
    title: "Stranger Things",
    status: "Returning Series",
    vote_average: 8.6,
    poster_url: "https://image.tmdb.org/t/p/w500/49WJfeN0moxb9IPfGn8AIqMGskD.jpg",
    progress: 41,
    watched: 17,
    total: 42,
  },
  {
    id: 3,
    tmdb_id: 94997,
    title: "House of the Dragon",
    status: "Returning Series",
    vote_average: 8.4,
    poster_url: "https://image.tmdb.org/t/p/w500/z2yahlsthc9HtPJk2Y7ggxfW6zH.jpg",
    progress: 100,
    watched: 18,
    total: 18,
  },
];

export const demoTrending: ShowItem[] = [
  {
    id: 4,
    tmdb_id: 1399,
    title: "Game of Thrones",
    vote_average: 8.4,
    poster_url: "https://image.tmdb.org/t/p/w500/1XS1oqL89opfnbLl8WnZY1O1uJx.jpg",
  },
  ...demoLibrary,
];

export const demoUpcoming = [
  { show: "Stranger Things", episode: "S5E01", date: "Jul 12", poster_url: demoLibrary[1].poster_url },
  { show: "House of the Dragon", episode: "S3E02", date: "Jul 14", poster_url: demoLibrary[2].poster_url },
];

export const demoSearch = (q: string): SearchItem[] => [
  { id: 1396, title: "Breaking Bad", media_type: "tv" as const, overview: "Chemistry teacher turned meth maker.", poster_url: demoLibrary[0].poster_url, vote_average: 8.9 },
  { id: 66732, title: "Stranger Things", media_type: "tv" as const, overview: "Supernatural small-town mystery.", poster_url: demoLibrary[1].poster_url, vote_average: 8.6 },
  { id: 17419, title: "Bryan Cranston", media_type: "person" as const, poster_url: "https://image.tmdb.org/t/p/w185/fnglrDfkcsEuHQgvhoL8n0kev1.jpg" },
].filter((i) => i.title.toLowerCase().includes(q.toLowerCase()) || q.length < 2);

export const demoCast: CastMember[] = [
  { tmdb_id: 17419, name: "Bryan Cranston", character: "Walter White", profile_url: "https://image.tmdb.org/t/p/w185/fnglrDfkcsEuHQgvhoL8n0kev1.jpg" },
  { tmdb_id: 84497, name: "Aaron Paul", character: "Jesse Pinkman", profile_url: "https://image.tmdb.org/t/p/w185/ycyc9NJOvZnsgFsqXJzLpYizwTr.jpg" },
  { tmdb_id: 134531, name: "Anna Gunn", character: "Skyler White", profile_url: "https://image.tmdb.org/t/p/w185/69Mf3FA2b9fVOpy2g2v0v1t2Y1h.jpg" },
  { tmdb_id: 209674, name: "Dean Norris", character: "Hank Schrader", profile_url: "https://image.tmdb.org/t/p/w185/ycyc9NJOvZnsgFsqXJzLpYizwTr.jpg" },
];

export const demoShow = (id: string): ShowDetail => ({
  id: 1,
  tmdb_id: Number(id) || 1396,
  title: "Breaking Bad",
  overview:
    "When an unassuming high school chemistry teacher discovers he has a rare form of lung cancer, he decides to team up with a former student and create a top of the line crystal meth in a used RV, to provide for his family once he is gone.",
  status: "Ended",
  vote_average: 8.9,
  poster_url: demoLibrary[0].poster_url,
  progress: 72,
  watched: 45,
  total: 62,
  cast: demoCast,
  seasons: [
    { id: 1, season_number: 1, name: "Season 1", episode_count: 7, episodes: [] },
    { id: 2, season_number: 2, name: "Season 2", episode_count: 13, episodes: [] },
    { id: 3, season_number: 3, name: "Season 3", episode_count: 13, episodes: [] },
    { id: 4, season_number: 4, name: "Season 4", episode_count: 13, episodes: [] },
    { id: 5, season_number: 5, name: "Season 5", episode_count: 16, episodes: [] },
  ],
});

export const demoPerson = (id: string): PersonDetail => ({
  id: 1,
  tmdb_id: Number(id) || 17419,
  name: "Bryan Cranston",
  biography:
    "Bryan Lee Cranston is an American actor, director, producer and screenwriter. He is known for his roles as Walter White in Breaking Bad and Hal in Malcolm in the Middle.",
  profile_url: "https://image.tmdb.org/t/p/w185/fnglrDfkcsEuHQgvhoL8n0kev1.jpg",
  credits: [
    { media_type: "tv", tmdb_id: 1396, title: "Breaking Bad", character: "Walter White", poster_url: demoLibrary[0].poster_url, vote_average: 8.9 },
    { media_type: "tv", tmdb_id: 1695, title: "Malcolm in the Middle", character: "Hal", poster_url: "https://image.tmdb.org/t/p/w500/6XqfrMnC8bk7t8xYy8j9mJQF9wF.jpg", vote_average: 8.0 },
    { media_type: "movie", tmdb_id: 106646, title: "Argo", character: "Jack O'Donnell", poster_url: "https://image.tmdb.org/t/p/w500/5aGhaIHYuQbzlHWowg5AI8TWhUY.jpg", vote_average: 7.5 },
  ],
});

export const demoStats = {
  shows: 12,
  episodes: 284,
  hours: 142,
  streak: 7,
};
