import { useEffect, useState } from "react";
import { getNews, getNewsByExternalID, rewriteNews, startNewsSync } from "../api/news";
import { defaultMood, getMoodLabel } from "../constants/moods";
import { DEFAULT_SYNC_LIMIT, INITIAL_STATUS, INITIAL_SYNC_LIMIT } from "./newsPageConstants";
import { useJobPolling } from "./useJobPolling";

function findSelectedNews(news, moodNews, activeMood, selectedExternalID) {
  const activeMoodItems = moodNews[activeMood] || [];
  const moodVersion = activeMoodItems.find((item) => item.ExternalID === selectedExternalID);
  if (moodVersion) {
    return moodVersion;
  }

  const baseVersion = news.find((item) => item.ExternalID === selectedExternalID);
  if (baseVersion) {
    return baseVersion;
  }

  if (activeMoodItems.length > 0) {
    return activeMoodItems[0];
  }

  return news[0] || null;
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
  const [activeJobMood, setActiveJobMood] = useState(defaultMood);

  const selectedBaseNews = news.find((item) => item.ExternalID === selectedExternalID) || null;
  const selectedNews = findSelectedNews(news, moodNews, activeMood, selectedExternalID);

  function handleMoodChange(mood) {
    setActiveMood(mood);
    setError("");
    setRewriteFallbackMessage("");
    setStatus(`Выбран режим «${getMoodLabel(mood)}».`);
  }

  async function loadSelectedNewsMood(mood, externalID) {
    if (!externalID || !selectedBaseNews) {
      return null;
    }

    setIsDetailLoading(true);
    setError("");
    setRewriteFallbackMessage("");

    try {
      const item = await rewriteNews(externalID, mood);
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
      setStatus(`Новости в режиме «${getMoodLabel(mood)}» готовы.`);
      return item;
    } catch (loadError) {
      setError(loadError.message);
      setRewriteFallbackMessage(buildRewriteFallbackMessage(mood, loadError.message));
      return null;
    } finally {
      setIsDetailLoading(false);
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

    setIsDetailLoading(true);
    setError("");
    setRewriteFallbackMessage("");

    try {
      const item = await getNewsByExternalID(externalID);
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
      return item;
    } catch (loadError) {
      setError(loadError.message);
      return null;
    } finally {
      setIsDetailLoading(false);
    }
  }

  async function loadNews(mood, options = {}) {
    const { replaceList = false } = options;

    if (replaceList) {
      setIsLoading(true);
    } else {
      setIsDetailLoading(true);
    }
    setError("");

    try {
      const items = await getNews(mood);
      setMoodNews((current) => ({
        ...current,
        [mood]: items
      }));

      if (replaceList) {
        setNews(items);
        setSelectedExternalID((current) => {
          if (items.some((item) => item.ExternalID === current)) {
            return current;
          }

          return items[0]?.ExternalID || "";
        });
      }

      if (items.length === 0) {
        setStatus("Пока здесь пусто.");
      } else {
        setStatus(`Загружено ${items.length} новостей в режиме «${getMoodLabel(mood)}».`);
      }
    } catch (loadError) {
      setError(loadError.message);
    } finally {
      setIsLoading(false);
      setIsDetailLoading(false);
    }
  }

  async function syncNews(mood, options = {}) {
    const { limit = DEFAULT_SYNC_LIMIT, silent = false } = options;

    setIsSyncing(true);
    setError("");

    try {
      const job = await startNewsSync(mood, limit);
      setActiveJobID(job.ID);
      setActiveJobMood(job.Mood || mood);
      if (!silent) {
        setStatus("Собираю свежие новости.");
      }
    } catch (syncError) {
      setError(syncError.message);
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
        setError(job.Error || "Синхронизация завершилась с ошибкой.");
        setStatus(`Не удалось закончить загрузку. Успела обработаться ${processed} из ${total}.`);
        setIsSyncing(false);
        setActiveJobID("");
        await loadNews(activeJobMood, { replaceList: activeJobMood === defaultMood });
        return;
      }

      if (job.Status === "completed") {
        setStatus(`Готово. Загружено ${processed} новостей.`);
        setIsSyncing(false);
        setActiveJobID("");
        await loadNews(activeJobMood, { replaceList: activeJobMood === defaultMood });
      }
    },
    onError: (pollError) => {
      setError(pollError.message);
      setIsSyncing(false);
      setActiveJobID("");
    }
  });

  useEffect(() => {
    if (!selectedExternalID || activeMood === defaultMood) {
      return;
    }

    if (!selectedBaseNews) {
      void ensureSelectedBaseNews(selectedExternalID);
      return;
    }

    if (moodNews[activeMood]?.some((item) => item.ExternalID === selectedExternalID)) {
      return;
    }

    void loadSelectedNewsMood(activeMood, selectedExternalID);
  }, [activeMood, selectedExternalID, selectedBaseNews]);

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
          await syncNews(defaultMood, { limit: INITIAL_SYNC_LIMIT, silent: true });
          return;
        }

        setMoodNews({
          [defaultMood]: items
        });
        setNews(items);
        setSelectedExternalID((current) => {
          if (current && items.some((item) => item.ExternalID === current)) {
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

  useEffect(() => {
    if (!selectedExternalID || selectedNews) {
      return;
    }

    setError("Эта новость пока недоступна в локальной подборке.");
  }, [selectedExternalID, selectedNews]);

  return {
    activeMood,
    error,
    isBootstrapping,
    isDetailLoading,
    isLoading,
    isSyncing,
    news,
    selectedNews,
    setActiveMood: handleMoodChange,
    setSelectedId: setSelectedExternalID,
    status,
    rewriteFallbackMessage,
    loadNews,
    selectedExternalID,
    syncNews
  };
}

function buildRewriteFallbackMessage(mood, serverMessage) {
  const label = getMoodLabel(mood).toLowerCase();
  if (serverMessage && serverMessage.toLowerCase().includes("distinct enough")) {
    return `Пока не удалось подготовить достаточно выразительную версию текста в режиме «${label}». Показываю исходный материал, чтобы не искажать смысл.`;
  }

  return `Пока не удалось подготовить версию текста в режиме «${label}». Показываю исходный материал, чтобы не искажать смысл.`;
}
