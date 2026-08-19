import { NewsImage } from "./NewsImage";
import { excerpt, formatDate } from "../utils/news";

export function NewsCard({ isActive, item, onOpen }) {
  return (
    <article
      className={isActive ? "card card--active" : "card"}
      onClick={() => onOpen(item.ExternalID)}
      role="button"
      tabIndex={0}
      onKeyDown={(event) => {
        if (event.key === "Enter" || event.key === " ") {
          event.preventDefault();
          onOpen(item.ExternalID);
        }
      }}
    >
      <NewsImage className="card__media" imageURL={item.ImageURL} title={item.Title} />
      <span className="card__source">{item.SourceName}</span>
      <h2>{item.Title}</h2>
      <p>{excerpt(item.OriginalText)}</p>
      <div className="card__meta">
        <span>{formatDate(item.PublishedAt)}</span>
        <a
          href={item.SourceURL}
          target="_blank"
          rel="noreferrer"
          onClick={(event) => event.stopPropagation()}
          onKeyDown={(event) => event.stopPropagation()}
        >
          Открыть оригинальный источник
        </a>
      </div>
      <div className="card__footer">
        <span className="card__link">Открыть сравнение</span>
      </div>
    </article>
  );
}
