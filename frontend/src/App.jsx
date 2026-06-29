import { lazy, Suspense, useCallback, useEffect, useState } from "react";
import "./index.css";

import Footer from "./components/Footer";
// --- Modular Components ---
import Header from "./components/Header";
import SearchForm from "./components/SearchForm";
import SearchResults from "./components/SearchResults";
import TopSearchKeywords from "./components/TopSearchKeywords";

const CartOffcanvas = lazy(() => import("./components/CartOffcanvas"));
const Modals = lazy(() => import("./components/Modals"));

// --- Constants ---
import { BASE_URL, LGS_MAP, LGS_OPTIONS, MIN_SEARCH_LENGTH } from "./constants";
// --- Hooks ---
import useCart from "./hooks/useCart";
import useSearch from "./hooks/useSearch";
import useTopSearchKeywords from "./hooks/useTopSearchKeywords";

const THEME_STORAGE_KEY = "gishathfetch-theme";

export default function App() {
  const {
    cart,
    showCart,
    setShowCart,
    addToCart,
    removeFromCart,
    removeFromCartByCard,
    clearCart,
    isCardInCart,
  } = useCart();

  const {
    searchQuery,
    setSearchQuery,
    isSearching,
    hasSearched,
    searchResults,
    searchProgress,
    searchError,
    searchStoreErrors,
    onDismissStoreErrors,
    storesWarning,
    cardKingdomPrice,
    suggestions,
    showSuggestions,
    setShowSuggestions,
    isLoadingSuggestions,
    showEmptySuggestions,
    selectedStores,
    setSelectedStores,
    handleQueryChange,
    handleClearQuery,
    handleSuggestionClick,
    handleSearchSubmit,
    toggleStore,
    selectAllStores,
    selectNoStores,
    performSearch,
    cancelSearch,
    retrySearch,
  } = useSearch();

  const {
    keywordsByPeriod: topSearchKeywordsByPeriod,
    isLoading: isLoadingTopSearchKeywords,
  } = useTopSearchKeywords(true);

  const [modalType, setModalType] = useState(null);
  const [theme, setTheme] = useState(() => {
    if (typeof window === "undefined") {
      return "light";
    }

    try {
      const savedTheme = localStorage.getItem(THEME_STORAGE_KEY);
      if (savedTheme === "light" || savedTheme === "dark") {
        return savedTheme;
      }
    } catch {
      // Ignore storage access issues and keep light mode as the default.
    }

    return "light";
  });

  useEffect(() => {
    document.documentElement.setAttribute("data-bs-theme", theme);
    try {
      localStorage.setItem(THEME_STORAGE_KEY, theme);
    } catch {
      // Ignore storage access issues and keep in-memory theme state only.
    }
  }, [theme]);

  // Support ?faq=1 in the URL so the FAQ modal can be opened for screenshots or deep links.
  useEffect(() => {
    if (typeof window === "undefined") {
      return;
    }
    const sp = new URLSearchParams(window.location.search);
    if (sp.get("faq") === "1") {
      setModalType("FAQ");
    }
  }, []);

  const handleThemeToggle = useCallback(() => {
    setTheme((currentTheme) => (currentTheme === "dark" ? "light" : "dark"));
  }, []);

  // --- Handlers ---
  const handleCardSearch = useCallback(
    (e, cardName, sourceStore) => {
      if (e?.preventDefault) e.preventDefault();
      setSearchQuery(cardName);
      setShowCart(false);
      setShowSuggestions(false); // Close suggestions dropdown

      // Update selected stores to show only the source store in checkboxes
      const storeArray = [sourceStore];
      setSelectedStores(storeArray);
      try {
        localStorage.setItem(
          "lgsSelected",
          encodeURIComponent(storeArray.join(",")),
        );
      } catch (err) {
        console.error("Failed to save selected stores:", err);
      }

      performSearch(cardName, storeArray);
    },
    [
      performSearch,
      setSearchQuery,
      setShowCart,
      setShowSuggestions,
      setSelectedStores,
    ],
  );

  // --- Main Render ---
  return (
    <div id="top" className="container-xl my-3 px-3 pb-3">
      <Header theme={theme} onToggleTheme={handleThemeToggle} />

      <SearchForm
        searchQuery={searchQuery}
        onQueryChange={handleQueryChange}
        onClearQuery={handleClearQuery}
        onSearchSubmit={handleSearchSubmit}
        suggestions={suggestions}
        showSuggestions={showSuggestions}
        showEmptySuggestions={showEmptySuggestions}
        isLoadingSuggestions={isLoadingSuggestions}
        onSuggestionClick={handleSuggestionClick}
        onFocus={() =>
          searchQuery.length > MIN_SEARCH_LENGTH - 1 && setShowSuggestions(true)
        }
        isSearching={isSearching}
        searchProgress={searchProgress}
        lgsOptions={LGS_OPTIONS}
        selectedStores={selectedStores}
        onStoreToggle={toggleStore}
        onSelectAll={selectAllStores}
        onSelectNone={selectNoStores}
        onCloseSuggestions={() => setShowSuggestions(false)}
        searchError={searchError}
        storesWarning={storesWarning}
        onCancelSearch={cancelSearch}
        popularSearchesSlot={
          <TopSearchKeywords
            keywordsByPeriod={topSearchKeywordsByPeriod}
            isLoading={isLoadingTopSearchKeywords}
            collapsible
            collapseOnSearch={isSearching}
            defaultExpanded={
              searchQuery.trim() === "" &&
              topSearchKeywordsByPeriod.last24Hours.length > 0
            }
          />
        }
      />

      <SearchResults
        results={searchResults}
        searchQuery={searchQuery}
        isSearching={isSearching}
        hasSearched={hasSearched}
        searchError={searchError}
        searchStoreErrors={searchStoreErrors}
        onDismissStoreErrors={onDismissStoreErrors}
        onRetrySearch={retrySearch}
        isCardInCart={isCardInCart}
        addToCart={addToCart}
        removeFromCart={removeFromCart}
        removeFromCartByCard={removeFromCartByCard}
        onSearchStore={handleCardSearch}
        cardKingdomPrice={cardKingdomPrice}
        baseUrl={BASE_URL}
      />

      <Footer
        cartCount={cart.length}
        onShowCart={() => setShowCart(true)}
        onShowMap={() => setModalType("MAP")}
        onShowFaq={() => setModalType("FAQ")}
      />

      <Suspense fallback={null}>
        <CartOffcanvas
          show={showCart}
          onHide={() => setShowCart(false)}
          cart={cart}
          isCardInCart={isCardInCart}
          removeFromCart={removeFromCart}
          onSearchStore={handleCardSearch}
          onClearCart={clearCart}
          baseUrl={BASE_URL}
        />

        <Modals
          showMap={modalType === "MAP"}
          onHideMap={() => setModalType(null)}
          showFaq={modalType === "FAQ"}
          onHideFaq={() => setModalType(null)}
          showPrivacy={modalType === "PRIVACY"}
          onHidePrivacy={() => setModalType(null)}
          onShowPrivacy={() => setModalType("PRIVACY")}
          lgsMapData={LGS_MAP}
        />
      </Suspense>
    </div>
  );
}
