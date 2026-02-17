import { useState, useEffect, useRef, useCallback } from 'react';
import { API_BASE_URL, LGS_OPTIONS, BASE_URL } from '../constants';

// Search configuration constants
const MIN_SEARCH_LENGTH = 3;
const AUTOCOMPLETE_DEBOUNCE_MS = 300;
const SEARCH_PROGRESS_INTERVAL_MS = 1000;
const MAX_PROGRESS_DOTS = 15;


export default function useSearch() {
    const [searchQuery, setSearchQuery] = useState(() => {
        const urlParams = new URLSearchParams(window.location.search);
        if (urlParams.has('s') && urlParams.get('s') !== "") {
            return decodeURIComponent(urlParams.get('s'));
        }
        return "";
    });
    const [isSearching, setIsSearching] = useState(false);
    const [searchResults, setSearchResults] = useState([]);
    const [hasSearched, setHasSearched] = useState(false);
    const [searchProgress, setSearchProgress] = useState("Search");
    const [suggestions, setSuggestions] = useState([]);
    const [showSuggestions, setShowSuggestions] = useState(false);
    const [selectedStores, setSelectedStores] = useState(() => {
        // First check URL parameters for store override
        const urlParams = new URLSearchParams(window.location.search);
        if (urlParams.has('src') && LGS_OPTIONS.includes(decodeURIComponent(urlParams.get('src')))) {
            const stores = [decodeURIComponent(urlParams.get('src'))];
            // Save to localStorage for URL-based navigation
            try {
                localStorage.setItem("lgsSelected", encodeURIComponent(stores.join(",")));
            } catch (err) {
                console.error("Failed to save selected stores:", err);
            }
            return stores;
        }

        // Otherwise check localStorage
        const storedLgs = localStorage.getItem('lgsSelected');
        if (storedLgs !== null) {
            const decoded = decodeURIComponent(storedLgs);
            return decoded === "" ? [] : decoded.split(",");
        }
        return LGS_OPTIONS;
    });

    // --- Helpers ---
    const skipSuggestionsRef = useRef(false);

    const updateUrlAndTitle = (query) => {
        if (window.location.hostname !== "localhost") {
            const newUrl = `${BASE_URL}?s=${encodeURIComponent(query.toLowerCase())}`;
            window.history.pushState(query.toLowerCase(), `${query.toLowerCase()} | Gishath Fetch`, newUrl);
            document.title = `${query.toLowerCase()} | Gishath Fetch`;
        }
    };

    const performSearch = useCallback((query, stores) => {
        if (!query || query.length < MIN_SEARCH_LENGTH) return;

        setIsSearching(true);
        setSearchProgress("Searching LGS");
        setSearchResults([]);
        setHasSearched(true);

        if (window.gtag) {
            window.gtag('event', 'search', { 'search_term': query.toLowerCase() });
        }

        const searchUrl = `${API_BASE_URL}?s=${encodeURIComponent(query.toLowerCase())}&lgs=${encodeURIComponent(stores.join(','))}`;

        let progressInterval = setInterval(() => {
            setSearchProgress(prev => {
                const dots = (prev.match(/\./g) || []).length;
                if (dots >= MAX_PROGRESS_DOTS) return "Searching LGS";
                return prev + " .";
            });
        }, SEARCH_PROGRESS_INTERVAL_MS);

        fetch(searchUrl)
            .then(res => res.json())
            .then(result => {
                if (result && result.data) {
                    setSearchResults(result.data);
                    updateUrlAndTitle(query);
                    if (window.gtag) {
                        window.gtag('event', 'view_search_results', { 'search_term': query.toLowerCase() });
                    }
                }
            })
            .catch(err => console.error("Search error:", err))
            .finally(() => {
                setIsSearching(false);
                setSearchProgress("Search");
                clearInterval(progressInterval);
                skipSuggestionsRef.current = false;
            });
    }, []);

    // --- Handlers ---
    const handleQueryChange = (e) => {
        setSearchQuery(e.target.value);
    };

    useEffect(() => {
        if (skipSuggestionsRef.current) {
            skipSuggestionsRef.current = false;
            return;
        }

        if (searchQuery.length > MIN_SEARCH_LENGTH - 1) {
            const timer = setTimeout(() => {
                fetch(`https://api.scryfall.com/cards/autocomplete?q=${encodeURIComponent(searchQuery.toLowerCase())}`)
                    .then(res => res.json())
                    .then(res => {
                        if (res.data && res.data.length > 0) {
                            setSuggestions(res.data);
                            setShowSuggestions(true);
                        } else {
                            setSuggestions([]);
                            setShowSuggestions(false);
                        }
                    })
                    .catch(err => console.error("Autocomplete error:", err));
            }, AUTOCOMPLETE_DEBOUNCE_MS);
            return () => clearTimeout(timer);
        }
        // When searchQuery is too short, suggestions will naturally be empty from previous state
    }, [searchQuery]);

    const handleSuggestionClick = (suggestion) => {
        skipSuggestionsRef.current = true;
        setSearchQuery(suggestion);
        setShowSuggestions(false);

        let storesToSearch = selectedStores;
        if (selectedStores.length === 0) {
            storesToSearch = LGS_OPTIONS;
            setSelectedStores(LGS_OPTIONS);
            try {
                localStorage.setItem("lgsSelected", encodeURIComponent(LGS_OPTIONS.join(",")));
            } catch (err) {
                console.error("Failed to save selected stores:", err);
            }
        }
        performSearch(suggestion, storesToSearch);
    };

    const handleSearchSubmit = (e) => {
        if (e) e.preventDefault();
        setShowSuggestions(false);

        let storesToSearch = selectedStores;
        if (selectedStores.length === 0) {
            storesToSearch = LGS_OPTIONS;
            setSelectedStores(LGS_OPTIONS);
            try {
                localStorage.setItem("lgsSelected", encodeURIComponent(LGS_OPTIONS.join(",")));
            } catch (err) {
                console.error("Failed to save selected stores:", err);
            }
        }

        performSearch(searchQuery, storesToSearch);
    };

    const toggleStore = (store) => {
        const newStores = selectedStores.includes(store)
            ? selectedStores.filter(s => s !== store)
            : [...selectedStores, store];
        setSelectedStores(newStores);
        try {
            localStorage.setItem("lgsSelected", encodeURIComponent(newStores.join(",")));
        } catch (err) {
            console.error("Failed to save selected stores:", err);
        }
    };

    const selectAllStores = () => {
        setSelectedStores(LGS_OPTIONS);
        try {
            localStorage.setItem("lgsSelected", encodeURIComponent(LGS_OPTIONS.join(",")));
        } catch (err) {
            console.error("Failed to save selected stores:", err);
        }
    };

    const selectNoStores = () => {
        setSelectedStores([]);
        try {
            localStorage.setItem("lgsSelected", encodeURIComponent(""));
        } catch (err) {
            console.error("Failed to save selected stores:", err);
        }
    };

    // --- Initialization ---
    // Note: performSearch is included in deps but is stable (empty dep array in useCallback)
    useEffect(() => {
        const urlParams = new URLSearchParams(window.location.search);
        if (urlParams.has('s') && urlParams.get('s') !== "") {
            const q = decodeURIComponent(urlParams.get('s'));
            skipSuggestionsRef.current = true;

            // Determine which stores to search (URL param takes precedence, already set in state initialization)
            const stores = urlParams.has('src') && LGS_OPTIONS.includes(decodeURIComponent(urlParams.get('src')))
                ? [decodeURIComponent(urlParams.get('src'))]
                : selectedStores;

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
        suggestions,
        showSuggestions,
        setShowSuggestions,
        selectedStores,
        handleQueryChange,
        handleSuggestionClick,
        handleSearchSubmit,
        toggleStore,
        selectAllStores,
        selectNoStores,
        performSearch
    };
}
