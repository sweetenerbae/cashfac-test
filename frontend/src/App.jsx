import { NewsCard } from "./components/NewsCard";
import { NewsDetail } from "./components/NewsDetail";
import { NewsSkeletonGrid } from "./components/NewsSkeletonGrid";
import { moods } from "./constants/moods";
import { useNewsPage } from "./hooks/useNewsPage";

function App() {
  const {
    activeMood,
    error,
    isBootstrapping,
    isLoading,
    isSyncing,
    loadNews,
    news,
    selectedNews,
    setActiveMood,
    setSelectedId,
    status,
    syncNews
  } = useNewsPage();

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
          {isBootstrapping || isLoading || isSyncing ? (
            <NewsSkeletonGrid />
          ) : news.length === 0 ? (
            <div className="empty-state">
              <h2>Новостей пока нет</h2>
              <p>Сделай первую синхронизацию, и здесь появятся реальные публикации из The Guardian.</p>
            </div>
          ) : (
            news.map((item) => (
              <NewsCard
                key={item.ID}
                isActive={item.ID === selectedNews?.ID}
                item={item}
                onSelect={setSelectedId}
              />
            ))
          )}
        </section>

        <aside className="detail">
          <div className="detail__head">
            <span className="detail__eyebrow">Выбранная новость</span>
            <h2>Сравнение исходного и переписанного текста</h2>
          </div>

          <NewsDetail isLoading={isBootstrapping || isLoading || isSyncing} selectedNews={selectedNews} />
        </aside>
      </main>
    </div>
  );
}

export default App;
