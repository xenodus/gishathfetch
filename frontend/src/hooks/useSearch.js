import { useState, useEffect, useRef } from 'react';
import { API_BASE_URL, LGS_OPTIONS, BASE_URL } from '../constants';

export default function useSearch() {
    const [searchQuery, setSearchQuery] = useState("");
    const [isSearching, setIsSearching] = useState(false);
    const [searchResults, setSearchResults] = useState([]);
    const [hasSearched, setHasSearched] = useState(false);
    const [searchProgress, setSearchProgress] = useState("Search");
    const [suggestions, setSuggestions] = useState([]);
    const [showSuggestions, setShowSuggestions] = useState(false);
    const [selectedStores, setSelectedStores] = useState(LGS_OPTIONS);

    const searchInputRef = useRef(null);

    // --- Helpers ---
    const updateUrlAndTitle = (query) => {
        if (window.location.hostname !== "localhost") {
            const newUrl = `${BASE_URL}?s=${encodeURIComponent(query.toLowerCase())}`;
            window.history.pushState(query.toLowerCase(), `${query.toLowerCase()} | Gishath Fetch`, newUrl);
            document.title = `${query.toLowerCase()} | Gishath Fetch`;
        }
    };

    const performSearch = (query, stores) => {
        if (!query || query.length < 3) return;

        setIsSearching(true);
        setSearchProgress("Searching LGS");
        setSearchResults([]);
        setHasSearched(true);

        if (window.gtag) {
            window.gtag('event', 'search', { 'search_term': query.toLowerCase() });
        }

        const searchUrl = `${API_BASE_URL}?s=${encodeURIComponent(query.toLowerCase())}&lgs=${encodeURIComponent(stores.join(','))}`;

        let progressInterval = setInterval(() => {
            setSearchProgress(prev => prev.length > 25 ? "Searching LGS" : prev + " .");
        }, 1000);

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
            });
    };

    // --- Handlers ---
    const handleQueryChange = (e) => {
        const val = e.target.value;
        setSearchQuery(val);
        if (val.length > 2) {
            // Scryfall autocomplete debounce
            const timer = setTimeout(() => {
                fetch(`https://api.scryfall.com/cards/autocomplete?q=${encodeURIComponent(val.toLowerCase())}`)
                    .then(res => res.json())
                    .then(res => {
                        if (res.data) {
                            setSuggestions(res.data);
                            setShowSuggestions(true);
                        }
                    })
                    .catch(err => console.error("Autocomplete error:", err));
            }, 300);
            return () => clearTimeout(timer);
        } else {
            setSuggestions([]);
            setShowSuggestions(false);
        }
    };

    const handleSuggestionClick = (suggestion) => {
        setSearchQuery(suggestion);
        setShowSuggestions(false);
        performSearch(suggestion, selectedStores);
    };

    const handleSearchSubmit = (e) => {
        if (e) e.preventDefault();
        setShowSuggestions(false);
        performSearch(searchQuery, selectedStores);
    };

    const toggleStore = (store) => {
        const newStores = selectedStores.includes(store)
            ? selectedStores.filter(s => s !== store)
            : [...selectedStores, store];
        setSelectedStores(newStores);
        localStorage.setItem("lgsSelected", encodeURIComponent(newStores.join(",")));
    };

    const selectAllStores = () => {
        setSelectedStores(LGS_OPTIONS);
        localStorage.setItem("lgsSelected", encodeURIComponent(LGS_OPTIONS.join(",")));
    };

    const selectNoStores = () => {
        setSelectedStores([]);
        localStorage.setItem("lgsSelected", encodeURIComponent(""));
    };

    // --- Initialization ---
    useEffect(() => {
        const storedLgs = localStorage.getItem('lgsSelected');
        if (storedLgs) {
            setSelectedStores(decodeURIComponent(storedLgs).split(","));
        }

        const urlParams = new URLSearchParams(window.location.search);
        if (urlParams.has('s') && urlParams.get('s') !== "") {
            const q = decodeURIComponent(urlParams.get('s'));
            setSearchQuery(q);

            let stores = LGS_OPTIONS;
            if (urlParams.has('src') && LGS_OPTIONS.includes(decodeURIComponent(urlParams.get('src')))) {
                stores = [decodeURIComponent(urlParams.get('src'))];
                setSelectedStores(stores);
                localStorage.setItem("lgsSelected", encodeURIComponent(stores.join(",")));
            }
            setTimeout(() => performSearch(q, stores), 100);
        }
    }, []);

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
