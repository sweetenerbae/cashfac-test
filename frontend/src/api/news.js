import { apiClient } from "./client";

export function getNews(mood) {
  return apiClient.get(`/api/v1/news?mood=${encodeURIComponent(mood)}`);
}

export function getNewsByExternalID(externalID, mood = "") {
  const searchParams = new URLSearchParams({
    external_id: externalID
  });

  if (mood) {
    searchParams.set("mood", mood);
  }

  return apiClient.get(`/api/v1/news/by-external?${searchParams.toString()}`);
}

export function rewriteNews(externalID, mood) {
  return apiClient.post(
    `/api/v1/news/rewrite?external_id=${encodeURIComponent(externalID)}&mood=${encodeURIComponent(mood)}`
  );
}

export function startNewsSync(limit = 10) {
  return apiClient.post(`/api/v1/news/sync?limit=${encodeURIComponent(limit)}`);
}
