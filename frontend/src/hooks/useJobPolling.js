import { useEffect, useRef } from "react";
import { getJob } from "../api/jobs";
import { SYNC_POLL_INTERVAL } from "./newsPageConstants";

export function useJobPolling(jobID, handlers) {
  const handlersRef = useRef(handlers);
  handlersRef.current = handlers;

  useEffect(() => {
    if (!jobID) {
      return undefined;
    }

    let cancelled = false;
    let timeoutID;

    async function pollJob() {
      try {
        const job = await getJob(jobID);
        if (cancelled) {
          return;
        }

        await handlersRef.current?.onUpdate?.(job);
        if (!cancelled && (job.Status === "pending" || job.Status === "running")) {
          timeoutID = window.setTimeout(pollJob, SYNC_POLL_INTERVAL);
        }
      } catch (error) {
        if (!cancelled) {
          await handlersRef.current?.onError?.(error);
        }
      }
    }

    void pollJob();

    return () => {
      cancelled = true;
      window.clearTimeout(timeoutID);
    };
  }, [jobID]);
}
