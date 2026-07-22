import { useEffect, useMemo, useState } from "react";
import { getNameMatchTier } from "../utils/nameMatch.js";

const DEFAULT_SORT = "price-asc";

export default function useResultFilters(results, searchQuery) {
  const [sortBy, setSortBy] = useState(DEFAULT_SORT);
  const [qualityFilter, setQualityFilter] = useState("all");
  const [foilOnly, setFoilOnly] = useState(false);
  const [cheapestPerStore, setCheapestPerStore] = useState(false);
  const [storeFilter, setStoreFilter] = useState(null);

  // Reset filters when the user runs a new search.
  // biome-ignore lint/correctness/useExhaustiveDependencies: searchQuery triggers filter reset on new search
  useEffect(() => {
    setSortBy(DEFAULT_SORT);
    setQualityFilter("all");
    setFoilOnly(false);
    setCheapestPerStore(false);
    setStoreFilter(null);
  }, [searchQuery]);

  const availableStores = useMemo(() => {
    const stores = new Set();
    for (const card of results) {
      if (card.src) {
        stores.add(card.src);
      }
    }
    return [...stores].sort();
  }, [results]);

  const selectedStores = storeFilter ?? availableStores;

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

    if (storeFilter !== null) {
      const allowedStores = new Set(storeFilter);
      filtered = filtered.filter((card) => allowedStores.has(card.src));
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
  }, [
    results,
    sortBy,
    qualityFilter,
    foilOnly,
    storeFilter,
    cheapestPerStore,
    searchQuery,
  ]);

  const isStoreFilterActive =
    storeFilter !== null && storeFilter.length < availableStores.length;

  const hasActiveFilters =
    qualityFilter !== "all" ||
    foilOnly ||
    cheapestPerStore ||
    isStoreFilterActive;

  const toggleStoreFilter = (store) => {
    const current = storeFilter ?? availableStores;
    const next = current.includes(store)
      ? current.filter((name) => name !== store)
      : [...current, store];

    if (
      next.length === availableStores.length &&
      availableStores.every((name) => next.includes(name))
    ) {
      setStoreFilter(null);
      return;
    }

    setStoreFilter(next);
  };

  const selectAllStores = () => {
    setStoreFilter(null);
  };

  const selectNoStores = () => {
    setStoreFilter([]);
  };

  const clearFilters = () => {
    setQualityFilter("all");
    setFoilOnly(false);
    setCheapestPerStore(false);
    setStoreFilter(null);
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
    availableStores,
    selectedStores,
    toggleStoreFilter,
    selectAllStores,
    selectNoStores,
    isStoreFilterActive,
    hasActiveFilters,
    clearFilters,
  };
}
