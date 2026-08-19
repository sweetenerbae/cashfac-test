import { useEffect, useRef, useState } from "react";
import { getNews, getNewsByExternalID, rewriteNews, startNewsSync } from "../api/news";
import { defaultMood, getMoodLabel } from "../constants/moods";
import { DEFAULT_SYNC_LIMIT, INITIAL_STATUS, INITIAL_SYNC_LIMIT } from "./newsPageConstants";
import { useJobPolling } from "./useJobPolling";

function findSelectedNews(news, moodNews, activeMood, selectedExternalID) {
  if (!selectedExternalID) {
    return null;
  }

  const baseVersion = news.find((item) => item.ExternalID === selectedExternalID);
  const activeMoodItems = moodNews[activeMood] || [];
  const moodVersion = activeMoodItems.find((item) => item.ExternalID === selectedExternalID);
  if (moodVersion && (!baseVersion || moodVersion.OriginalDigest === baseVersion.OriginalDigest)) {
    return baseVersion?.ImageURL ? { ...moodVersion, ImageURL: baseVersion.ImageURL } : moodVersion;
  }

  if (baseVersion) {
    return baseVersion;
  }

  return null;
}

function hasRewrite(item, baseItem) {
  return Boolean(
    item?.RewrittenText?.trim() && (!baseItem || item.OriginalDigest === baseItem.OriginalDigest)
  );
}

function rewriteKey(externalID, mood) {
  return `${externalID}:${mood}`;
}

export function useNewsPage() {
  const [activeMood, setActiveMood] = useState(defaultMood);
  const [news, setNews] = useState([]);
  const [moodNews, setMoodNews] = useState({});
  const [selectedExternalID, setSelectedExternalID] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [isDetailLoading, setIsDetailLoading] = useState(false);
  const [isSyncing, setIsSyncing] = useState(false);
  const [isBootstrapping, setIsBootstrapping] = useState(true);
  const [error, setError] = useState("");
  const [status, setStatus] = useState(INITIAL_STATUS);
  const [rewriteFallbackMessage, setRewriteFallbackMessage] = useState("");
  const [activeJobID, setActiveJobID] = useState("");
  const [syncFeedback, setSyncFeedback] = useState("");
  const [rewriteMeta, setRewriteMeta] = useState({});
  const detailRequestIDRef = useRef(0);
  const syncInProgressRef = useRef(false);

  const selectedBaseNews = news.find((item) => item.ExternalID === selectedExternalID) || null;
  const selectedNews = findSelectedNews(news, moodNews, activeMood, selectedExternalID);
  const selectedRewriteMeta = rewriteMeta[rewriteKey(selectedExternalID, activeMood)] || null;

  function beginDetailRequest() {
    const requestID = detailRequestIDRef.current + 1;
    detailRequestIDRef.current = requestID;
    setIsDetailLoading(true);
    return requestID;
  }

  function isCurrentDetailRequest(requestID) {
    return detailRequestIDRef.current === requestID;
  }

  function finishDetailRequest(requestID) {
    if (isCurrentDetailRequest(requestID)) {
      setIsDetailLoading(false);
    }
  }

  function handleSelectedIDChange(externalID) {
    detailRequestIDRef.current += 1;
    setIsDetailLoading(false);
    setError("");
    setRewriteFallbackMessage("");
    setSelectedExternalID(externalID);
  }

  function handleMoodChange(mood) {
    const cachedItem = moodNews[mood]?.find((item) => item.ExternalID === selectedExternalID);

    detailRequestIDRef.current += 1;
    setIsDetailLoading(false);
    setActiveMood(mood);
    setError("");
    setRewriteFallbackMessage("");
    setStatus(`Выбран режим «${getMoodLabel(mood)}».`);

    if (!selectedBaseNews) {
      return;
    }
    if (hasRewrite(cachedItem, selectedBaseNews)) {
      const key = rewriteKey(selectedExternalID, mood);
      setRewriteMeta((current) => ({
        ...current,
        [key]: current[key] || {
          Source: "local",
          LLMRequests: 0,
          SavedLLMRequests: 1,
          DurationMs: 0
        }
      }));
      return;
    }

    void loadSelectedNewsMood(mood, selectedExternalID);
  }

  async function loadSelectedNewsMood(mood, externalID) {
    if (!externalID || !selectedBaseNews) {
      return null;
    }

    const requestID = beginDetailRequest();
    setError("");
    setRewriteFallbackMessage("");

    try {
      const result = await rewriteNews(externalID, mood);
      const item = result.News || result;
      setMoodNews((current) => {
        const currentItems = current[mood] || [];
        const nextItems = currentItems.some((newsItem) => newsItem.ExternalID === externalID)
          ? currentItems.map((newsItem) => (newsItem.ExternalID === externalID ? item : newsItem))
          : [item, ...currentItems];

        return {
          ...current,
          [mood]: nextItems
        };
      });
      if (result.Meta) {
        setRewriteMeta((current) => ({
          ...current,
          [rewriteKey(externalID, mood)]: result.Meta
        }));
      }
      if (isCurrentDetailRequest(requestID)) {
        setStatus(`Новости в режиме «${getMoodLabel(mood)}» готовы.`);
      }
      return item;
    } catch (loadError) {
      if (isCurrentDetailRequest(requestID)) {
        setError(loadError.message);
        setRewriteFallbackMessage(buildRewriteFallbackMessage(mood, loadError.message));
      }
      return null;
    } finally {
      finishDetailRequest(requestID);
    }
  }

  async function ensureSelectedBaseNews(externalID) {
    if (!externalID) {
      return null;
    }

    const existingItem = news.find((item) => item.ExternalID === externalID);
    if (existingItem) {
      return existingItem;
    }

    const requestID = beginDetailRequest();
    setError("");
    setRewriteFallbackMessage("");

    try {
      const item = await getNewsByExternalID(externalID);
      if (!isCurrentDetailRequest(requestID)) {
        return item;
      }

      setNews((current) => {
        if (current.some((newsItem) => newsItem.ExternalID === externalID)) {
          return current;
        }

        return [item, ...current];
      });
      setMoodNews((current) => {
        const neutralItems = current[defaultMood] || [];
        if (neutralItems.some((newsItem) => newsItem.ExternalID === externalID)) {
          return current;
        }

        return {
          ...current,
          [defaultMood]: [item, ...neutralItems]
        };
      });
      if (isCurrentDetailRequest(requestID)) {
        setError("");
      }
      return item;
    } catch (loadError) {
      if (isCurrentDetailRequest(requestID)) {
        setError(loadError.message);
      }
      return null;
    } finally {
      finishDetailRequest(requestID);
    }
  }

  async function loadNews(mood, options = {}) {
    const { clearError = true } = options;

    setIsLoading(true);
    if (clearError) {
      setError("");
    }

    try {
      const items = await getNews(mood);
      setMoodNews((current) => ({
        ...current,
        [mood]: items
      }));

      setNews(items);
      setSelectedExternalID((current) => {
        if (items.some((item) => item.ExternalID === current)) {
          return current;
        }

        return items[0]?.ExternalID || "";
      });

      if (items.length === 0) {
        setStatus("Пока здесь пусто.");
      } else {
        setStatus(`Загружено ${items.length} новостей в режиме «${getMoodLabel(mood)}».`);
      }

      return items;
    } catch (loadError) {
      setError(loadError.message);
      return null;
    } finally {
      setIsLoading(false);
    }
  }

  async function syncNews(options = {}) {
    const { limit = DEFAULT_SYNC_LIMIT, silent = false } = options;

    if (syncInProgressRef.current) {
      return;
    }

    syncInProgressRef.current = true;
    setIsSyncing(true);
    setError("");
    setSyncFeedback("");

    try {
      const job = await startNewsSync(limit);
      setActiveJobID(job.ID);
      if (!silent) {
        setStatus("Собираю свежие новости.");
      }
    } catch (syncError) {
      setError(syncError.message);
      setIsSyncing(false);
      syncInProgressRef.current = false;
    }
  }

  useJobPolling(activeJobID, {
    onUpdate: async (job) => {
      const total = job.TotalCount || job.Limit || 0;
      const processed = job.ProcessedCount || 0;

      if (job.Status === "running" || job.Status === "pending") {
        setStatus(`Собираю новости: ${processed} из ${total}.`);
        return;
      }

      if (job.Status === "failed") {
        const failureMessage = job.Error || "Синхронизация завершилась с ошибкой.";
        setError(failureMessage);
        setStatus(`Не удалось закончить загрузку. Успела обработаться ${processed} из ${total}.`);
        setActiveJobID("");
        await loadNews(defaultMood, { clearError: false });
        setError(failureMessage);
        setIsSyncing(false);
        syncInProgressRef.current = false;
        return;
      }

      if (job.Status === "completed") {
        const previousExternalIDs = new Set(news.map((item) => item.ExternalID));
        setActiveJobID("");
        const items = await loadNews(defaultMood);
        if (items) {
          const addedCount = items.filter((item) => !previousExternalIDs.has(item.ExternalID)).length;
          setSyncFeedback(
            addedCount === 0
              ? "Новых публикаций пока нет."
              : formatAddedNewsMessage(addedCount)
          );
        }
        setIsSyncing(false);
        syncInProgressRef.current = false;
        return;
      }

      throw new Error(`Неизвестный статус синхронизации: ${job.Status}`);
    },
    onError: (pollError) => {
      setError(pollError.message);
      setIsSyncing(false);
      syncInProgressRef.current = false;
      setActiveJobID("");
    }
  });

  useEffect(() => {
    if (!syncFeedback) {
      return undefined;
    }

    const timeoutID = window.setTimeout(() => {
      setSyncFeedback("");
    }, 5000);

    return () => {
      window.clearTimeout(timeoutID);
    };
  }, [syncFeedback]);

  useEffect(() => {
    if (!selectedExternalID || selectedBaseNews) {
      return;
    }

    void ensureSelectedBaseNews(selectedExternalID);
  }, [selectedExternalID, selectedBaseNews]);

  useEffect(() => {
    let ignore = false;

    async function bootstrap() {
      setIsBootstrapping(true);
      setStatus("Загружаю новости.");

      try {
        const items = await getNews(defaultMood);
        if (ignore) {
          return;
        }

        if (items.length === 0) {
          setStatus("Собираю первые новости для ленты.");
          await syncNews({ limit: INITIAL_SYNC_LIMIT, silent: true });
          return;
        }

        setMoodNews({
          [defaultMood]: items
        });
        setNews(items);
        setSelectedExternalID((current) => {
          if (current) {
            return current;
          }

          return items[0]?.ExternalID || "";
        });
        setStatus(`Загружено ${items.length} новостей.`);
      } catch (loadError) {
        if (!ignore) {
          setError(loadError.message);
        }
      } finally {
        if (!ignore) {
          setIsBootstrapping(false);
        }
      }
    }

    void bootstrap();

    return () => {
      ignore = true;
    };
  }, []);

  return {
    activeMood,
    error,
    isBootstrapping,
    isDetailLoading,
    isLoading,
    isSyncing,
    news,
    selectedNews,
    selectedRewriteMeta,
    setActiveMood: handleMoodChange,
    setSelectedId: handleSelectedIDChange,
    status,
    syncFeedback,
    rewriteFallbackMessage,
    loadNews,
    selectedExternalID,
    syncNews
  };
}

function formatAddedNewsMessage(count) {
  const lastTwoDigits = count % 100;
  const lastDigit = count % 10;

  if (lastTwoDigits >= 11 && lastTwoDigits <= 14) {
    return `Добавлено ${count} новых публикаций.`;
  }
  if (lastDigit === 1) {
    return `Добавлена ${count} новая публикация.`;
  }
  if (lastDigit >= 2 && lastDigit <= 4) {
    return `Добавлены ${count} новые публикации.`;
  }

  return `Добавлено ${count} новых публикаций.`;
}

function buildRewriteFallbackMessage(mood, serverMessage) {
  const label = getMoodLabel(mood).toLowerCase();
  if (serverMessage && serverMessage.toLowerCase().includes("distinct enough")) {
    return `Пока не удалось подготовить достаточно выразительную версию текста в режиме «${label}». Показываю исходный материал, чтобы не искажать смысл.`;
  }

  return `Пока не удалось подготовить версию текста в режиме «${label}». Показываю исходный материал, чтобы не искажать смысл.`;
}
