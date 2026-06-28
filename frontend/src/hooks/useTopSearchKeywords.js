import { useEffect, useState } from "react";
import {
  TOP_SEARCH_KEYWORDS_DISPLAY_LIMIT,
  TOP_SEARCH_KEYWORDS_URL,
} from "../constants";

function parseTopSearchKeywords(payload) {
  const keywords = payload?.periods?.last24Hours?.keywords;
  if (!Array.isArray(keywords)) {
    return [];
  }

  return keywords
    .map((item) => (typeof item?.term === "string" ? item.term.trim() : ""))
    .filter(Boolean)
    .slice(0, TOP_SEARCH_KEYWORDS_DISPLAY_LIMIT);
}

export default function useTopSearchKeywords(enabled) {
  const [keywords, setKeywords] = useState([]);
  const [isLoading, setIsLoading] = useState(false);

  useEffect(() => {
    if (!enabled) {
      return;
    }

    const controller = new AbortController();
    let cancelled = false;

    const loadKeywords = async () => {
      setIsLoading(true);
      try {
        const response = await fetch(TOP_SEARCH_KEYWORDS_URL, {
          signal: controller.signal,
        });
        if (!response.ok) {
          throw new Error(
            `Failed to load top search keywords (${response.status})`,
          );
        }

        const payload = await response.json();
        if (!cancelled) {
          setKeywords(parseTopSearchKeywords(payload));
        }
      } catch (error) {
        if (!cancelled && error.name !== "AbortError") {
          console.error("Failed to load top search keywords:", error);
          setKeywords([]);
        }
      } finally {
        if (!cancelled) {
          setIsLoading(false);
        }
      }
    };

    loadKeywords();

    return () => {
      cancelled = true;
      controller.abort();
    };
  }, [enabled]);

  return { keywords, isLoading };
}
