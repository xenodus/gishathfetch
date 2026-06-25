import { useEffect, useMemo, useState } from "react";

const DEFAULT_SORT = "price-asc";

export default function useResultFilters(results, searchQuery) {
  const [sortBy, setSortBy] = useState(DEFAULT_SORT);
  const [qualityFilter, setQualityFilter] = useState("all");
  const [foilOnly, setFoilOnly] = useState(false);

  // Reset filters when the user runs a new search.
  // biome-ignore lint/correctness/useExhaustiveDependencies: searchQuery triggers filter reset on new search
  useEffect(() => {
    setSortBy(DEFAULT_SORT);
    setQualityFilter("all");
    setFoilOnly(false);
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

    filtered.sort((a, b) => {
      switch (sortBy) {
        case "price-desc":
          return b.price - a.price;
        case "store-asc":
          return a.src.localeCompare(b.src) || a.price - b.price;
        default:
          return a.price - b.price;
      }
    });

    return filtered;
  }, [results, sortBy, qualityFilter, foilOnly]);

  const hasActiveFilters = qualityFilter !== "all" || foilOnly;

  const clearFilters = () => {
    setQualityFilter("all");
    setFoilOnly(false);
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
    hasActiveFilters,
    clearFilters,
  };
}
