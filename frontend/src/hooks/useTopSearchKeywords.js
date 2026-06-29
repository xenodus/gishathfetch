import { useEffect, useState } from "react";
import {
  TOP_SEARCH_KEYWORDS_DISPLAY_LIMIT,
  TOP_SEARCH_KEYWORDS_URL,
} from "../constants";

function parseKeywordList(keywords) {
  if (!Array.isArray(keywords)) {
    return [];
  }

  return keywords
    .map((item) => (typeof item?.term === "string" ? item.term.trim() : ""))
    .filter(Boolean)
    .slice(0, TOP_SEARCH_KEYWORDS_DISPLAY_LIMIT);
}

function parseTopSearchKeywords(payload) {
  return {
    last24Hours: parseKeywordList(payload?.periods?.last24Hours?.keywords),
    last30Days: parseKeywordList(payload?.periods?.last30Days?.keywords),
    last6Months: parseKeywordList(payload?.periods?.last6Months?.keywords),
    last1Year: parseKeywordList(payload?.periods?.last1Year?.keywords),
  };
}

const EMPTY_KEYWORDS_BY_PERIOD = {
  last24Hours: [],
  last30Days: [],
  last6Months: [],
  last1Year: [],
};

export default function useTopSearchKeywords(enabled) {
  const [keywordsByPeriod, setKeywordsByPeriod] = useState(
    EMPTY_KEYWORDS_BY_PERIOD,
  );
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
          setKeywordsByPeriod(parseTopSearchKeywords(payload));
        }
      } catch (error) {
        if (!cancelled && error.name !== "AbortError") {
          console.error("Failed to load top search keywords:", error);
          setKeywordsByPeriod(EMPTY_KEYWORDS_BY_PERIOD);
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

  return { keywordsByPeriod, isLoading };
}
