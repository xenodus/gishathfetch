import { useEffect, useMemo, useState } from "react";
import { getNameMatchTier } from "../utils/nameMatch.js";

const DEFAULT_SORT = "price-asc";

export default function useResultFilters(results, searchQuery) {
  const [sortBy, setSortBy] = useState(DEFAULT_SORT);
  const [qualityFilter, setQualityFilter] = useState("all");
  const [foilOnly, setFoilOnly] = useState(false);
  const [cheapestPerStore, setCheapestPerStore] = useState(false);

  // Reset filters when the user runs a new search.
  // biome-ignore lint/correctness/useExhaustiveDependencies: searchQuery triggers filter reset on new search
  useEffect(() => {
    setSortBy(DEFAULT_SORT);
    setQualityFilter("all");
    setFoilOnly(false);
    setCheapestPerStore(false);
  }, [searchQuery]);

  const availableQualities = useMemo(() => {
    const qualities = new Set();
    for (const card of results) {
      if (card.quality) {
        qualities.add(card.quality);
      }
    }
    return [...qualities].sort();
  }, [results]);

  const filteredResults = useMemo(() => {
    let filtered = [...results];

    if (qualityFilter !== "all") {
      filtered = filtered.filter((card) => card.quality === qualityFilter);
    }

    if (foilOnly) {
      filtered = filtered.filter((card) => card.isFoil);
    }

    if (cheapestPerStore) {
      const cheapestByStore = new Map();
      for (const card of filtered) {
        const storeName = card.src || "Unknown Store";
        const existing = cheapestByStore.get(storeName);
        if (!existing || card.price < existing.price) {
          cheapestByStore.set(storeName, card);
        }
      }
      filtered = [...cheapestByStore.values()];
    }

    filtered.sort((a, b) => {
      const tierDiff =
        getNameMatchTier(a.name, searchQuery) -
        getNameMatchTier(b.name, searchQuery);
      if (tierDiff !== 0) {
        return tierDiff;
      }

      if (sortBy === "price-desc") {
        return b.price - a.price;
      }
      return a.price - b.price;
    });

    return filtered;
  }, [results, sortBy, qualityFilter, foilOnly, cheapestPerStore, searchQuery]);

  const hasActiveFilters =
    qualityFilter !== "all" || foilOnly || cheapestPerStore;

  const clearFilters = () => {
    setQualityFilter("all");
    setFoilOnly(false);
    setCheapestPerStore(false);
  };

  return {
    filteredResults,
    sortBy,
    setSortBy,
    qualityFilter,
    setQualityFilter,
    availableQualities,
    foilOnly,
    setFoilOnly,
    cheapestPerStore,
    setCheapestPerStore,
    hasActiveFilters,
    clearFilters,
  };
}
