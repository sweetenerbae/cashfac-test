import { excerpt, formatDate } from "../utils/news";

export function NewsCard({ isActive, item, onSelect }) {
  return (
    <article
      className={isActive ? "card card--active" : "card"}
      onClick={() => onSelect(item.ID)}
      role="button"
      tabIndex={0}
      onKeyDown={(event) => {
        if (event.key === "Enter" || event.key === " ") {
          event.preventDefault();
          onSelect(item.ID);
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
  );
}
