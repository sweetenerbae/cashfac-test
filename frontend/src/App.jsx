const moods = [
  { id: "neutral", label: "Нейтрально" },
  { id: "happy", label: "Радостно" },
  { id: "sad", label: "Грустно" },
  { id: "ironic", label: "Иронично" }
];

const previewNews = [
  {
    id: "1",
    title: "Здесь будет заголовок новости",
    source: "Источник",
    origin: "Превью исходного текста",
    rewritten: "Превью переписанного текста",
    link: "#"
  },
  {
    id: "2",
    title: "Вторая карточка новости для сетки",
    source: "Источник",
    origin: "Превью исходного текста",
    rewritten: "Превью переписанного текста",
    link: "#"
  },
  {
    id: "3",
    title: "Еще одна карточка с длинным заголовком в несколько строк",
    source: "Источник",
    origin: "Превью исходного текста",
    rewritten: "Превью переписанного текста",
    link: "#"
  }
];

function App() {
  return (
    <div className="page">
      <header className="hero">
        <div className="hero__copy">
          <p className="eyebrow">Тестовое задание Cash Factories</p>
          <h1>Новости с переключением эмоционального режима.</h1>
          <p className="hero__text">
            Каркас клиентской части для грида реальных новостей, выбора
            настроения и сравнения исходного текста с переписанной версией.
          </p>
        </div>

        <div className="hero__panel">
          <span className="panel__label">Режим</span>
          <div className="moods">
            {moods.map((mood) => (
              <button key={mood.id} className={mood.id === "neutral" ? "mood mood--active" : "mood"}>
                {mood.label}
              </button>
            ))}
          </div>
        </div>
      </header>

      <main className="layout">
        <section className="news-grid">
          {previewNews.map((item) => (
            <article key={item.id} className="card">
              <span className="card__source">{item.source}</span>
              <h2>{item.title}</h2>
              <p>{item.origin}</p>
              <a href={item.link}>Открыть оригинальный источник</a>
            </article>
          ))}
        </section>

        <aside className="detail">
          <div className="detail__head">
            <span className="detail__eyebrow">Выбранная новость</span>
            <h2>Сравнение исходного и переписанного текста</h2>
          </div>

          <div className="compare">
            <section>
              <h3>Оригинал</h3>
              <p>
                После выбора новости здесь будет показан исходный текст статьи.
              </p>
            </section>

            <section>
              <h3>Рерайт</h3>
              <p>
                После выбора новости здесь будет показан текст в выбранном
                эмоциональном режиме.
              </p>
            </section>
          </div>
        </aside>
      </main>
    </div>
  );
}

export default App;
