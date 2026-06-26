import { useCallback, useEffect, useRef, useState } from "react";
import {
  API_BASE_URL,
  BASE_URL,
  LGS_OPTIONS,
  MAX_SEARCH_LENGTH,
  MIN_SEARCH_LENGTH,
} from "../constants";
import {
  buildSearchUrl,
  getInitialSelectedStores,
  getStoresFromUrl,
  persistSelectedStores,
} from "../utils/searchUrl";

const SEARCH_TOO_LONG_ERROR = `Card name is too long (maximum ${MAX_SEARCH_LENGTH} characters).`;
const SEARCH_TOO_SHORT_ERROR = `Enter at least ${MIN_SEARCH_LENGTH} characters to search.`;
const NO_STORES_WARNING =
  "No stores selected — searching all stores. Select specific stores to search faster.";

function formatErrorWithStatusCode(message, statusCode) {
  if (!statusCode) {
    return message;
  }
  return `Error (${statusCode}): ${message}`;
}

// Search configuration constants
const AUTOCOMPLETE_DEBOUNCE_MS = 300;
const SEARCH_PROGRESS_INTERVAL_MS = 1000;
const MAX_PROGRESS_DOTS = 15;

export default function useSearch() {
  const [searchQuery, setSearchQuery] = useState(() => {
    const urlParams = new URLSearchParams(window.location.search);
    if (urlParams.has("s") && urlParams.get("s") !== "") {
      return decodeURIComponent(urlParams.get("s"));
    }
    return "";
  });
  const [isSearching, setIsSearching] = useState(false);
  const [searchResults, setSearchResults] = useState([]);
  const [hasSearched, setHasSearched] = useState(false);
  const [searchProgress, setSearchProgress] = useState("Search");
  const [suggestions, setSuggestions] = useState([]);
  const [showSuggestions, setShowSuggestions] = useState(false);
  const [searchError, setSearchError] = useState(null);
  const [searchStoreErrors, setSearchStoreErrors] = useState([]);
  const [dismissedStoreErrorsKey, setDismissedStoreErrorsKey] = useState(null);
  const [storesWarning, setStoresWarning] = useState(null);
  const [selectedStores, setSelectedStores] = useState(() =>
    getInitialSelectedStores(),
  );

  // --- Helpers ---
  const skipSuggestionsRef = useRef(
    new URLSearchParams(window.location.search).has("s"),
  );
  const progressIntervalRef = useRef(null);
  const searchAbortControllerRef = useRef(null);
  const autocompleteAbortControllerRef = useRef(null);
  const activeSearchRequestIdRef = useRef(0);
  const searchResultsRef = useRef([]);
  const resultsBeforeSearchRef = useRef([]);
  const userCancelledRef = useRef(false);
  const lastSearchRef = useRef({ query: "", stores: [] });

  useEffect(() => {
    searchResultsRef.current = searchResults;
  }, [searchResults]);

  const updateUrlAndTitle = useCallback((query, stores) => {
    if (window.location.hostname !== "localhost") {
      const newUrl = buildSearchUrl(BASE_URL, query, stores);
      window.history.pushState(
        { query, stores },
        `${query} @ Gishath Fetch`,
        newUrl,
      );
      document.title = `${query} @ Gishath Fetch`;
    }
  }, []);

  const resolveStoresToSearch = useCallback((stores) => {
    if (stores.length > 0) {
      setStoresWarning(null);
      return stores;
    }

    setStoresWarning(NO_STORES_WARNING);
    setSelectedStores(LGS_OPTIONS);
    persistSelectedStores(LGS_OPTIONS);
    return LGS_OPTIONS;
  }, []);

  const performSearch = useCallback(
    (query, stores) => {
      if (!query || query.length < MIN_SEARCH_LENGTH) {
        setSearchError(SEARCH_TOO_SHORT_ERROR);
        setHasSearched(false);
        return;
      }

      if (query.length > MAX_SEARCH_LENGTH) {
        setHasSearched(true);
        setSearchResults([]);
        setSearchError(SEARCH_TOO_LONG_ERROR);
        return;
      }

      if (progressIntervalRef.current) {
        clearInterval(progressIntervalRef.current);
        progressIntervalRef.current = null;
      }
      if (searchAbortControllerRef.current) {
        searchAbortControllerRef.current.abort();
      }

      const requestId = activeSearchRequestIdRef.current + 1;
      activeSearchRequestIdRef.current = requestId;
      const searchAbortController = new AbortController();
      searchAbortControllerRef.current = searchAbortController;

      // Prevent suggestions from appearing after programmatic search
      skipSuggestionsRef.current = true;
      userCancelledRef.current = false;
      resultsBeforeSearchRef.current = searchResultsRef.current;
      lastSearchRef.current = { query, stores };

      setIsSearching(true);
      setSearchProgress("Searching LGS");
      setSearchResults([]);
      setHasSearched(true);
      setSearchError(null);
      setSearchStoreErrors([]);
      setDismissedStoreErrorsKey(null);

      if (window.gtag) {
        window.gtag("event", "search", { search_term: query });
      }

      const searchUrl = `${API_BASE_URL}?s=${encodeURIComponent(query)}&lgs=${encodeURIComponent(stores.join(","))}`;

      const progressInterval = setInterval(() => {
        setSearchProgress((prev) => {
          const dots = (prev.match(/\./g) || []).length;
          if (dots >= MAX_PROGRESS_DOTS) return "Searching LGS";
          return `${prev} .`;
        });
      }, SEARCH_PROGRESS_INTERVAL_MS);
      progressIntervalRef.current = progressInterval;

      fetch(searchUrl, { signal: searchAbortController.signal })
        .then(async (res) => {
          if (!res.ok) {
            let errorBody = null;
            try {
              errorBody = await res.json();
            } catch {
              // Ignore malformed error responses.
            }

            const validationMessage =
              typeof errorBody?.error === "string" && errorBody.error
                ? errorBody.error
                : null;
            const statusCode = errorBody?.statusCode || res.status;

            if (validationMessage) {
              throw new Error(
                formatErrorWithStatusCode(validationMessage, statusCode),
              );
            }

            throw new Error(
              formatErrorWithStatusCode(
                res.statusText || "The server returned an error.",
                statusCode,
              ),
            );
          }
          return res.json();
        })
        .then((result) => {
          if (requestId !== activeSearchRequestIdRef.current) return;
          if (result && Object.hasOwn(result, "data")) {
            // Treat null data as empty array
            setSearchResults(result.data || []);
            const storeErrors = Array.isArray(result.errors)
              ? result.errors
              : [];
            setSearchStoreErrors(storeErrors);
            setDismissedStoreErrorsKey(null);
            updateUrlAndTitle(query, stores);
            if (window.gtag) {
              window.gtag("event", "view_search_results", {
                search_term: query,
              });
            }
          } else {
            throw new Error("Invalid response format from server");
          }
        })
        .catch((err) => {
          if (err.name === "AbortError") {
            if (
              userCancelledRef.current &&
              requestId === activeSearchRequestIdRef.current
            ) {
              const previousResults = resultsBeforeSearchRef.current;
              setSearchResults(previousResults);
              setHasSearched(previousResults.length > 0);
              setSearchError(null);
              setSearchStoreErrors([]);
              setDismissedStoreErrorsKey(null);
              userCancelledRef.current = false;
            }
            return;
          }
          if (requestId !== activeSearchRequestIdRef.current) return;
          console.error("Search error:", err);
          setSearchResults([]);
          setSearchStoreErrors([]);
          setDismissedStoreErrorsKey(null);

          // Set user-friendly error message
          if (
            err.message.includes("Failed to fetch") ||
            err.name === "TypeError"
          ) {
            setSearchError(
              "Unable to connect to the server. Please check your internet connection and try again.",
            );
          } else if (err.message.startsWith("Error (")) {
            setSearchError(err.message);
          } else if (err.message.includes("Server error")) {
            setSearchError(
              "The server is experiencing issues. Please try again later.",
            );
          } else {
            setSearchError(
              "An error occurred while searching. Please try again.",
            );
          }
        })
        .finally(() => {
          clearInterval(progressInterval);
          if (progressIntervalRef.current === progressInterval) {
            progressIntervalRef.current = null;
          }

          if (searchAbortControllerRef.current === searchAbortController) {
            searchAbortControllerRef.current = null;
          }

          if (requestId !== activeSearchRequestIdRef.current) return;
          setIsSearching(false);
          setSearchProgress("Search");
          skipSuggestionsRef.current = false;
        });
    },
    [updateUrlAndTitle],
  );

  const cancelSearch = useCallback(() => {
    if (!searchAbortControllerRef.current) return;

    userCancelledRef.current = true;
    searchAbortControllerRef.current.abort();

    if (progressIntervalRef.current) {
      clearInterval(progressIntervalRef.current);
      progressIntervalRef.current = null;
    }

    searchAbortControllerRef.current = null;
    setIsSearching(false);
    setSearchProgress("Search");
    skipSuggestionsRef.current = false;
  }, []);

  const retrySearch = useCallback(() => {
    const { query, stores } = lastSearchRef.current;
    if (query) {
      performSearch(query, stores);
    }
  }, [performSearch]);

  const storeErrorsKey = searchStoreErrors
    .map((entry) => `${entry.store}:${entry.error}`)
    .join("|");
  const visibleStoreErrors =
    searchStoreErrors.length > 0 && dismissedStoreErrorsKey !== storeErrorsKey
      ? searchStoreErrors
      : [];

  const dismissStoreErrors = useCallback(() => {
    setDismissedStoreErrorsKey(storeErrorsKey);
  }, [storeErrorsKey]);

  // --- Handlers ---
  const handleQueryChange = (e) => {
    setSearchQuery(e.target.value);
    if (searchError === SEARCH_TOO_SHORT_ERROR) {
      setSearchError(null);
    }
  };

  useEffect(() => {
    if (skipSuggestionsRef.current) {
      skipSuggestionsRef.current = false;
      return;
    }

    if (
      searchQuery.length > MIN_SEARCH_LENGTH - 1 &&
      searchQuery.length <= MAX_SEARCH_LENGTH
    ) {
      const timer = setTimeout(() => {
        if (skipSuggestionsRef.current) return;

        if (autocompleteAbortControllerRef.current) {
          autocompleteAbortControllerRef.current.abort();
        }
        const autocompleteAbortController = new AbortController();
        autocompleteAbortControllerRef.current = autocompleteAbortController;

        fetch(
          `https://api.scryfall.com/cards/autocomplete?q=${encodeURIComponent(searchQuery.toLowerCase())}`,
          { signal: autocompleteAbortController.signal },
        )
          .then((res) => {
            if (!res.ok) {
              throw new Error(`Autocomplete error: ${res.status}`);
            }
            return res.json();
          })
          .then((res) => {
            if (
              autocompleteAbortControllerRef.current !==
              autocompleteAbortController
            ) {
              return;
            }
            if (res.data && res.data.length > 0) {
              setSuggestions(res.data);
              setShowSuggestions(true);
            } else {
              setSuggestions([]);
              setShowSuggestions(false);
            }
          })
          .catch((err) => {
            if (err.name === "AbortError") return;
            console.error("Autocomplete error:", err);
            // Silently fail for autocomplete - not critical
            setSuggestions([]);
            setShowSuggestions(false);
          })
          .finally(() => {
            if (
              autocompleteAbortControllerRef.current ===
              autocompleteAbortController
            ) {
              autocompleteAbortControllerRef.current = null;
            }
          });
      }, AUTOCOMPLETE_DEBOUNCE_MS);
      return () => clearTimeout(timer);
    }

    if (autocompleteAbortControllerRef.current) {
      autocompleteAbortControllerRef.current.abort();
      autocompleteAbortControllerRef.current = null;
    }
    setSuggestions([]);
    setShowSuggestions(false);
  }, [searchQuery]);

  useEffect(() => {
    return () => {
      if (progressIntervalRef.current) {
        clearInterval(progressIntervalRef.current);
        progressIntervalRef.current = null;
      }
      if (searchAbortControllerRef.current) {
        searchAbortControllerRef.current.abort();
        searchAbortControllerRef.current = null;
      }
      if (autocompleteAbortControllerRef.current) {
        autocompleteAbortControllerRef.current.abort();
        autocompleteAbortControllerRef.current = null;
      }
    };
  }, []);

  const handleSuggestionClick = (suggestion) => {
    skipSuggestionsRef.current = true;
    setSearchQuery(suggestion);
    setShowSuggestions(false);

    performSearch(suggestion, resolveStoresToSearch(selectedStores));
  };

  const handleSearchSubmit = (e) => {
    if (e) e.preventDefault();
    setShowSuggestions(false);
    performSearch(searchQuery, resolveStoresToSearch(selectedStores));
  };

  const toggleStore = (store) => {
    setStoresWarning(null);
    const newStores = selectedStores.includes(store)
      ? selectedStores.filter((s) => s !== store)
      : [...selectedStores, store];
    setSelectedStores(newStores);
    persistSelectedStores(newStores);
  };

  const selectAllStores = () => {
    setStoresWarning(null);
    setSelectedStores(LGS_OPTIONS);
    persistSelectedStores(LGS_OPTIONS);
  };

  const selectNoStores = () => {
    setStoresWarning(null);
    setSelectedStores([]);
    persistSelectedStores([]);
  };

  // --- Initialization ---
  // Note: performSearch is included in deps but is stable (empty dep array in useCallback)
  // This effect should only run once on mount, not when selectedStores changes
  const hasInitializedRef = useRef(false);

  useEffect(() => {
    if (hasInitializedRef.current) return;
    hasInitializedRef.current = true;

    const urlParams = new URLSearchParams(window.location.search);
    if (urlParams.has("s") && urlParams.get("s") !== "") {
      const q = decodeURIComponent(urlParams.get("s"));
      skipSuggestionsRef.current = true;

      const urlStores = getStoresFromUrl(urlParams);
      const stores = urlStores ?? selectedStores;

      setTimeout(() => performSearch(q, stores), 100);
    }
  }, [performSearch, selectedStores]);

  return {
    searchQuery,
    setSearchQuery,
    isSearching,
    hasSearched,
    searchResults,
    searchProgress,
    searchError,
    searchStoreErrors: visibleStoreErrors,
    onDismissStoreErrors: dismissStoreErrors,
    storesWarning,
    suggestions,
    showSuggestions,
    setShowSuggestions,
    selectedStores,
    setSelectedStores,
    handleQueryChange,
    handleSuggestionClick,
    handleSearchSubmit,
    toggleStore,
    selectAllStores,
    selectNoStores,
    performSearch,
    cancelSearch,
    retrySearch,
  };
}
