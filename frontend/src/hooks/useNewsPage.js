import { useEffect, useState } from "react";
import { getNews, syncNews as syncNewsRequest } from "../api/news";
import { defaultMood, getMoodLabel } from "../constants/moods";

const initialStatus = "Подготавливаю ленту новостей.";

export function useNewsPage() {
  const [activeMood, setActiveMood] = useState(defaultMood);
  const [news, setNews] = useState([]);
  const [selectedId, setSelectedId] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [isSyncing, setIsSyncing] = useState(false);
  const [isBootstrapping, setIsBootstrapping] = useState(true);
  const [error, setError] = useState("");
  const [status, setStatus] = useState(initialStatus);

  const selectedNews = news.find((item) => item.ID === selectedId) || news[0] || null;

  async function loadNews(mood) {
    setIsLoading(true);
    setError("");

    try {
      const items = await getNews(mood);
      setNews(items);
      setSelectedId((current) => {
        if (items.some((item) => item.ID === current)) {
          return current
        }

        return items[0]?.ID || "";
      });

      if (items.length === 0) {
        setStatus("Новостей пока нет в хранилище. Сначала выполни загрузку.");
      } else {
        setStatus(`Загружено ${items.length} новостей в режиме «${getMoodLabel(mood)}».`);
      }
    } catch (loadError) {
      setError(loadError.message);
    } finally {
      setIsLoading(false);
    }
  }

  async function syncNews(mood, options = {}) {
    const { silent = false } = options;

    setIsSyncing(true);
    setError("");

    try {
      const payload = await syncNewsRequest(mood);
      if (!silent) {
        setStatus(`Синхронизация завершена. Загружено ${payload.count} новостей.`);
      }
      await loadNews(mood);
    } catch (syncError) {
      setError(syncError.message);
    } finally {
      setIsSyncing(false);
    }
  }

  useEffect(() => {
    let ignore = false;

    async function bootstrap() {
      setIsBootstrapping(true);
      setStatus("Загружаю новости.");

      try {
        const items = await getNews(activeMood);
        if (ignore) {
          return;
        }

        if (items.length === 0) {
          setStatus("Новостей пока нет. Загружаю 10 реальных публикаций из источника.");
          await syncNews(activeMood, { silent: true });
          return;
        }

        setNews(items);
        setSelectedId(items[0]?.ID || "");
        setStatus(`Загружено ${items.length} новостей в режиме «${getMoodLabel(activeMood)}».`);
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
  }, [activeMood]);

  return {
    activeMood,
    error,
    isBootstrapping,
    isLoading,
    isSyncing,
    news,
    selectedNews,
    setActiveMood,
    setSelectedId,
    status,
    loadNews,
    syncNews
  };
}
