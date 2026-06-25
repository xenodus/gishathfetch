import { useEffect, useMemo, useState } from "react";

const DEFAULT_SORT = "price-asc";

export default function useResultFilters(results, searchQuery) {
  const [sortBy, setSortBy] = useState(DEFAULT_SORT);
  const [foilOnly, setFoilOnly] = useState(false);
  const [selectedStores, setSelectedStores] = useState(null);
  const [qualityFilter, setQualityFilter] = useState("all");
  const [priceMin, setPriceMin] = useState("");
  const [priceMax, setPriceMax] = useState("");

  // Reset filters when the user runs a new search.
  // biome-ignore lint/correctness/useExhaustiveDependencies: searchQuery triggers filter reset on new search
  useEffect(() => {
    setSortBy(DEFAULT_SORT);
    setFoilOnly(false);
    setSelectedStores(null);
    setQualityFilter("all");
    setPriceMin("");
    setPriceMax("");
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

    if (foilOnly) {
      filtered = filtered.filter((card) => card.isFoil);
    }

    if (selectedStores !== null) {
      filtered = filtered.filter((card) => selectedStores.includes(card.src));
    }

    if (qualityFilter !== "all") {
      filtered = filtered.filter((card) => card.quality === qualityFilter);
    }

    const min = priceMin !== "" ? Number.parseFloat(priceMin) : null;
    const max = priceMax !== "" ? Number.parseFloat(priceMax) : null;
    if (min !== null && !Number.isNaN(min)) {
      filtered = filtered.filter((card) => card.price >= min);
    }
    if (max !== null && !Number.isNaN(max)) {
      filtered = filtered.filter((card) => card.price <= max);
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
  }, [
    results,
    sortBy,
    foilOnly,
    selectedStores,
    qualityFilter,
    priceMin,
    priceMax,
  ]);

  const hasActiveFilters =
    foilOnly ||
    selectedStores !== null ||
    qualityFilter !== "all" ||
    priceMin !== "" ||
    priceMax !== "";

  const clearFilters = () => {
    setFoilOnly(false);
    setSelectedStores(null);
    setQualityFilter("all");
    setPriceMin("");
    setPriceMax("");
  };

  const toggleStore = (store) => {
    setSelectedStores((prev) => {
      const current = prev ?? availableStores;
      const next = current.includes(store)
        ? current.filter((s) => s !== store)
        : [...current, store];
      if (next.length === availableStores.length) {
        return null;
      }
      return next;
    });
  };

  const isStoreSelected = (store) => {
    const current = selectedStores ?? availableStores;
    return current.includes(store);
  };

  return {
    filteredResults,
    sortBy,
    setSortBy,
    foilOnly,
    setFoilOnly,
    qualityFilter,
    setQualityFilter,
    priceMin,
    setPriceMin,
    priceMax,
    setPriceMax,
    availableStores,
    availableQualities,
    toggleStore,
    isStoreSelected,
    hasActiveFilters,
    clearFilters,
  };
}
