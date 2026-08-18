import { apiClient } from "./client";

export function getNews(mood) {
  return apiClient.get(`/api/v1/news?mood=${encodeURIComponent(mood)}`);
}

export function rewriteNews(externalID, mood) {
  return apiClient.post(
    `/api/v1/news/rewrite?external_id=${encodeURIComponent(externalID)}&mood=${encodeURIComponent(mood)}`
  );
}

export function syncNews(mood) {
  return apiClient.post(`/api/v1/news/sync?mood=${encodeURIComponent(mood)}`);
}

export function startNewsSync(mood, limit = 10) {
  return apiClient.post(`/api/v1/news/sync?mood=${encodeURIComponent(mood)}&limit=${encodeURIComponent(limit)}`);
}
