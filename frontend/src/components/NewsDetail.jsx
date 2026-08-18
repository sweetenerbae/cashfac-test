import { formatDate } from "../utils/news";

export function NewsDetail({ isLoading, selectedNews }) {
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
          Ниже можно сравнить исходную подачу материала и версию с выбранной эмоциональной интонацией, не меняя сами
          факты.
        </p>
        <a href={selectedNews.SourceURL} target="_blank" rel="noreferrer">
          Читать оригинал на сайте источника
        </a>
      </div>

      <div className="compare">
        <section className="compare__panel compare__panel--original">
          <div className="compare__panel-head">
            <span className="compare__eyebrow">Исходный текст</span>
            <h3>Оригинал</h3>
          </div>
          <div className="article-text">
            <p>{selectedNews.OriginalText}</p>
          </div>
        </section>

        <section className="compare__panel compare__panel--rewrite">
          <div className="compare__panel-head">
            <span className="compare__eyebrow">Новая подача</span>
            <h3>Рерайт</h3>
          </div>
          <div className="article-text">
            <p>{selectedNews.RewrittenText}</p>
          </div>
        </section>
      </div>
    </article>
  );
}
