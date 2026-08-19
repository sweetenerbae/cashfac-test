import { useEffect, useState } from "react";

export function NewsImage({ className = "", imageURL, loading = "lazy", title }) {
  const [hasError, setHasError] = useState(false);

  useEffect(() => {
    setHasError(false);
  }, [imageURL]);

  const classes = ["news-image", className, !imageURL || hasError ? "news-image--placeholder" : ""]
    .filter(Boolean)
    .join(" ");

  return (
    <div className={classes}>
      {imageURL && !hasError ? (
        <img
          src={imageURL}
          alt={`Иллюстрация к новости «${title}»`}
          loading={loading}
          decoding="async"
          onError={() => setHasError(true)}
        />
      ) : null}
    </div>
  );
}
