import { NewsDetail } from "../components/NewsDetail";
import { moods } from "../constants/moods";

export function NewsArticlePage({
  activeMood,
  error,
  isBootstrapping,
  isDetailLoading,
  isLoading,
  isSyncing,
  onBack,
  selectedNews,
  setActiveMood,
  status
}) {
  return (
    <div className="page page--article">
      <header className="article-topbar">
        <button type="button" className="back-link" onClick={onBack}>
          Все новости
        </button>

        <div className="article-topbar__panel">
          <div className="article-topbar__head">
            <div>
              <span className="panel__label">Режим</span>
              <h2 className="article-topbar__title">Выбери интонацию, в которой показать эту новость</h2>
            </div>
            {selectedNews ? (
              <div className="article-topbar__meta">
                <span>{selectedNews.SourceName}</span>
              </div>
            ) : null}
          </div>
          <div className="moods moods--article">
            {moods.map((mood) => (
              <button
                key={mood.id}
                type="button"
                className={mood.id === activeMood ? "mood mood--active" : "mood"}
                onClick={() => setActiveMood(mood.id)}
                disabled={isLoading || isSyncing}
              >
                <span className="mood__emoji" aria-hidden="true">
                  {mood.emoji}
                </span>
                <span>{mood.label}</span>
              </button>
            ))}
          </div>
          <p className="status-text">{status}</p>
          {error ? <p className="error-text">{error}</p> : null}
        </div>
      </header>

      <section className="article-intro">
        <p className="eyebrow">Выбранная новость</p>
        <h1>Сравнение исходного и переписанного текста</h1>
      </section>

      <main className="article-layout">
        <NewsDetail isLoading={isBootstrapping || isDetailLoading} selectedNews={selectedNews} />
      </main>
    </div>
  );
}
