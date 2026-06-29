import { useCallback, useEffect, useState } from "react";
import { AFFILIATE_LINKS_URL } from "../constants";

const EMPTY_LINKS = [];

function parseAffiliateLinks(payload) {
  if (!Array.isArray(payload?.data)) {
    return EMPTY_LINKS;
  }
  return payload.data.filter(
    (item) =>
      typeof item?.link === "string" &&
      typeof item?.imageUrl === "string" &&
      typeof item?.price === "string" &&
      (item.platform === undefined || typeof item?.platform === "string"),
  );
}

export default function useAffiliateLinks(enabled = true) {
  const [links, setLinks] = useState(EMPTY_LINKS);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");

  const reload = useCallback(async (signal) => {
    setIsLoading(true);
    setError("");
    try {
      const response = await fetch(AFFILIATE_LINKS_URL, { signal });
      if (!response.ok) {
        throw new Error(`Failed to load affiliate links (${response.status})`);
      }
      const payload = await response.json();
      setLinks(parseAffiliateLinks(payload));
    } catch (err) {
      if (err.name !== "AbortError") {
        console.error("Failed to load affiliate links:", err);
        setLinks(EMPTY_LINKS);
        setError("Could not load featured products.");
      }
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    if (!enabled) {
      return undefined;
    }

    const controller = new AbortController();
    reload(controller.signal);
    return () => controller.abort();
  }, [enabled, reload]);

  return { links, isLoading, error };
}
