import { useEffect, useState } from "react";

const moods = [
  { id: "neutral", label: "Нейтрально" },
  { id: "happy", label: "Радостно" },
  { id: "sad", label: "Грустно" },
  { id: "ironic", label: "Иронично" }
];

const defaultMood = "neutral";
const initialStatus = "Нажми «Загрузить новости», чтобы подтянуть 10 реальных публикаций.";

function formatDate(value) {
  if (!value) {
    return "";
  }

  return new Intl.DateTimeFormat("ru-RU", {
    dateStyle: "medium",
    timeStyle: "short"
  }).format(new Date(value));
}

function excerpt(value, maxLength = 220) {
  if (!value) {
    return "";
  }

  if (value.length <= maxLength) {
    return value;
  }

  return `${value.slice(0, maxLength).trim()}...`;
}

function App() {
  const [activeMood, setActiveMood] = useState(defaultMood);
  const [news, setNews] = useState([]);
  const [selectedId, setSelectedId] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [isSyncing, setIsSyncing] = useState(false);
  const [error, setError] = useState("");
  const [status, setStatus] = useState(initialStatus);

  const selectedNews = news.find((item) => item.ID === selectedId) || news[0] || null;

  async function loadNews(mood) {
    setIsLoading(true);
    setError("");

    try {
      const response = await fetch(`/api/v1/news?mood=${encodeURIComponent(mood)}`);
      if (!response.ok) {
        throw new Error("Не удалось получить список новостей.");
      }

      const items = await response.json();
      setNews(items);
      setSelectedId((current) => {
        if (items.some((item) => item.ID === current)) {
          return current;
        }
        return items[0]?.ID || "";
      });

      if (items.length === 0) {
        setStatus("Новостей пока нет в хранилище. Сначала выполни загрузку.");
      } else {
        setStatus(`Загружено ${items.length} новостей в режиме «${moods.find((moodItem) => moodItem.id === mood)?.label || mood}».`);
      }
    } catch (loadError) {
      setError(loadError.message);
    } finally {
      setIsLoading(false);
    }
  }

  async function syncNews(mood) {
    setIsSyncing(true);
    setError("");

    try {
      const response = await fetch(`/api/v1/news/sync?mood=${encodeURIComponent(mood)}`, {
        method: "POST"
      });

      if (!response.ok) {
        throw new Error("Не удалось загрузить новости из источника.");
      }

      const payload = await response.json();
      setStatus(`Синхронизация завершена. Загружено ${payload.count} новостей.`);
      await loadNews(mood);
    } catch (syncError) {
      setError(syncError.message);
    } finally {
      setIsSyncing(false);
    }
  }

  useEffect(() => {
    void loadNews(activeMood);
  }, [activeMood]);

  return (
    <div className="page">
      <header className="hero">
        <div className="hero__copy">
          <p className="eyebrow">Тестовое задание Cash Factories</p>
          <h1>Новости с переключением эмоционального режима.</h1>
          <p className="hero__text">
            Каркас клиентской части для грида реальных новостей, выбора
            настроения и сравнения исходного текста с переписанной версией.
          </p>
        </div>

        <div className="hero__panel">
          <span className="panel__label">Режим</span>
          <div className="moods">
            {moods.map((mood) => (
              <button
                key={mood.id}
                type="button"
                className={mood.id === activeMood ? "mood mood--active" : "mood"}
                onClick={() => setActiveMood(mood.id)}
                disabled={isLoading || isSyncing}
              >
                {mood.label}
              </button>
            ))}
          </div>

          <div className="hero__actions">
            <button
              type="button"
              className="action-button"
              onClick={() => void syncNews(activeMood)}
              disabled={isSyncing}
            >
              {isSyncing ? "Загружаю..." : "Загрузить новости"}
            </button>
            <button
              type="button"
              className="action-button action-button--ghost"
              onClick={() => void loadNews(activeMood)}
              disabled={isLoading}
            >
              {isLoading ? "Обновляю..." : "Обновить список"}
            </button>
          </div>

          <p className="status-text">{status}</p>
          {error ? <p className="error-text">{error}</p> : null}
        </div>
      </header>

      <main className="layout">
        <section className="news-grid">
          {news.length === 0 ? (
            <div className="empty-state">
              <h2>Новостей пока нет</h2>
              <p>Сделай первую синхронизацию, и здесь появятся реальные публикации из The Guardian.</p>
            </div>
          ) : (
            news.map((item) => (
              <article
                key={item.ID}
                className={item.ID === selectedNews?.ID ? "card card--active" : "card"}
                onClick={() => setSelectedId(item.ID)}
                role="button"
                tabIndex={0}
                onKeyDown={(event) => {
                  if (event.key === "Enter" || event.key === " ") {
                    event.preventDefault();
                    setSelectedId(item.ID);
                  }
                }}
              >
                <span className="card__source">{item.SourceName}</span>
                <h2>{item.Title}</h2>
                <p>{excerpt(item.OriginalText)}</p>
                <div className="card__meta">
                  <span>{formatDate(item.PublishedAt)}</span>
                  <a href={item.SourceURL} target="_blank" rel="noreferrer" onClick={(event) => event.stopPropagation()}>
                    Открыть оригинальный источник
                  </a>
                </div>
              </article>
            ))
          )}
        </section>

        <aside className="detail">
          <div className="detail__head">
            <span className="detail__eyebrow">Выбранная новость</span>
            <h2>Сравнение исходного и переписанного текста</h2>
          </div>

          {selectedNews ? (
            <>
              <div className="detail__summary">
                <h3>{selectedNews.Title}</h3>
                <p>
                  {selectedNews.SourceName} · {formatDate(selectedNews.PublishedAt)}
                </p>
              </div>

              <div className="compare">
                <section>
                  <h3>Оригинал</h3>
                  <p>{selectedNews.OriginalText}</p>
                </section>

                <section>
                  <h3>Рерайт</h3>
                  <p>{selectedNews.RewrittenText}</p>
                </section>
              </div>
            </>
          ) : (
            <div className="empty-state empty-state--detail">
              <h3>Пока нечего показывать</h3>
              <p>После загрузки новостей выбери карточку, и здесь появится подробное сравнение.</p>
            </div>
          )}
        </aside>
      </main>
    </div>
  );
}

export default App;
