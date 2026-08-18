import { useEffect, useRef } from "react";
import { getJob } from "../api/jobs";
import { SYNC_POLL_INTERVAL } from "./newsPageConstants";

export function useJobPolling(jobID, handlers) {
  const handlersRef = useRef(handlers);

  useEffect(() => {
    handlersRef.current = handlers;
  }, [handlers]);

  useEffect(() => {
    if (!jobID) {
      return undefined;
    }

    let cancelled = false;

    async function pollJob() {
      try {
        const job = await getJob(jobID);
        if (cancelled) {
          return;
        }

        handlersRef.current?.onUpdate?.(job);
      } catch (error) {
        if (!cancelled) {
          handlersRef.current?.onError?.(error);
        }
      }
    }

    void pollJob();
    const intervalID = window.setInterval(() => {
      void pollJob();
    }, SYNC_POLL_INTERVAL);

    return () => {
      cancelled = true;
      window.clearInterval(intervalID);
    };
  }, [jobID]);
}
