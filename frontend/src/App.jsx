import { useEffect, useState } from "react";
import { NewsArticlePage } from "./pages/NewsArticlePage";
import { NewsFeedPage } from "./pages/NewsFeedPage";
import { useNewsPage } from "./hooks/useNewsPage";

const articlePathPrefix = "/news/";

function getRoute(pathname) {
  if (pathname.startsWith(articlePathPrefix)) {
    const rawExternalID = pathname.slice(articlePathPrefix.length);
    return {
      name: "article",
      externalID: decodeURIComponent(rawExternalID)
    };
  }

  return {
    name: "feed",
    externalID: ""
  };
}

export default function App() {
  const newsPage = useNewsPage();
  const [route, setRoute] = useState(() => getRoute(window.location.pathname));

  useEffect(() => {
    function handlePopState() {
      setRoute(getRoute(window.location.pathname));
    }

    window.addEventListener("popstate", handlePopState);
    return () => {
      window.removeEventListener("popstate", handlePopState);
    };
  }, []);

  useEffect(() => {
    if (route.name !== "article" || !route.externalID) {
      return;
    }

    newsPage.setSelectedId(route.externalID);
  }, [route.name, route.externalID]);

  function navigate(pathname) {
    if (pathname === window.location.pathname) {
      return;
    }

    window.history.pushState({}, "", pathname);
    setRoute(getRoute(pathname));
  }

  function openNews(externalID) {
    newsPage.setSelectedId(externalID);
    navigate(`${articlePathPrefix}${encodeURIComponent(externalID)}`);
  }

  function openFeed() {
    newsPage.setSelectedId("");
    navigate("/");
  }

  if (route.name === "article") {
    return <NewsArticlePage {...newsPage} onBack={openFeed} onOpenNews={openNews} />;
  }

  return <NewsFeedPage {...newsPage} onOpenNews={openNews} />;
}
