export function NewsSkeletonGrid() {
  return Array.from({ length: 6 }).map((_, index) => (
    <article key={`skeleton-${index}`} className="card card--skeleton" aria-hidden="true">
      <div className="skeleton skeleton--card-media" />
      <div className="skeleton skeleton--eyebrow" />
      <div className="skeleton skeleton--title" />
      <div className="skeleton skeleton--title skeleton--title-short" />
      <div className="skeleton skeleton--text" />
      <div className="skeleton skeleton--text" />
      <div className="skeleton skeleton--text skeleton--text-short" />
    </article>
  ));
}
