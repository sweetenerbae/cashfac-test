import { apiClient } from "./client";

export function getJob(jobID) {
  return apiClient.get(`/api/v1/jobs/${encodeURIComponent(jobID)}`);
}
