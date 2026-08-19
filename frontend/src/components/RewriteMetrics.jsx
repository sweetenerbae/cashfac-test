const sourceLabels = {
  generated: "Новая генерация",
  cache: "Из кеша backend",
  shared: "Общий параллельный запрос",
  local: "Из кеша браузера"
};

export function RewriteMetrics({ meta }) {
  if (!meta) {
    return null;
  }

  return (
    <aside className="rewrite-metrics" aria-label="Статистика использования языковой модели">
      <span className="rewrite-metrics__label">Экономия LLM</span>
      <div className="rewrite-metrics__items">
        <strong>{sourceLabels[meta.Source] || "Рерайт готов"}</strong>
        <span>{formatRequests(meta.LLMRequests)}</span>
        {meta.SavedLLMRequests > 0 ? (
          <span className="rewrite-metrics__saved">Сэкономлено: {formatSavedRequests(meta.SavedLLMRequests)}</span>
        ) : null}
        {meta.DurationMs > 0 ? <span>{formatDuration(meta.DurationMs)}</span> : null}
      </div>
    </aside>
  );
}

function formatRequests(count) {
  return count === 1 ? "1 запрос к LLM" : `${count} запросов к LLM`;
}

function formatSavedRequests(count) {
  return count === 1 ? "1 запрос" : `${count} запроса`;
}

function formatDuration(durationMs) {
  if (durationMs < 1000) {
    return `${durationMs} мс`;
  }

  return `${(durationMs / 1000).toFixed(1)} с`;
}
