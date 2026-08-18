import { useEffect, useState } from "react";
import { getJob } from "../api/jobs";
import { getNews, rewriteNews, startNewsSync } from "../api/news";
import { defaultMood, getMoodLabel } from "../constants/moods";

const initialStatus = "Подготавливаю ленту новостей.";
const syncPollInterval = 2500;

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
  const [status, setStatus] = useState(initialStatus);
  const [activeJobID, setActiveJobID] = useState("");
  const [activeJobMood, setActiveJobMood] = useState(defaultMood);

  const selectedBaseNews = news.find((item) => item.ExternalID === selectedExternalID) || news[0] || null;
  const selectedNews =
    moodNews[activeMood]?.find((item) => item.ExternalID === selectedExternalID) ||
    selectedBaseNews ||
    moodNews[activeMood]?.[0] ||
    null;

  function handleMoodChange(mood) {
    setActiveMood(mood);
    setError("");
    setStatus(`Выбран режим «${getMoodLabel(mood)}».`);
  }

  async function loadSelectedNewsMood(mood, externalID) {
    setIsDetailLoading(true);
    setError("");

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
    const { limit = 10, silent = false } = options;

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

  useEffect(() => {
    if (!activeJobID) {
      return undefined;
    }

    let cancelled = false;

    async function pollJob() {
      try {
        const job = await getJob(activeJobID);
        if (cancelled) {
          return;
        }

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
      } catch (pollError) {
        if (!cancelled) {
          setError(pollError.message);
          setIsSyncing(false);
          setActiveJobID("");
        }
      }
    }

    void pollJob()
    const intervalID = window.setInterval(() => {
      void pollJob();
    }, syncPollInterval);

    return () => {
      cancelled = true;
      window.clearInterval(intervalID);
    };
  }, [activeJobID, activeMood]);

  useEffect(() => {
    if (!selectedExternalID || activeMood === defaultMood) {
      return;
    }

    if (moodNews[activeMood]?.some((item) => item.ExternalID === selectedExternalID)) {
      return;
    }

    void loadSelectedNewsMood(activeMood, selectedExternalID);
  }, [activeMood, selectedExternalID]);

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
          await syncNews(defaultMood, { limit: 3, silent: true });
          return;
        }

        setMoodNews({
          [defaultMood]: items
        });
        setNews(items);
        setSelectedExternalID(items[0]?.ExternalID || "");
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
    setActiveMood: handleMoodChange,
    setSelectedId: setSelectedExternalID,
    status,
    loadNews,
    selectedExternalID,
    syncNews
  };
}
