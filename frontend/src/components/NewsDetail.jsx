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
        <h3>{selectedNews.Title}</h3>
        <p>
          {selectedNews.SourceName} · {formatDate(selectedNews.PublishedAt)}
        </p>
        <a href={selectedNews.SourceURL} target="_blank" rel="noreferrer">
          Читать оригинал на сайте источника
        </a>
      </div>

      <div className="compare">
        <section className="compare__panel">
          <h3>Оригинал</h3>
          <div className="article-text">
            <p>{selectedNews.OriginalText}</p>
          </div>
        </section>

        <section className="compare__panel">
          <h3>Рерайт</h3>
          <div className="article-text">
            <p>{selectedNews.RewrittenText}</p>
          </div>
        </section>
      </div>
    </article>
  );
}
