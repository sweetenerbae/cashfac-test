import { useEffect, useState } from "react";
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
  rewriteFallbackMessage,
  selectedNews,
  setActiveMood,
  status
}) {
  const [showOriginal, setShowOriginal] = useState(true);

  useEffect(() => {
    setShowOriginal(true);
  }, [selectedNews?.ExternalID]);

  return (
    <div className="page page--article">
      <header className="article-topbar">
        <button type="button" className="back-link back-link--strong" onClick={onBack}>
          <span className="back-link__arrow" aria-hidden="true">
            ←
          </span>
          <span>Назад ко всем новостям</span>
        </button>
        {selectedNews ? (
          <div className="article-topbar__meta">
            <span>{selectedNews.SourceName}</span>
          </div>
        ) : null}
      </header>

      <section className="article-hero">
        <div className="article-intro">
          <p className="eyebrow">Выбранная новость</p>
          <h1>Чтение новости в выбранной подаче</h1>
        </div>

        <aside className="article-moods-card">
          <span className="panel__label">Режим</span>
          <h2 className="article-moods-card__title">Выбери интонацию</h2>
          <div className="moods moods--article moods--compact">
            {moods.map((mood) => (
              <button
                key={mood.id}
                type="button"
                className={!showOriginal && mood.id === activeMood ? "mood mood--active" : "mood"}
                onClick={() => {
                  setShowOriginal(false);
                  setActiveMood(mood.id);
                }}
                disabled={isLoading || isSyncing}
              >
                <span className="mood__emoji" aria-hidden="true">
                  {mood.emoji}
                </span>
                <span>{mood.label}</span>
              </button>
            ))}
            <button
              type="button"
              className={showOriginal ? "mood mood--secondary mood--active-alt" : "mood mood--secondary"}
              onClick={() => setShowOriginal(true)}
              disabled={isLoading || isSyncing}
            >
              <span className="mood__emoji" aria-hidden="true">
                ↺
              </span>
              <span>Сбросить</span>
            </button>
          </div>
          <p className="status-text">{status}</p>
          {error ? <p className="error-text">{error}</p> : null}
        </aside>
      </section>

      <main className="article-layout">
        <NewsDetail
          activeMood={activeMood}
          isLoading={isBootstrapping || isDetailLoading}
          rewriteFallbackMessage={rewriteFallbackMessage}
          selectedNews={selectedNews}
          showOriginal={showOriginal}
        />
      </main>
    </div>
  );
}
