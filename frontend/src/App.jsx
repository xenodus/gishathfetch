import React, { useState, useEffect, useRef } from 'react';
import './index.css';

// --- Modular Components ---
import Header from './components/Header';
import SearchForm from './components/SearchForm';
import SearchResults from './components/SearchResults';
import CartOffcanvas from './components/CartOffcanvas';
import Modals from './components/Modals';
import Footer from './components/Footer';

// --- Constants ---
import {
  PAGE_TITLE,
  LGS_OPTIONS,
  LGS_MAP,
  API_BASE_URL,
  BASE_URL
} from './constants';

export default function App() {
  // --- State ---
  const [searchQuery, setSearchQuery] = useState("");
  const [suggestions, setSuggestions] = useState([]);
  const [selectedStores, setSelectedStores] = useState([]);
  const [searchResults, setSearchResults] = useState([]);
  const [isSearching, setIsSearching] = useState(false);
  const [searchProgress, setSearchProgress] = useState("");
  const [cart, setCart] = useState([]);

  // UI State
  const [showCart, setShowCart] = useState(false);
  const [showMap, setShowMap] = useState(false);
  const [showFaq, setShowFaq] = useState(false);
  const [showPrivacy, setShowPrivacy] = useState(false);
  const [showSuggestions, setShowSuggestions] = useState(false);

  const debounceTimer = useRef(null);

  // --- Initialization ---
  useEffect(() => {
    // 1. Initial Load from LocalStorage
    const storedLgs = localStorage.getItem('lgsSelected');
    if (storedLgs) {
      setSelectedStores(decodeURIComponent(storedLgs).split(","));
    } else {
      setSelectedStores(LGS_OPTIONS);
    }

    const storedCart = localStorage.getItem('cart');
    if (storedCart) {
      setCart(JSON.parse(storedCart));
    }

    // 2. Initial Search from URL
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

      // Delay search slightly to ensure state is set
      setTimeout(() => performSearch(q, stores), 100);
    }
  }, []);

  // --- Handlers ---
  const performSearch = (query, stores) => {
    if (!query || query.length < 3) return;

    setIsSearching(true);
    setSearchProgress("Searching LGS");
    setSearchResults([]);

    // Analytics
    if (window.gtag) {
      window.gtag('event', 'search', { 'search_term': query.toLowerCase() });
    }

    const searchUrl = `${API_BASE_URL}?s=${encodeURIComponent(query.toLowerCase())}&lgs=${encodeURIComponent(stores.join(','))}`;

    // Progress Animation Simulation
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
        clearInterval(progressInterval);
      });
  };

  const updateUrlAndTitle = (query) => {
    if (window.location.hostname !== "localhost") {
      const newUrl = `${BASE_URL}?s=${encodeURIComponent(query.toLowerCase())}`;
      window.history.pushState(query.toLowerCase(), `${query.toLowerCase()} | ${PAGE_TITLE}`, newUrl);
      document.title = `${query.toLowerCase()} | ${PAGE_TITLE}`;
    }
  };

  const handleSearchSubmit = (e) => {
    e.preventDefault();
    performSearch(searchQuery, selectedStores);
  };

  const handleStoreToggle = (store) => {
    let newStores;
    if (selectedStores.includes(store)) {
      newStores = selectedStores.filter(s => s !== store);
    } else {
      newStores = [...selectedStores, store];
    }
    setSelectedStores(newStores);
    localStorage.setItem("lgsSelected", encodeURIComponent(newStores.join(",")));
  };

  const handleSelectAll = () => {
    setSelectedStores(LGS_OPTIONS);
    localStorage.setItem("lgsSelected", encodeURIComponent(LGS_OPTIONS.join(",")));
  };

  const handleSelectNone = () => {
    setSelectedStores([]);
    localStorage.setItem("lgsSelected", "");
  };

  const handleQueryChange = (e) => {
    const val = e.target.value;
    setSearchQuery(val);

    if (val.trim().length > 2) {
      clearTimeout(debounceTimer.current);
      debounceTimer.current = setTimeout(() => {
        fetch(`https://api.scryfall.com/cards/autocomplete?q=${encodeURIComponent(val.toLowerCase())}`)
          .then(res => res.json())
          .then(data => {
            if (data && data.data) {
              setSuggestions(data.data);
              setShowSuggestions(true);
            }
          });
      }, 300);
    } else {
      setSuggestions([]);
      setShowSuggestions(false);
    }
  };

  const handleSuggestionClick = (suggestion) => {
    setSearchQuery(suggestion);
    setShowSuggestions(false);
  };

  const addToCart = (card) => {
    const newCart = [card, ...cart];
    setCart(newCart);
    localStorage.setItem("cart", JSON.stringify(newCart));
  };

  const removeFromCart = (index) => {
    const newCart = [...cart];
    newCart.splice(index, 1);
    setCart(newCart);
    localStorage.setItem("cart", JSON.stringify(newCart));
  };

  const clearCart = () => {
    if (window.confirm("Are you sure you want to remove all saved cards?")) {
      setCart([]);
      localStorage.removeItem("cart");
    }
  };

  const isCardInCart = (card) => {
    return cart.some(item => JSON.stringify(item) === JSON.stringify(card));
  };

  const handleSearchStore = (e, name, src) => {
    e.preventDefault();
    setShowCart(false);
    setSearchQuery(name);
    performSearch(name, [src]);
  };

  // --- Main Render ---
  return (
    <div id="top" className="container-xl my-3 px-3 pb-3">
      <Header />

      <SearchForm
        searchQuery={searchQuery}
        onQueryChange={handleQueryChange}
        onSearchSubmit={handleSearchSubmit}
        suggestions={suggestions}
        showSuggestions={showSuggestions}
        onSuggestionClick={handleSuggestionClick}
        onFocus={() => searchQuery.length > 2 && setShowSuggestions(true)}
        isSearching={isSearching}
        searchProgress={searchProgress}
        lgsOptions={LGS_OPTIONS}
        selectedStores={selectedStores}
        onStoreToggle={handleStoreToggle}
        onSelectAll={handleSelectAll}
        onSelectNone={handleSelectNone}
      />

      <SearchResults
        results={searchResults}
        isCardInCart={isCardInCart}
        addToCart={addToCart}
        removeFromCart={removeFromCart}
        onSearchStore={handleSearchStore}
        baseUrl={BASE_URL}
      />

      <Footer
        cartCount={cart.length}
        onShowCart={() => setShowCart(true)}
        onShowMap={() => setShowMap(true)}
        onShowFaq={() => setShowFaq(true)}
      />

      <CartOffcanvas
        show={showCart}
        onHide={() => setShowCart(false)}
        cart={cart}
        isCardInCart={isCardInCart}
        removeFromCart={removeFromCart}
        onSearchStore={handleSearchStore}
        onClearCart={clearCart}
        baseUrl={BASE_URL}
      />

      <Modals
        showMap={showMap}
        onHideMap={() => setShowMap(false)}
        showFaq={showFaq}
        onHideFaq={() => setShowFaq(false)}
        showPrivacy={showPrivacy}
        onHidePrivacy={() => setShowPrivacy(false)}
        onShowPrivacy={() => { setShowMap(false); setShowFaq(false); setShowPrivacy(true); }}
        lgsMapData={LGS_MAP}
      />
    </div>
  );
}
