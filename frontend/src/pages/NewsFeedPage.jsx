import { NewsCard } from "../components/NewsCard";
import { NewsSkeletonGrid } from "../components/NewsSkeletonGrid";

export function NewsFeedPage({
  error,
  isBootstrapping,
  isLoading,
  isSyncing,
  news,
  onOpenNews,
  status,
  syncFeedback,
  syncNews
}) {
  const shouldShowStatus = isBootstrapping || isLoading || isSyncing || Boolean(error);

  return (
    <div className="page">
      <header className="hero hero--feed">
        <div className="hero__copy">
          <p className="eyebrow">Тестовое задание Cash Factories</p>
          <h1>Новости с разным настроением подачи.</h1>
          <p className="hero__text">
            В ленте лежат реальные публикации из The Guardian. На карточке можно открыть новость и сравнить
            исходный текст с версией, переписанной в выбранной интонации.
          </p>
        </div>

        <div className="hero__controls">
          <div className="hero__actions">
            <button
              type="button"
              className="action-button"
              onClick={() => void syncNews()}
              disabled={isSyncing}
            >
              {isSyncing ? "Загружаю..." : "Загрузить свежие новости"}
            </button>
          </div>

          {shouldShowStatus ? <p className="status-text">{status}</p> : null}
          {syncFeedback ? <p className="status-text status-text--result">{syncFeedback}</p> : null}
          {error ? <p className="error-text">{error}</p> : null}
        </div>
      </header>

      <main className="news-grid news-grid--full">
        {isBootstrapping || isLoading ? (
          <NewsSkeletonGrid />
        ) : news.length === 0 ? (
          <div className="empty-state empty-state--wide">
            <h2>Новостей пока нет</h2>
            <p>Сделай первую синхронизацию, и здесь появятся реальные публикации из The Guardian.</p>
          </div>
        ) : (
          news.map((item) => (
            <NewsCard
              key={item.ID}
              isActive={false}
              item={item}
              onOpen={onOpenNews}
            />
          ))
        )}
      </main>
    </div>
  );
}
