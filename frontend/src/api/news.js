import { apiClient } from "./client";

export function getNews(mood) {
  return apiClient.get(`/api/v1/news?mood=${encodeURIComponent(mood)}`);
}

export function syncNews(mood) {
  return apiClient.post(`/api/v1/news/sync?mood=${encodeURIComponent(mood)}`);
}
