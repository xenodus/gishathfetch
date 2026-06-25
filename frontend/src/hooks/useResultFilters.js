import { useEffect, useMemo, useState } from "react";

const DEFAULT_SORT = "price-asc";

export default function useResultFilters(results, searchQuery) {
  const [sortBy, setSortBy] = useState(DEFAULT_SORT);
  const [foilOnly, setFoilOnly] = useState(false);

  // Reset filters when the user runs a new search.
  // biome-ignore lint/correctness/useExhaustiveDependencies: searchQuery triggers filter reset on new search
  useEffect(() => {
    setSortBy(DEFAULT_SORT);
    setFoilOnly(false);
  }, [searchQuery]);

  const filteredResults = useMemo(() => {
    const filtered = foilOnly
      ? results.filter((card) => card.isFoil)
      : [...results];

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
  }, [results, sortBy, foilOnly]);

  const clearFilters = () => {
    setFoilOnly(false);
  };

  return {
    filteredResults,
    sortBy,
    setSortBy,
    foilOnly,
    setFoilOnly,
    hasActiveFilters: foilOnly,
    clearFilters,
  };
}
