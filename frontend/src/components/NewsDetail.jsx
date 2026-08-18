import { defaultMood, getMoodLabel } from "../constants/moods";
import { formatDate } from "../utils/news";

export function NewsDetail({ activeMood, isLoading, rewriteFallbackMessage, selectedNews, showOriginal }) {
  if (isLoading) {
    return (
      <div className="detail detail--article detail-skeleton" aria-hidden="true">
        <div className="skeleton skeleton--detail-title" />
        <div className="skeleton skeleton--detail-meta" />
        <div className="compare compare--detail-loading">
          <section>
            <div className="skeleton skeleton--section-title" />
            <div className="skeleton skeleton--paragraph" />
            <div className="skeleton skeleton--paragraph" />
            <div className="skeleton skeleton--paragraph skeleton--paragraph-short" />
          </section>
          <section>
            <div className="skeleton skeleton--section-title" />
            <div className="skeleton skeleton--paragraph" />
            <div className="skeleton skeleton--paragraph" />
            <div className="skeleton skeleton--paragraph skeleton--paragraph-short" />
          </section>
        </div>
      </div>
    );
  }

  if (!selectedNews) {
    return (
      <div className="empty-state empty-state--detail empty-state--wide">
        <h3>Пока нечего показывать</h3>
        <p>После загрузки новостей выбери карточку, и здесь появится подробное сравнение.</p>
      </div>
    );
  }

  const currentTitle = showOriginal ? "Оригинал" : activeMood === defaultMood ? "Нейтральная версия" : "Рерайт";
  const currentEyebrow = showOriginal ? "Исходный текст" : "Текущая подача";
  const currentText = showOriginal ? selectedNews.OriginalText : selectedNews.RewrittenText;

  return (
    <article className="detail detail--article">
      <div className="detail__summary">
        <div className="detail__summary-meta">
          <span className="detail__badge">Оригинальный материал</span>
          <span className="detail__divider" aria-hidden="true">
            /
          </span>
          <span className="detail__source-line">
            {selectedNews.SourceName} · {formatDate(selectedNews.PublishedAt)}
          </span>
        </div>
        <h3>{selectedNews.Title}</h3>
        <p className="detail__lead">
          Здесь показана версия материала в выбранной подаче. Оригинальный текст можно открыть отдельной кнопкой,
          не теряя саму новость и факты.
        </p>
        <div className="detail__actions">
          <a href={selectedNews.SourceURL} target="_blank" rel="noreferrer">
            Читать оригинал на сайте источника
          </a>
        </div>
      </div>

      <section className={showOriginal ? "compare__panel compare__panel--original" : "compare__panel compare__panel--rewrite"}>
        <div className="compare__panel-head">
          <span className="compare__eyebrow">{currentEyebrow}</span>
          <h3>{currentTitle}</h3>
        </div>
        {!showOriginal && activeMood !== defaultMood && rewriteFallbackMessage ? (
          <div className="rewrite-fallback">
            <p className="rewrite-fallback__title">Режим «{getMoodLabel(activeMood)}» пока недоступен для этой статьи</p>
            <p>{rewriteFallbackMessage}</p>
          </div>
        ) : (
          <div className="article-text">
            <p>{currentText}</p>
          </div>
        )}
      </section>
    </article>
  );
}
