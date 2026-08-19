import { useEffect, useState } from "react";
import { NewsArticlePage } from "./pages/NewsArticlePage";
import { NewsFeedPage } from "./pages/NewsFeedPage";
import { useNewsPage } from "./hooks/useNewsPage";

const articlePathPrefix = "/news/";

function getRoute(pathname) {
  if (pathname.startsWith(articlePathPrefix)) {
    const rawExternalID = pathname.slice(articlePathPrefix.length);
    if (rawExternalID) {
      try {
        const externalID = decodeURIComponent(rawExternalID);
        if (externalID) {
          return {
            name: "article",
            externalID
          };
        }
      } catch {
        return { name: "feed", externalID: "" };
      }
    }
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
    newsPage.setSelectedId(route.name === "article" ? route.externalID : "");
  }, [route.name, route.externalID]);

  function navigate(pathname) {
    if (pathname === window.location.pathname) {
      return;
    }

    window.history.pushState({}, "", pathname);
    setRoute(getRoute(pathname));
  }

  function openNews(externalID) {
    navigate(`${articlePathPrefix}${encodeURIComponent(externalID)}`);
  }

  function openFeed() {
    navigate("/");
  }

  if (route.name === "article") {
    return <NewsArticlePage {...newsPage} onBack={openFeed} onOpenNews={openNews} />;
  }

  return <NewsFeedPage {...newsPage} onOpenNews={openNews} />;
}
